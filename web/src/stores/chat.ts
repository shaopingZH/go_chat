import { defineStore } from "pinia"
import {
  apiRequest,
  authHeaders,
  buildWSURLs,
  uploadImage,
  getGroupMembers as apiGetGroupMembers,
  getPublicProfile as apiGetPublicProfile,
  leaveGroup,
  kickGroupMember,
  batchGetUserOnlineStatus,
  updateGroup,
  markConversationRead,
  type UserOnlineStatus,
} from "../api/http"
import { useAuthStore } from "./auth"
import { shouldApplySearchResponse } from "../utils/chatSearch"
import { createPendingImageMessage, findPendingImageReplacementIndex } from "../utils/chatImageUpload"
import type {
  ChatMessage,
  ChatType,
  ConversationItem,
  GroupItem,
  MessageRecord,
  PrivateConversationSummary,
  WSIncomingEnvelope,
  GroupMember,
  UserProfile,
  UserOnlineStatusPayload,
} from "../types/chat"

type WSStatus = "disconnected" | "connecting" | "connected"

interface MessageListResponse {
  items: MessageRecord[]
  has_more: boolean
}

interface GroupListResponse {
  items: GroupItem[]
}

interface PrivateConversationResponse {
  items: PrivateConversationSummary[]
}

function makeKey(chatType: ChatType, targetId: number): string {
  return `${chatType}:${targetId}`
}

function parseKey(key: string): { chatType: ChatType; targetId: number } {
  const [chatType, target] = key.split(":")
  return {
    chatType: (chatType === "group" ? "group" : "private") as ChatType,
    targetId: Number(target),
  }
}

function sortConversations(items: ConversationItem[]): ConversationItem[] {
  return [...items].sort((a, b) => {
    const t1 = a.lastAt ? new Date(a.lastAt).getTime() : 0
    const t2 = b.lastAt ? new Date(b.lastAt).getTime() : 0
    return t2 - t1
  })
}

function parseHost(url: string): string {
  try {
    return new URL(url).host
  } catch {
    return url
  }
}

function toChatMessage(record: MessageRecord, chatType: ChatType): ChatMessage {
  const senderId = Number(record.sender?.id || record.sender_id)
  const username = record.sender?.display_name || record.sender?.username || `用户 ${senderId}`
  const avatar = record.sender?.avatar || ""

  return {
    id: record.id,
    sender: {
      id: senderId,
      username: username,
      avatar: avatar,
    },
    target_id: chatType === "group" ? Number(record.group_id || 0) : Number(record.receiver_id || 0),
    chat_type: chatType,
    msg_type: normalizeMsgType(record.msg_type),
    content: record.content,
    created_at: record.created_at,
  }
}

function parseMillis(raw: string): number {
  if (!raw) {
    return 0
  }
  const ts = new Date(raw).getTime()
  return Number.isNaN(ts) ? 0 : ts
}

function normalizeMsgType(msgType: number | null | undefined): number {
  return Number(msgType) === 2 ? 2 : 1
}

function buildConversationPreview(content: string, msgType: number): string {
  if (normalizeMsgType(msgType) === 2) {
    return "[图片]"
  }
  return content
}

import { useFriendStore } from "./friend"

const CONVERSATIONS_CACHE_KEY = "go_chat_conversations"
const ACTIVE_KEY_CACHE_KEY = "go_chat_active_key"
const HIDDEN_CONVS_CACHE_KEY = "go_chat_hidden_convs"

function loadCachedConversations(): ConversationItem[] {
  try {
    const raw = localStorage.getItem(CONVERSATIONS_CACHE_KEY)
    return raw ? (JSON.parse(raw) as ConversationItem[]) : []
  } catch {
    return []
  }
}

function loadCachedHiddenConvs(): string[] {
  try {
    const raw = localStorage.getItem(HIDDEN_CONVS_CACHE_KEY)
    return raw ? (JSON.parse(raw) as string[]) : []
  } catch {
    return []
  }
}

function loadCachedActiveKey(): string {
  return localStorage.getItem(ACTIVE_KEY_CACHE_KEY) || ""
}

export interface AppToast {
  id: number
  title: string
  message: string
  type?: "info" | "success" | "warning" | "error"
}

let toastIdCounter = 0
let searchRequestCounter = 0

export const useChatStore = defineStore("chat", {
  state: () => ({
    ws: null as WebSocket | null,
    wsStatus: "disconnected" as WSStatus,
    conversations: loadCachedConversations(),
    hiddenConversations: loadCachedHiddenConvs(),
    messagesByKey: {} as Record<string, ChatMessage[]>,
    loadedHistory: {} as Record<string, boolean>,
    hasMoreByKey: {} as Record<string, boolean>,
    activeKey: "",
    errorMessage: "",
    isLoadingConversations: false,
    reconnectTimer: null as ReturnType<typeof setTimeout> | null,
    manualClose: false,
    groupMembersCache: {} as Record<number, GroupMember[]>,
    userOnlineStatusCache: {} as Record<number, UserOnlineStatus>,
    userProfileCache: {} as Record<number, UserProfile>,
    showRightSidebar: false,
    showLeftSidebar: false,
    leftSidebarType: "profile" as "profile" | "friends" | "addFriend" | "createGroup" | "joinGroup" | "groupOptions",
    rightSidebarType: "none" as "none" | "profile" | "groupProfile" | "publicProfile",
    rightSidebarPayload: null as any,

    // UI State
    toasts: [] as AppToast[],

    // Search State
    isSearching: false,
    searchKeyword: "",
    searchResults: [] as ChatMessage[],
    searchCurrentIndex: 0,
    isSearchingAPI: false,
  }),
  getters: {
    activeConversation(state): ConversationItem | null {
      return state.conversations.find((item) => item.key === state.activeKey) || null
    },
    activeMessages(state): ChatMessage[] {
      return state.messagesByKey[state.activeKey] || []
    },
    isConnected(state): boolean {
      return state.wsStatus === "connected"
    },
  },
  actions: {
    persistConversations() {
      localStorage.setItem(CONVERSATIONS_CACHE_KEY, JSON.stringify(this.conversations))
      if (this.activeKey) {
        localStorage.setItem(ACTIVE_KEY_CACHE_KEY, this.activeKey)
      } else {
        localStorage.removeItem(ACTIVE_KEY_CACHE_KEY)
      }
    },
    persistHiddenConversations() {
      localStorage.setItem(HIDDEN_CONVS_CACHE_KEY, JSON.stringify(this.hiddenConversations))
    },
    unhideConversation(key: string) {
      if (this.hiddenConversations.includes(key)) {
        this.hiddenConversations = this.hiddenConversations.filter(k => k !== key)
        this.persistHiddenConversations()
      }
    },
    resetState() {
      this.disconnect()
      this.conversations = []
      localStorage.removeItem(CONVERSATIONS_CACHE_KEY)
      localStorage.removeItem(ACTIVE_KEY_CACHE_KEY)
      this.messagesByKey = {}
      this.activeKey = ""
      this.errorMessage = ""
      this.isLoadingConversations = false
      this.groupMembersCache = {}
      this.userOnlineStatusCache = {}
      this.userProfileCache = {}
      this.showRightSidebar = false
      this.showLeftSidebar = false
      this.leftSidebarType = "profile"
      this.rightSidebarType = "none"
      this.rightSidebarPayload = null
    },
    upsertConversation({ chatType, targetId, title, avatar, ownerId }: { chatType: ChatType; targetId: number; title?: string; avatar?: string; ownerId?: number }): ConversationItem {
      const key = makeKey(chatType, targetId)
      const index = this.conversations.findIndex((item) => item.key === key)
      
      const placeholder = chatType === "group" ? `群组 ${targetId}` : `用户 ${targetId}`
      const newTitle = title || placeholder

      if (index >= 0) {
        const existing = this.conversations[index]
        // Only update if the new title is not a placeholder, OR if the existing title IS a placeholder
        const existingIsPlaceholder = existing.title === `用户 ${targetId}` || existing.title === `群组 ${targetId}`
        const newIsPlaceholder = !title || newTitle === placeholder

        if (!newIsPlaceholder || existingIsPlaceholder) {
          if (title) existing.title = title
        }
        
        if (avatar !== undefined && avatar !== "") {
          existing.avatar = avatar
        }

        if (ownerId !== undefined) {
          existing.owner_id = ownerId
        }

        this.conversations = sortConversations(this.conversations)
        this.persistConversations()
        return this.conversations[index]
      }

      const created: ConversationItem = {
        key,
        chatType,
        targetId,
        title: newTitle,
        avatar: avatar || "",
        unread: 0,
        lastText: "",
        lastAt: new Date().toISOString(),
        owner_id: ownerId,
      }
      this.conversations = sortConversations([created, ...this.conversations])
      this.persistConversations()

      // Fetch profile for private chats if title/avatar is missing or we only have a placeholder
      if (chatType === "private" && (!title || !avatar || newTitle === placeholder)) {
        void this.getUserProfile(targetId).then((profile) => {
          if (profile) {
            created.title = profile.display_name || profile.username
            created.avatar = profile.avatar
            this.conversations = sortConversations([...this.conversations])
            this.persistConversations()
          }
        })
      }

      return created
    },
    markRead(key: string) {
      console.log(`[ChatStore] Marking ${key} as read`);
      const conv = this.conversations.find((item) => item.key === key)
      if (conv) {
        conv.unread = 0
        this.persistConversations()

        // Report to server
        const auth = useAuthStore()
        if (auth.token) {
          void markConversationRead(auth.token, conv.chatType, conv.targetId).catch(console.error)
        }
      }
    },
    async activateConversation(key: string, silent = false) {
      // Clear search state when switching conversations
      if (this.activeKey !== key && this.isSearching) {
        this.closeSearch()
      }

      this.activeKey = key
      this.errorMessage = ""
      this.persistConversations()
      if (!silent) {
        this.markRead(key)
      }
      await this.ensureHistoryLoaded(key)
    },
    async openPrivateByUserID(userID: number) {
      const key = makeKey("private", userID)
      this.unhideConversation(key)
      const conv = this.upsertConversation({ chatType: "private", targetId: userID })

      // Try to get a fresh profile immediately if we don't have a real name yet
      void this.getUserProfile(userID).then((profile) => {
        if (profile) {
          conv.title = profile.display_name || profile.username
          conv.avatar = profile.avatar
          this.persistConversations()
        }
      })

      this.loadedHistory[key] = false
      await this.activateConversation(key)
    },

    async ensureHistoryLoaded(key: string, force = false) {
      if (!force && this.loadedHistory[key]) {
        return
      }

      const auth = useAuthStore()
      if (!auth.token) {
        return
      }

      const { chatType, targetId } = parseKey(key)
      const data = await apiRequest<MessageListResponse>(`/api/messages?target_id=${targetId}&type=${chatType}&limit=80`, {
        headers: authHeaders(auth.token),
      })

      const list = Array.isArray(data?.items) ? data.items.map((item) => toChatMessage(item, chatType)) : []

      this.messagesByKey[key] = list
      this.loadedHistory[key] = true
      this.hasMoreByKey[key] = data?.has_more ?? false

      if (list.length > 0) {
        const last = list[list.length - 1]
        this.touchConversation(key, last.content, last.created_at, last.msg_type)
      }
    },

    async loadMoreHistory(key: string): Promise<boolean> {
      const auth = useAuthStore()
      if (!auth.token) return false

      if (this.hasMoreByKey[key] === false) return false

      const currentMessages = this.messagesByKey[key] || []
      if (currentMessages.length === 0) return false

      const oldestMessage = currentMessages[0]
      const { chatType, targetId } = parseKey(key)
      
      try {
        const data = await apiRequest<MessageListResponse>(`/api/messages?target_id=${targetId}&type=${chatType}&last_msg_id=${oldestMessage.id}&limit=80`, {
          headers: authHeaders(auth.token),
        })
        
        if (data && Array.isArray(data.items) && data.items.length > 0) {
          const olderMessages = data.items.map((item) => toChatMessage(item, chatType))
          this.messagesByKey[key] = [...olderMessages, ...currentMessages]
          this.hasMoreByKey[key] = data.has_more
          return true
        }
        this.hasMoreByKey[key] = false
        return false
      } catch (error) {
        console.error("Failed to load more history:", error)
        return false
      }
    },

    async loadHistoryAround(key: string, messageId: number): Promise<boolean> {
      const auth = useAuthStore()
      if (!auth.token) return false

      const { chatType, targetId } = parseKey(key)
      
      try {
        const data = await apiRequest<MessageListResponse>(`/api/messages?target_id=${targetId}&type=${chatType}&around_msg_id=${messageId}&limit=40`, {
          headers: authHeaders(auth.token),
        })
        
        if (data && Array.isArray(data.items) && data.items.length > 0) {
          const messages = data.items.map((item) => toChatMessage(item, chatType))
          // When jumping to a specific context, we replace the currently loaded view
          // similar to how Telegram Desktop works
          this.messagesByKey[key] = messages
          this.hasMoreByKey[key] = data.has_more
          return true
        }
        return false
      } catch (error) {
        console.error("Failed to load history around message:", error)
        return false
      }
    },
    
    removeConversation(key: string) {
      if (!this.hiddenConversations.includes(key)) {
        this.hiddenConversations.push(key)
        this.persistHiddenConversations()
      }
      this.conversations = this.conversations.filter(c => c.key !== key)
      if (this.activeKey === key) {
        this.activeKey = ""
      }
      this.persistConversations()
    },

    // UI Toast Actions
    addToast(toast: Omit<AppToast, "id">, duration = 4000) {
      const id = ++toastIdCounter
      this.toasts.push({ ...toast, id })
      if (duration > 0) {
        setTimeout(() => {
          this.removeToast(id)
        }, duration)
      }
    },
    removeToast(id: number) {
      this.toasts = this.toasts.filter(t => t.id !== id)
    },

    // Search Actions
    openSearch() {
      this.isSearching = true
      this.searchKeyword = ""
      this.searchResults = []
      this.searchCurrentIndex = 0
    },
    closeSearch() {
      this.isSearching = false
      this.searchKeyword = ""
      this.searchResults = []
      this.searchCurrentIndex = 0
      searchRequestCounter++
      this.isSearchingAPI = false
    },
    async executeSearch() {
      if (!this.activeKey || !this.searchKeyword.trim()) {
        this.searchResults = []
        this.searchCurrentIndex = 0
        searchRequestCounter++
        this.isSearchingAPI = false
        return
      }

      const auth = useAuthStore()
      if (!auth.token) return

      const requestKeyword = this.searchKeyword.trim()
      const requestKey = this.activeKey
      const requestId = ++searchRequestCounter
      this.isSearchingAPI = true
      try {
        const { chatType, targetId } = parseKey(requestKey)
        const keyword = encodeURIComponent(requestKeyword)
        // Fetch up to 100 results for the search
        const data = await apiRequest<MessageListResponse>(`/api/messages?target_id=${targetId}&type=${chatType}&keyword=${keyword}&limit=100`, {
          headers: authHeaders(auth.token),
        })

        if (!shouldApplySearchResponse({
          requestId,
          activeRequestId: searchRequestCounter,
          requestKeyword,
          currentKeyword: this.searchKeyword,
          requestKey,
          currentKey: this.activeKey,
        })) {
          return
        }

        this.searchResults = Array.isArray(data?.items) ? data.items.map((item) => toChatMessage(item, chatType)) : []
        
        // Sort results so the newest message is at index 0
        this.searchResults.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        
        this.searchCurrentIndex = this.searchResults.length > 0 ? 0 : -1
      } catch (error) {
        if (!shouldApplySearchResponse({
          requestId,
          activeRequestId: searchRequestCounter,
          requestKeyword,
          currentKeyword: this.searchKeyword,
          requestKey,
          currentKey: this.activeKey,
        })) {
          return
        }
        console.error("Search failed:", error)
        this.searchResults = []
      } finally {
        if (requestId === searchRequestCounter) {
          this.isSearchingAPI = false
        }
      }
    },
    nextSearchResult() {
      if (this.searchResults.length === 0) return
      if (this.searchCurrentIndex < this.searchResults.length - 1) {
        this.searchCurrentIndex++
      }
    },
    prevSearchResult() {
      if (this.searchResults.length === 0) return
      if (this.searchCurrentIndex > 0) {
        this.searchCurrentIndex--
      }
    },

    touchConversation(key: string, content: string, createdAt: string, msgType = 1) {
      const conv = this.conversations.find((item) => item.key === key)
      if (!conv) {
        return
      }
      conv.lastText = buildConversationPreview(content, msgType)
      conv.lastAt = createdAt
      this.conversations = sortConversations(this.conversations)
      this.persistConversations()
    },
    markConversationNeedRefresh(key: string, latestAt: string) {
      const existing = this.messagesByKey[key]
      if (!existing || existing.length === 0) {
        this.loadedHistory[key] = false
        return
      }

      const latestLocal = parseMillis(existing[existing.length - 1].created_at)
      const latestRemote = parseMillis(latestAt)
      if (latestRemote > latestLocal) {
        this.loadedHistory[key] = false
      }
    },
    ingestMessage(payload: ChatMessage) {
      const auth = useAuthStore()
      const selfID = Number(auth.user?.id || 0)

      const normalizedPayload: ChatMessage = {
        ...payload,
        msg_type: normalizeMsgType(payload.msg_type),
      }

      let key = ""
      let title = ""
      let avatar = ""
      let targetId = Number(normalizedPayload.target_id)

      if (normalizedPayload.chat_type === "group") {
        key = makeKey("group", targetId)
        title = this.conversations.find((item) => item.key === key)?.title || `群组 ${targetId}`
      } else {
        const senderID = Number(normalizedPayload.sender?.id || 0)
        targetId = senderID === selfID ? Number(normalizedPayload.target_id) : senderID
        key = makeKey("private", targetId)
        
        // Strategy: 1. Cache, 2. Existing Conv Title, 3. Sender Info, 4. Placeholder
        const cached = this.userProfileCache[targetId]
        const existing = this.conversations.find(c => c.key === key)
        
        if (cached) {
          title = cached.display_name || cached.username
          avatar = cached.avatar
        } else if (existing && existing.title !== `用户 ${targetId}`) {
          title = existing.title
          avatar = existing.avatar || ""
        } else if (senderID !== selfID && normalizedPayload.sender) {
          title = normalizedPayload.sender.display_name || normalizedPayload.sender.username
          avatar = normalizedPayload.sender.avatar
        } else {
          title = `用户 ${targetId}`
        }
      }

      this.unhideConversation(key)
      this.upsertConversation({ chatType: normalizedPayload.chat_type, targetId, title, avatar: avatar || undefined })

      const bucket = this.messagesByKey[key] || []
      const exists = bucket.some((item) => item.id === normalizedPayload.id)
      const pendingIndex = findPendingImageReplacementIndex(bucket, {
        senderId: Number(normalizedPayload.sender?.id || 0),
        chatType: normalizedPayload.chat_type,
        targetId,
      })

      if (!exists && pendingIndex >= 0 && Number(normalizedPayload.msg_type) === 2 && Number(normalizedPayload.sender?.id || 0) === selfID) {
        const pendingMessage = bucket[pendingIndex]
        if (pendingMessage?.content?.startsWith("blob:")) {
          URL.revokeObjectURL(pendingMessage.content)
        }
        bucket[pendingIndex] = normalizedPayload
        this.messagesByKey[key] = bucket
        this.touchConversation(key, normalizedPayload.content, normalizedPayload.created_at, normalizedPayload.msg_type)
        return
      }

      if (!exists) {
        bucket.push(normalizedPayload)
      }
      this.messagesByKey[key] = bucket
      this.touchConversation(key, normalizedPayload.content, normalizedPayload.created_at, normalizedPayload.msg_type)

      const isSelf = Number(normalizedPayload.sender?.id || 0) === selfID
      if (this.activeKey === key && !isSelf) {
        // We are currently looking at this chat, so mark it as read on the server immediately
        void this.markRead(key)
      } else if (this.activeKey !== key && !isSelf) {
        const conv = this.conversations.find((item) => item.key === key)
        if (conv) {
          conv.unread += 1
          this.persistConversations()
        }
      }
    },
    async listGroups(activateDefault = true) {
      const auth = useAuthStore()
      if (!auth.token) {
        return
      }
      const data = await apiRequest<GroupListResponse>("/api/groups", {
        headers: authHeaders(auth.token),
      })
      const groups = data?.items || []
      groups.forEach((group) => {
        const key = makeKey("group", Number(group.id))
        if (this.hiddenConversations.includes(key)) return

        const conversation = this.upsertConversation({
          chatType: "group",
          targetId: Number(group.id),
          title: group.name,
          ownerId: group.owner_id,
        })

        // Use server unread count
        if (group.unread_count !== undefined) {
          conversation.unread = group.unread_count
        }

        const lastMessage = group.last_message
        if (lastMessage?.content && lastMessage?.created_at) {
          this.touchConversation(conversation.key, lastMessage.content, lastMessage.created_at, normalizeMsgType(lastMessage.msg_type))
        }
      })
      if (activateDefault && !this.activeKey && this.conversations.length > 0) {
        await this.activateConversation(this.conversations[0].key)
      }
    },
    async listPrivateConversations(activateDefault = true) {
      const auth = useAuthStore()
      if (!auth.token) {
        return
      }

      const data = await apiRequest<PrivateConversationResponse>("/api/conversations/private?limit=50", {
        headers: authHeaders(auth.token),
      })

      const items = data?.items || []
      items.forEach((item) => {
        const targetId = Number(item.partner_id)
        if (!targetId) {
          return
        }

        const key = makeKey("private", targetId)
        if (this.hiddenConversations.includes(key)) return

        const conversation = this.upsertConversation({
          chatType: "private",
          targetId,
          title: item.partner_username || `用户 ${targetId}`,
          avatar: item.partner_avatar,
        })

        // Use server unread count
        conversation.unread = item.unread_count || 0

        this.touchConversation(key, item.last_content || "", item.last_created_at || "", normalizeMsgType(item.last_msg_type))
        this.markConversationNeedRefresh(key, item.last_created_at || "")
      })

      if (activateDefault && !this.activeKey && this.conversations.length > 0) {
        await this.activateConversation(this.conversations[0].key)
      }
    },
    async bootstrapConversations() {
      this.isLoadingConversations = true
      try {
        const cachedKey = loadCachedActiveKey()

        // Load both lists in parallel but handle errors individually
        await Promise.all([
          this.listGroups(false).catch((err) => console.error("Failed to list groups:", err)),
          this.listPrivateConversations(false).catch((err) => console.error("Failed to list private convs:", err)),
        ])

        // Try to restore previous active key if still exists in current list
        // Use silent=true to avoid marking as read automatically
        if (cachedKey && this.conversations.some(c => c.key === cachedKey)) {
          await this.activateConversation(cachedKey, true)
        }
      } finally {
        this.isLoadingConversations = false
      }
    },
    async createGroup(name: string) {
      const auth = useAuthStore()
      const data = await apiRequest<GroupItem>("/api/groups", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeaders(auth.token),
        },
        body: JSON.stringify({ name }),
      })
      const key = makeKey("group", Number(data.id))
      this.unhideConversation(key)
      this.upsertConversation({ chatType: "group", targetId: Number(data.id), title: data.name })
      await this.activateConversation(key)
    },
    async joinGroup(groupID: number) {
      const auth = useAuthStore()
      await apiRequest(`/api/groups/${groupID}/join`, {
        method: "POST",
        headers: authHeaders(auth.token),
      })
      if (this.errorMessage === "not group member") {
        this.errorMessage = ""
      }
      await this.listGroups()
    },

    async updateGroupInfo(groupId: number, payload: { name?: string; avatar?: string }) {
      const auth = useAuthStore()
      if (!auth.token) return
      await updateGroup(auth.token, groupId, payload)
      // Update local store immediately to reflect changes
      const key = makeKey("group", groupId)
      const conv = this.conversations.find(c => c.key === key)
      if (conv) {
        if (payload.name !== undefined) conv.title = payload.name
        if (payload.avatar !== undefined) conv.avatar = payload.avatar
        this.persistConversations()
      }
    },

    async refreshGroupMembers(groupId: number): Promise<GroupMember[]> {
      const auth = useAuthStore()
      if (!auth.token) return this.groupMembersCache[groupId] || []

      const members = await apiGetGroupMembers(auth.token, groupId)
      this.groupMembersCache[groupId] = members
      return members
    },
    async getGroupMembers(groupId: number): Promise<GroupMember[]> {
      if (!groupId) {
        return []
      }

      const hasCached = Object.prototype.hasOwnProperty.call(this.groupMembersCache, groupId)
      if (hasCached) {
        void this.refreshGroupMembers(groupId).catch(() => {})
        return this.groupMembersCache[groupId]
      }

      return this.refreshGroupMembers(groupId)
    },
    async fetchGroupMembersPresence(groupId: number) {
      const auth = useAuthStore()
      if (!auth.token) return

      const members = this.groupMembersCache[groupId]
      if (!members || members.length === 0) return

      const userIds = members.map(m => Number(m.user_id || m.id))
      try {
        const statuses = await batchGetUserOnlineStatus(auth.token, userIds)
        // Update cache with a new object to ensure reactivity
        const newCache = { ...this.userOnlineStatusCache }
        statuses.forEach(s => {
          newCache[s.user_id] = s
        })
        this.userOnlineStatusCache = newCache
      } catch {
        // ignore presence fetch error
      }
    },
    async fetchUserPresence(userId: number) {
      const auth = useAuthStore()
      if (!auth || !auth.token) return

      try {
        const statuses = await batchGetUserOnlineStatus(auth.token, [userId])
        if (statuses.length > 0) {
          this.userOnlineStatusCache = {
            ...this.userOnlineStatusCache,
            [userId]: statuses[0]
          }
        }
      } catch {
        // ignore
      }
    },
    async refreshUserProfile(userId: number): Promise<UserProfile | null> {
      const auth = useAuthStore()
      if (!auth.token) return this.userProfileCache[userId] || null

      const profile = await apiGetPublicProfile(auth.token, userId)
      this.userProfileCache[userId] = profile
      return profile
    },
    async getUserProfile(userId: number): Promise<UserProfile | null> {
      if (!userId) {
        return null
      }

      const hasCached = Object.prototype.hasOwnProperty.call(this.userProfileCache, userId)
      if (hasCached) {
        void this.refreshUserProfile(userId).catch(() => {})
        return this.userProfileCache[userId]
      }

      return this.refreshUserProfile(userId)
    },
    async leaveCurrentGroup() {
      const conv = this.activeConversation
      if (!conv || conv.chatType !== "group") return

      const auth = useAuthStore()
      if (!auth.token) return

      await leaveGroup(auth.token, conv.targetId)
      
      // Clean up local state
      this.conversations = this.conversations.filter(c => c.key !== conv.key)
      delete this.groupMembersCache[conv.targetId]
      if (this.activeKey === conv.key) {
        this.activeKey = ""
      }
    },
    async kickMember(groupId: number, userId: number) {
      const auth = useAuthStore()
      if (!auth.token) return
      await kickGroupMember(auth.token, groupId, userId)

      const members = this.groupMembersCache[groupId]
      if (members) {
        this.groupMembersCache[groupId] = members.filter((member) => (member.user_id || member.id) !== userId)
      }
    },
    openRightSidebar(type: "profile" | "groupProfile" | "publicProfile", payload?: any) {
      this.rightSidebarType = type
      this.rightSidebarPayload = payload
      this.showRightSidebar = true
    },
    closeRightSidebar() {
      this.showRightSidebar = false
      // setTimeout(() => { this.rightSidebarType = "none" }, 300) // wait for animation
    },
    openLeftSidebar(type: "profile" | "friends" | "addFriend" | "createGroup" | "joinGroup" | "groupOptions" = "profile") {
      this.leftSidebarType = type
      this.showLeftSidebar = true
    },
    closeLeftSidebar() {
      this.showLeftSidebar = false
    },
    connect() {
      const auth = useAuthStore()
      if (!auth.token) {
        return
      }

      this.manualClose = false
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer)
        this.reconnectTimer = null
      }

      const wsURLs = buildWSURLs(auth.token)
      this.tryConnectWithFallback(wsURLs, 0)
    },
    tryConnectWithFallback(wsURLs: string[], index: number) {
      if (this.manualClose) {
        return
      }

      if (index >= wsURLs.length) {
        this.wsStatus = "disconnected"
        this.reconnectTimer = setTimeout(() => {
          this.connect()
        }, 1800)
        return
      }

      const wsURL = wsURLs[index]
      const ws = new WebSocket(wsURL)
      this.wsStatus = "connecting"
      this.ws = ws
      let opened = false

      ws.onopen = () => {
        if (this.ws !== ws) {
          return
        }
        opened = true
        this.wsStatus = "connected"
        this.errorMessage = ""

        void (async () => {
          try {
            await this.listPrivateConversations(false)
            await this.listGroups(false)
            if (this.activeKey) {
              await this.ensureHistoryLoaded(this.activeKey, true)
            }
          } catch {
            // keep ws alive even if refresh failed
          }
        })()
      }

      ws.onerror = () => {
        if (this.ws !== ws) {
          return
        }
        this.errorMessage = `实时连接异常（${parseHost(wsURL)}）`
      }

      ws.onclose = () => {
        if (this.ws !== ws) {
          return
        }
        this.wsStatus = "disconnected"
        this.ws = null

        if (this.manualClose) {
          return
        }

        if (!opened && index + 1 < wsURLs.length) {
          this.tryConnectWithFallback(wsURLs, index + 1)
          return
        }

        // Exponential backoff or simple fixed retry
        this.reconnectTimer = setTimeout(() => {
          this.connect()
        }, 3000) // increased from 1800 to 3000
      }

// ... inside onmessage ...
      ws.onmessage = (event: MessageEvent<string>) => {
        if (this.ws !== ws) {
          return
        }
        try {
          const data = JSON.parse(event.data) as WSIncomingEnvelope
          if (data.type === "message") {
            this.ingestMessage(data.payload as ChatMessage)
          } else if (data.type === "friend_request") {
            const friendStore = useFriendStore()
            const req = data.payload as any
            friendStore.addRequest(req) // Type cast if needed or update envelope type
            const title = "新好友申请"
            const body = `${req.from_user?.display_name || req.from_user?.username || "有人"} 请求添加您为好友`
            this.addToast({ title, message: body, type: "info" })
          } else if (data.type === "friend_accepted") {
            const friendStore = useFriendStore()
            friendStore.loadFriends()
            
            // Clear relationship error if we were blocked
            if (this.errorMessage === "not friend") {
              this.errorMessage = ""
            }

            const payload = data.payload as any
            const title = "好友申请已通过"
            const body = `${payload.friend?.display_name || payload.friend?.username || "对方"} 已同意您的好友请求`
            this.addToast({ title, message: body, type: "success" })
          } else if (data.type === "user_online_status") {
            const payload = data.payload as UserOnlineStatusPayload
            // Update the reactive cache immediately
            this.userOnlineStatusCache = {
              ...this.userOnlineStatusCache,
              [payload.user_id]: {
                user_id: payload.user_id,
                online: payload.online,
                last_seen_at: payload.last_seen_at
              }
            }
          } else if (data.type === "error") {
            this.errorMessage = (data.payload as { message?: string })?.message || "消息处理失败"
          }
        } catch {
          this.errorMessage = "收到无法解析的消息"
        }
      }
    },
    disconnect() {
      this.manualClose = true
      if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer)
        this.reconnectTimer = null
      }
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.wsStatus = "disconnected"
    },
    sendMessage(content: string) {
      const conv = this.activeConversation
      if (!conv) {
        throw new Error("请先选择会话")
      }
      const text = content.trim()
      if (!text) {
        return
      }
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error("WebSocket 未连接")
      }

      this.ws.send(
        JSON.stringify({
          type: "chat",
          payload: {
            target_id: conv.targetId,
            chat_type: conv.chatType,
            msg_type: 1,
            content: text,
          },
        }),
      )
    },
    async sendImage(file: File) {
      const conv = this.activeConversation
      if (!conv) {
        throw new Error("请先选择会话")
      }

      const auth = useAuthStore()
      if (!auth.token) {
        throw new Error("请先登录")
      }

      if (!auth.user) {
        throw new Error("用户信息未初始化")
      }

      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error("WebSocket 未连接")
      }

      const key = makeKey(conv.chatType, conv.targetId)
      const pendingMessage = createPendingImageMessage({
        previewUrl: URL.createObjectURL(file),
        targetId: conv.targetId,
        chatType: conv.chatType,
        sender: auth.user,
      })
      const bucket = this.messagesByKey[key] || []
      bucket.push(pendingMessage)
      this.messagesByKey[key] = bucket
      this.touchConversation(key, pendingMessage.content, pendingMessage.created_at, pendingMessage.msg_type)

      try {
        const uploaded = await uploadImage(file, auth.token)
        if (!uploaded.url) {
          throw new Error("上传失败")
        }

        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
          throw new Error("WebSocket 未连接")
        }

        this.ws.send(
          JSON.stringify({
            type: "chat",
            payload: {
              target_id: conv.targetId,
              chat_type: conv.chatType,
              msg_type: 2,
              content: uploaded.url,
            },
          }),
        )
      } catch (error) {
        const currentBucket = this.messagesByKey[key] || []
        const pendingIndex = currentBucket.findIndex((item) => item.id === pendingMessage.id)
        if (pendingIndex >= 0) {
          const [removed] = currentBucket.splice(pendingIndex, 1)
          if (removed?.content?.startsWith("blob:")) {
            URL.revokeObjectURL(removed.content)
          }
          this.messagesByKey[key] = currentBucket
        }
        throw error
      }
    },
  },
})
