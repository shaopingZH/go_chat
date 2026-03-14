export type ChatType = "private" | "group"

export interface UserProfile {
  id: number
  username: string
  avatar: string
  display_name?: string
  bio?: string
}

export interface LoginResponse {
  token: string
  user: UserProfile
}

export interface MessageRecord {
  id: number
  sender_id: number
  sender?: SenderInfo
  receiver_id: number | null
  group_id: number | null
  content: string
  msg_type: number
  created_at: string
}

export interface SenderInfo {
  id: number
  username: string
  display_name?: string
  avatar: string
}

export interface ChatMessage {
  id: number
  sender: SenderInfo
  target_id: number
  chat_type: ChatType
  msg_type: number
  content: string
  created_at: string
  uploading?: boolean
  uploadProgressLabel?: string
}

export interface ConversationItem {
  key: string
  chatType: ChatType
  targetId: number
  title: string
  avatar?: string
  unread: number
  lastText: string
  lastAt: string
  owner_id?: number // Added for group ownership
}

export interface GroupItem {
  id: number
  name: string
  owner_id: number
  created_at: string
  unread_count?: number
  last_read_message_id?: number
  last_message?: GroupLastMessage
}

export interface GroupListResponse {
  items: GroupItem[]
  unread_count?: number
  last_read_message_id?: number
}

export interface GroupLastMessage {
  id: number
  sender_id: number
  msg_type: number
  content: string
  created_at: string
}
export interface GroupMember {
  id: number
  group_id: number
  user_id?: number // if backend returned user_id
  user?: UserProfile // if backend returned nested user
  // Flat properties backend actually returns:
  username?: string
  display_name?: string
  avatar?: string
}

export interface PrivateConversationSummary {
  partner_id: number
  partner_username: string
  partner_avatar: string
  unread_count: number
  last_message_id: number
  last_sender_id: number
  last_msg_type: number
  last_content: string
  last_created_at: string
}

export interface PrivateConversationResponse {
  items: PrivateConversationSummary[]
}

export interface WSIncomingEnvelope {
  type: "message" | "error" | "friend_request" | "friend_accepted" | "user_online_status"
  payload: ChatMessage | { message?: string } | FriendRequest | FriendAcceptedPayload | UserOnlineStatusPayload
}

export interface UserOnlineStatusPayload {
  user_id: number
  online: boolean
  last_seen_at?: string
}

export interface UserOnlineStatus {
  user_id: number
  online: boolean
  last_seen_at?: string
}

export interface FriendUser {
  id: number
  username: string
  display_name: string
  avatar: string
}

export interface FriendRequest {
  id: number
  from_user: FriendUser
  status: "pending" | "accepted" | "rejected"
  created_at: string
}

export interface FriendAcceptedPayload {
  request_id: number
  friend: FriendUser
  accepted_at: string
}
