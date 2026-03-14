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

interface ImageVisualStateInput {
  uploading: boolean
  complete: boolean
  naturalWidth: number
}

interface ImageOverlayPresentationInput {
  uploading: boolean
  loading: boolean
  uploadProgressLabel?: string
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

export function createDeferredImageMessage(pending: ChatMessage, delivered: ChatMessage): ChatMessage {
  return {
    ...delivered,
    content: pending.content,
    resolvedContent: delivered.content,
    uploading: false,
    uploadProgressLabel: undefined,
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

export function resolveImageVisualState(input: ImageVisualStateInput): { loading: boolean; error: boolean } {
  if (input.uploading) {
    return { loading: true, error: false }
  }

  if (!input.complete) {
    return { loading: true, error: false }
  }

  if (input.naturalWidth > 0) {
    return { loading: false, error: false }
  }

  return { loading: false, error: true }
}

export function resolveImageOverlayPresentation(input: ImageOverlayPresentationInput): {
  visible: boolean
  mode: 'uploading' | 'loading'
  label: string
  reserveSpace: boolean
} {
  if (input.uploading) {
    return {
      visible: true,
      mode: 'uploading',
      label: input.uploadProgressLabel || '图片上传中...',
      reserveSpace: false,
    }
  }

  return {
    visible: input.loading,
    mode: 'loading',
    label: '',
    reserveSpace: input.loading,
  }
}
