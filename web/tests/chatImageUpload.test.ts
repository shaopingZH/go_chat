import test from 'node:test'
import assert from 'node:assert/strict'

import type { ChatMessage } from '../src/types/chat.ts'
import {
  createPendingImageMessage,
  findPendingImageReplacementIndex,
  resolveImageVisualState,
} from '../src/utils/chatImageUpload.ts'

test('createPendingImageMessage creates uploading image placeholder', () => {
  const message = createPendingImageMessage({
    previewUrl: 'blob:preview-1',
    targetId: 8,
    chatType: 'private',
    sender: {
      id: 1,
      username: 'alice',
      display_name: 'Alice',
      avatar: '',
    },
  })

  assert.equal(message.msg_type, 2)
  assert.equal(message.content, 'blob:preview-1')
  assert.equal(message.uploading, true)
  assert.equal(message.uploadProgressLabel, '图片上传中...')
  assert.ok(message.id < 0)
})

test('findPendingImageReplacementIndex finds earliest self uploading image in same chat', () => {
  const messages = [
    { id: -2, msg_type: 2, uploading: true, sender: { id: 1 }, chat_type: 'private', target_id: 8 },
    { id: 5, msg_type: 1, sender: { id: 2 }, chat_type: 'private', target_id: 8 },
    { id: -3, msg_type: 2, uploading: true, sender: { id: 1 }, chat_type: 'private', target_id: 8 },
  ] as ChatMessage[]

  const index = findPendingImageReplacementIndex(messages, {
    senderId: 1,
    chatType: 'private',
    targetId: 8,
  })

  assert.equal(index, 0)
})

test('findPendingImageReplacementIndex ignores non-uploading, other sender, or other chat placeholders', () => {
  const messages = [
    { id: -2, msg_type: 2, uploading: false, sender: { id: 1 }, chat_type: 'private', target_id: 8 },
    { id: -3, msg_type: 2, uploading: true, sender: { id: 2 }, chat_type: 'private', target_id: 8 },
    { id: -4, msg_type: 2, uploading: true, sender: { id: 1 }, chat_type: 'group', target_id: 8 },
  ] as ChatMessage[]

  const index = findPendingImageReplacementIndex(messages, {
    senderId: 1,
    chatType: 'private',
    targetId: 8,
  })

  assert.equal(index, -1)
})

test('resolveImageVisualState hides loading overlay for already completed non-uploading image', () => {
  const state = resolveImageVisualState({
    uploading: false,
    complete: true,
    naturalWidth: 320,
  })

  assert.deepEqual(state, { loading: false, error: false })
})

test('resolveImageVisualState keeps overlay for uploading image even if preview is already complete', () => {
  const state = resolveImageVisualState({
    uploading: true,
    complete: true,
    naturalWidth: 320,
  })

  assert.deepEqual(state, { loading: true, error: false })
})
