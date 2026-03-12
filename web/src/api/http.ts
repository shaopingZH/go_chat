const API_BASE = (import.meta.env.VITE_API_BASE || "").replace(/\/$/, "")

function normalizeBase(input: string | undefined): string {
  return (input || "").trim().replace(/\/$/, "")
}

function toWSBase(httpBase: string): string {
  return httpBase.replace(/^http:/, "ws:").replace(/^https:/, "wss:")
}

function pushUnique(list: string[], value: string): void {
  if (value && !list.includes(value)) {
    list.push(value)
  }
}

function toPath(path: string): string {
  if (!path.startsWith("/")) {
    return `/${path}`
  }
  return path
}

function toURL(path: string): string {
  const normalized = toPath(path)
  return `${API_BASE}${normalized}`
}

export async function apiRequest<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(toURL(path), options)
  const contentType = response.headers.get("content-type") || ""
  const isJSON = contentType.includes("application/json")
  const payload = (isJSON ? await response.json() : null) as T | { error?: string } | null

  if (!response.ok) {
    const message = (payload as { error?: string } | null)?.error || `request failed: ${response.status}`
    throw new Error(message)
  }

  return payload as T
}

export interface UploadImageResponse {
  url: string
  mime: string
  size: number
}

export async function uploadImage(file: File, token: string): Promise<UploadImageResponse> {
  const formData = new FormData()
  formData.append("file", file)

  const response = await fetch(toURL("/api/uploads/images"), {
    method: "POST",
    headers: authHeaders(token),
    body: formData,
  })

  const contentType = response.headers.get("content-type") || ""
  const isJSON = contentType.includes("application/json")
  const payload = (isJSON ? await response.json() : null) as UploadImageResponse | { error?: string } | null

  if (!response.ok) {
    const message = (payload as { error?: string } | null)?.error || `request failed: ${response.status}`
    throw new Error(message)
  }

  return payload as UploadImageResponse
}

export function authHeaders(token: string): HeadersInit {
  return {
    Authorization: `Bearer ${token}`,
  }
}

export function buildWSURLs(token: string): string[] {
  const encoded = encodeURIComponent(token)
  const urls: string[] = []

  const customWSBase = normalizeBase(import.meta.env.VITE_WS_BASE)
  if (customWSBase) {
    pushUnique(urls, `${customWSBase}/ws?token=${encoded}`)
  }

  const apiBase = normalizeBase(API_BASE)
  if (apiBase.startsWith("http://") || apiBase.startsWith("https://")) {
    pushUnique(urls, `${toWSBase(apiBase)}/ws?token=${encoded}`)
  }

  const devTarget = normalizeBase(import.meta.env.VITE_DEV_PROXY_TARGET || (import.meta.env.DEV ? "http://localhost:8080" : ""))
  if (devTarget.startsWith("http://") || devTarget.startsWith("https://")) {
    pushUnique(urls, `${toWSBase(devTarget)}/ws?token=${encoded}`)
  }

  const protocol = window.location.protocol === "https:" ? "wss" : "ws"
  pushUnique(urls, `${protocol}://${window.location.host}/ws?token=${encoded}`)

  if (import.meta.env.DEV) {
    const hostname = window.location.hostname
    const altHost = hostname === "localhost" ? "127.0.0.1" : hostname === "127.0.0.1" ? "localhost" : ""
    if (altHost) {
      pushUnique(urls, `${protocol}://${altHost}:8080/ws?token=${encoded}`)
    }
  }

  return urls
}

import type { UserProfile } from "../types/chat"

export async function patchMyProfile(token: string, payload: { display_name?: string; bio?: string; avatar?: string }): Promise<UserProfile> {
  return apiRequest<UserProfile>("/api/users/me", {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(token),
    },
    body: JSON.stringify(payload),
  })
}

export async function updateGroup(token: string, groupId: number, payload: { name?: string; avatar?: string }): Promise<any> {
  return apiRequest(`/api/groups/${groupId}`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(token),
    },
    body: JSON.stringify(payload),
  })
}

export async function getPublicProfile(token: string, userId: number): Promise<UserProfile> {
  return apiRequest<UserProfile>(`/api/users/${userId}/profile`, {
    method: "GET",
    headers: authHeaders(token),
  })
}

import type { FriendRequest, FriendUser, GroupMember } from "../types/chat"


export interface SearchUsersResponse {
  items: FriendUser[]
}

export async function searchUsers(token: string, query: string): Promise<FriendUser[]> {
  const data = await apiRequest<SearchUsersResponse>(`/api/users/search?keyword=${encodeURIComponent(query)}`, {
    headers: authHeaders(token),
  })
  return data.items || []
}

export async function sendFriendRequest(token: string, targetId: number): Promise<void> {
  await apiRequest("/api/friends/requests", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(token),
    },
    body: JSON.stringify({ target_id: targetId }),
  })
}

export interface ListFriendRequestsResponse {
  items: FriendRequest[]
}

export async function listFriendRequests(token: string): Promise<FriendRequest[]> {
  const data = await apiRequest<{ items: any[] }>("/api/friends/requests/pending", {
    headers: authHeaders(token),
  })
  // Backend returns request_id, mapping to id
  return (data.items || []).map(item => ({
    ...item,
    id: item.request_id || item.id
  }))
}

export async function handleFriendRequest(token: string, requestId: number, action: "accept" | "reject"): Promise<void> {
  await apiRequest(`/api/friends/requests/${requestId}/${action}`, {
    method: "PUT",
    headers: authHeaders(token),
  })
}

export interface ListFriendsResponse {
  items: FriendUser[]
}

export async function listFriends(token: string): Promise<FriendUser[]> {
  const data = await apiRequest<ListFriendsResponse>("/api/friends", {
    headers: authHeaders(token),
  })
  return data.items || []
}

export async function deleteFriend(token: string, friendId: number): Promise<void> {
  await apiRequest(`/api/friends/${friendId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  })
}
export interface GroupMemberResponse {
  items: GroupMember[]
}

export async function getGroupMembers(token: string, groupId: number): Promise<GroupMember[]> {
  const data = await apiRequest<GroupMemberResponse>(`/api/groups/${groupId}/members`, {
    headers: authHeaders(token),
  })
  return data.items || []
}

export async function leaveGroup(token: string, groupId: number): Promise<void> {
  await apiRequest(`/api/groups/${groupId}/leave`, {
    method: "POST",
    headers: authHeaders(token),
  })
}

export async function kickGroupMember(token: string, groupId: number, userId: number): Promise<void> {
  await apiRequest(`/api/groups/${groupId}/members/${userId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  })
}

export interface UserOnlineStatus {
  user_id: number
  online: boolean
  last_seen_at?: string
}

export async function batchGetUserOnlineStatus(token: string, userIds: number[]): Promise<UserOnlineStatus[]> {
  const data = await apiRequest<{ items: UserOnlineStatus[] }>("/api/users/online/batch", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(token),
    },
    body: JSON.stringify({ user_ids: userIds }),
  })
  return data.items || []
}

import type { ChatType } from "../types/chat"

export async function markConversationRead(token: string, chatType: ChatType, targetId: number): Promise<void> {
  await apiRequest("/api/conversations/read", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(token),
    },
    body: JSON.stringify({ chat_type: chatType, target_id: targetId }),
  })
}
