import type { ChatMessage, ChatType, SenderInfo } from '../types/chat'

interface PendingImageInput {
  previewUrl: string
  targetId: number
  chatType: ChatType
  sender: SenderInfo
}

interface PendingImageMatchInput {
  senderId: number
  chatType: ChatType
  targetId: number
}

let pendingImageIdSeed = -1

export function createPendingImageMessage(input: PendingImageInput): ChatMessage {
  return {
    id: pendingImageIdSeed--,
    sender: input.sender,
    target_id: input.targetId,
    chat_type: input.chatType,
    msg_type: 2,
    content: input.previewUrl,
    created_at: new Date().toISOString(),
    uploading: true,
    uploadProgressLabel: '图片上传中...',
  }
}

export function findPendingImageReplacementIndex(
  messages: ChatMessage[],
  input: PendingImageMatchInput,
): number {
  return messages.findIndex((message) =>
    Boolean(message.uploading) &&
    Number(message.msg_type) === 2 &&
    message.chat_type === input.chatType &&
    Number(message.target_id) === input.targetId &&
    Number(message.sender?.id) === input.senderId,
  )
}
