<template>
  <div class="tg-bubble-row" :class="mine ? 'tg-bubble-row--mine' : 'tg-bubble-row--other'" :data-message-id="message.id">
    <!-- Avatar for Group Chat (other users only) -->
    <div 
      v-if="!mine && message.chat_type === 'group'" 
      class="w-10 h-10 rounded-full overflow-hidden flex-shrink-0 mt-auto mb-1 mr-2 cursor-pointer border border-[var(--tg-border)] flex items-center justify-center select-none"
      :style="{ background: avatarBg }"
      @click="showSenderProfile"
    >
      <img v-if="message.sender?.avatar" :src="message.sender.avatar" class="w-full h-full object-cover">
      <span v-else class="text-[13px] font-bold text-white uppercase">
        {{ (message.sender?.display_name || message.sender?.username || "?")[0] }}
      </span>
    </div>

    <div class="tg-bubble bubble-enter" :class="[mine ? 'tg-bubble--mine' : 'tg-bubble--other', { 'tg-bubble--image-only': isImageMessage }]">
      <!-- Sender Name (Group only, Other only) -->
      <div v-if="!mine && message.chat_type === 'group'" class="text-[14px] font-medium text-[#a695e7] mb-1 leading-none cursor-pointer" :class="{ 'px-3 pt-2': isImageMessage }" @click="showSenderProfile">
        {{ message.sender?.display_name || message.sender?.username || "Unknown" }}
      </div>
      
      <!-- Content -->
      <div v-if="isImageMessage" class="relative group">
        <img v-if="!imageLoadError" :src="message.content" alt="image message" class="tg-image-message" :class="{ 'tg-image-message--loading': imageLoading || message.uploading }" @click="handlePreview" @error="onImageError" @load="onImageLoad">
        <div v-else class="tg-image-fallback">图片加载失败</div>

        <Transition name="fade">
          <div v-if="showImageOverlay" class="tg-image-upload-overlay">
            <div class="tg-image-upload-spinner"></div>
            <span class="tg-image-upload-label">{{ imageOverlayLabel }}</span>
          </div>
        </Transition>
        
        <!-- Meta (Time) floating on image -->
        <span class="tg-bubble-meta tg-bubble-meta--floating">
          {{ timeText }}
        </span>
      </div>
      <template v-else>
        <span class="whitespace-pre-wrap tg-bubble-text" v-html="highlightedContent"></span>
        <!-- Meta (Time) normal -->
        <span class="tg-bubble-meta">
          {{ timeText }}
        </span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import type { ChatMessage } from "../types/chat"
import { useChatStore } from "../stores/chat"

const chat = useChatStore()

const props = defineProps<{
  message: ChatMessage
  mine: boolean
}>()

const emit = defineEmits<{
  preview: [url: string]
  imageLoaded: []
}>()

const imageLoadError = ref(false)
const imageLoading = ref(true)

const isImageMessage = computed(() => Number(props.message.msg_type) === 2)

watch(
  () => [props.message.id, props.message.content],
  () => {
    imageLoadError.value = false
    imageLoading.value = true
  },
)

function onImageError() {
  imageLoadError.value = true
  imageLoading.value = false
}

function onImageLoad() {
  imageLoading.value = false
  emit('imageLoaded')
}

function handlePreview() {
  if (props.message.uploading) {
    return
  }
  emit('preview', props.message.content)
}

const avatarBg = computed(() => {
  if (props.message.sender?.avatar) return "transparent"
  const name = props.message.sender?.display_name || props.message.sender?.username || "?"
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  const index = (Math.abs(hash) % 7) + 1
  return `var(--tg-avatar-${index})`
})

function showSenderProfile() {
  const sid = props.message.sender?.id || 0
  if (sid) {
    chat.openRightSidebar('publicProfile', sid)
  }
}

const highlightedContent = computed(() => {
  const content = props.message.content
  const keyword = chat.searchKeyword.trim()
  
  if (!chat.isSearching || !keyword) {
    return escapeHtml(content)
  }

  const escapedContent = escapeHtml(content)
  const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escapedKeyword})`, 'gi')
  
  return escapedContent.replace(regex, '<mark class="tg-search-mark">$1</mark>')
})

function escapeHtml(unsafe: string) {
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;")
}

const timeText = computed(() => {
  if (!props.message.created_at) return ""
  const date = new Date(props.message.created_at)
  if (Number.isNaN(date.getTime())) return ""
  return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`
})

const showImageOverlay = computed(() => !imageLoadError.value && (Boolean(props.message.uploading) || imageLoading.value))

const imageOverlayLabel = computed(() => props.message.uploadProgressLabel || '图片加载中...')
</script>

<style scoped>
.bubble-enter {
  animation: bubble-pop 0.3s cubic-bezier(0.2, 0, 0, 1) forwards;
  will-change: transform, opacity;
}

.tg-image-message--loading {
  filter: saturate(0.9) brightness(0.82);
}

.tg-image-upload-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.14), rgba(15, 23, 42, 0.48));
  backdrop-filter: blur(2px);
}

.tg-image-upload-spinner {
  width: 28px;
  height: 28px;
  border-radius: 9999px;
  border: 2.5px solid rgba(255, 255, 255, 0.28);
  border-top-color: rgba(255, 255, 255, 0.96);
  animation: tg-image-spin 0.9s linear infinite;
}

.tg-image-upload-label {
  padding: 4px 10px;
  border-radius: 9999px;
  background: rgba(15, 23, 42, 0.52);
  color: rgba(255, 255, 255, 0.96);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.01em;
}

@keyframes bubble-pop {
  0% {
    opacity: 0;
    transform: translateY(10px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes tg-image-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
