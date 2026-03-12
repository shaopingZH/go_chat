<template>
  <div class="tg-item relative" :class="{ 'tg-item--active': active }" @click="$emit('select')" @contextmenu.prevent="onContextMenu">
    <div class="tg-item-avatar" :style="{ background: avatarBg }">
      <img v-if="displayAvatar" :src="displayAvatar" class="w-full h-full rounded-full object-cover" alt="avatar">
      <span v-else>{{ avatarText }}</span>
    </div>
    <div class="tg-item-content">
      <div class="tg-item-top">
        <span class="tg-item-title">{{ displayTitle }}</span>
        <span class="tg-item-time">{{ displayTime }}</span>
      </div>
      <div class="tg-item-bottom">
        <p class="tg-item-subtitle">{{ subtitle || "暂无消息" }}</p>
        <span v-if="unread > 0" class="tg-badge">{{ unread > 99 ? "99+" : unread }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import type { ChatType } from "../types/chat"
import { useChatStore } from "../stores/chat"

const chatStore = useChatStore()

const props = withDefaults(
  defineProps<{
    targetId: number
    title: string
    subtitle?: string
    unread?: number
    active?: boolean
    lastAt?: string
    chatType: ChatType
    avatar?: string
  }>(),
  {
    subtitle: "",
    unread: 0,
    active: false,
    lastAt: "",
    avatar: "",
  },
)

const emit = defineEmits<{ 
  select: []
  contextmenu: [MouseEvent]
}>()

function onContextMenu(event: MouseEvent) {
  emit('contextmenu', event)
}

const displayTitle = computed(() => {
  if (props.chatType === 'private') {
    const profile = chatStore.userProfileCache[props.targetId]
    if (profile) return profile.display_name || profile.username
  }
  return props.title
})

const displayAvatar = computed(() => {
  if (props.chatType === 'private') {
    const profile = chatStore.userProfileCache[props.targetId]
    if (profile?.avatar) return profile.avatar
  }
  return props.avatar
})

const avatarText = computed(() => {
  if (displayAvatar.value) return ""
  const source = displayTitle.value.trim()
  if (!source) {
    return "?"
  }
  return source.replace(/^#\s*/, "").slice(0, 1).toUpperCase()
})

const avatarBg = computed(() => {
  if (displayAvatar.value) return "transparent"
  // Simple hash for consistency
  let hash = 0
  for (let i = 0; i < displayTitle.value.length; i++) {
    hash = displayTitle.value.charCodeAt(i) + ((hash << 5) - hash)
  }
  const index = (Math.abs(hash) % 7) + 1 // 1-7
  return `var(--tg-avatar-${index})`
})

const displayTime = computed(() => {
  if (!props.lastAt) {
    return ""
  }
  const date = new Date(props.lastAt)
  if (Number.isNaN(date.getTime())) {
    return ""
  }
  
  const now = new Date()
  const isToday = now.toDateString() === date.toDateString()
  
  if (isToday) {
    return `${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`
  }
  
  return `${date.getMonth() + 1}/${date.getDate()}`
})
</script>
