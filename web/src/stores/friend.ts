import { defineStore } from "pinia"
import { useAuthStore } from "./auth"
import { useChatStore } from "./chat"
import {
  searchUsers,
  sendFriendRequest,
  listFriendRequests,
  handleFriendRequest,
  listFriends,
  deleteFriend,
} from "../api/http"
import type { FriendUser, FriendRequest } from "../types/chat"

export const useFriendStore = defineStore("friend", {
  state: () => ({
    friends: [] as FriendUser[],
    requests: [] as FriendRequest[],
    searchResults: [] as FriendUser[],
    loading: false,
    error: "",
  }),
  actions: {
    async search(query: string) {
      const auth = useAuthStore()
      if (!auth.token || !query.trim()) return
      
      this.loading = true
      this.error = ""
      try {
        this.searchResults = await searchUsers(auth.token, query)
      } catch (err) {
        this.error = "Search failed"
        this.searchResults = []
      } finally {
        this.loading = false
      }
    },

    async sendRequest(targetId: number) {
      const auth = useAuthStore()
      if (!auth.token) return

      try {
        await sendFriendRequest(auth.token, targetId)
        // Optimistic update or refresh? Refresh for now
        // But search results might need to reflect "requested" state.
        // For MVP, just alerting or assume success.
      } catch (err) {
        this.error = err instanceof Error ? err.message : "Failed to send request"
        throw err
      }
    },

    async loadRequests() {
      const auth = useAuthStore()
      if (!auth.token) return

      try {
        this.requests = await listFriendRequests(auth.token)
      } catch (err) {
        // silent fail
      }
    },

    async loadFriends() {
      const auth = useAuthStore()
      if (!auth.token) return

      try {
        this.friends = await listFriends(auth.token)
      } catch (err) {
        // silent fail
      }
    },

    async handleRequest(requestId: number, action: "accept" | "reject") {
      const auth = useAuthStore()
      if (!auth.token) return

      try {
        await handleFriendRequest(auth.token, requestId, action)
        // Remove from local list
        this.requests = this.requests.filter(r => r.id !== requestId)
        
        if (action === "accept") {
          await this.loadFriends() // Refresh friends list
          const chatStore = useChatStore()
          if (chatStore.errorMessage === "not friend") {
            chatStore.errorMessage = ""
          }
        }
      } catch (err) {
        this.error = `Failed to ${action} request`
        throw err
      }
    },

    async removeFriend(friendId: number) {
      const auth = useAuthStore()
      if (!auth.token) return

      try {
        await deleteFriend(auth.token, friendId)
        this.friends = this.friends.filter(f => f.id !== friendId)
      } catch (err) {
        this.error = "Failed to remove friend"
        throw err
      }
    },
    
    // Called by WebSocket
    addRequest(req: FriendRequest) {
      // Avoid duplicate
      if (!this.requests.some(r => r.id === req.id)) {
        this.requests.unshift(req)
      }
    }
  },
})
