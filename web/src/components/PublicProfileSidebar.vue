<template>
  <div class="h-full flex flex-col bg-[var(--tg-bg-sidebar)] border-l border-[var(--tg-border)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
      <button class="tg-icon-btn -ml-2" @click="chat.closeRightSidebar()">
        <IconBack class="rotate-180" />
      </button>
      <h2 class="text-lg font-bold text-[var(--tg-text-white)]">User Profile</h2>
    </div>

    <!-- Content -->
    <div v-if="user" class="flex-1 p-6 flex flex-col gap-6 overflow-y-auto">
        <!-- Avatar Section -->
        <div class="flex flex-col items-center gap-4">
          <div class="w-24 h-24 rounded-full overflow-hidden border-2 border-[var(--tg-border)] shadow-lg bg-[var(--tg-bg-sidebar)] flex items-center justify-center">
            <img v-if="user.avatar" :src="user.avatar" class="w-full h-full object-cover">
            <span v-else class="text-3xl font-bold text-[var(--tg-text-gray)]">{{ (user.username || "?")[0].toUpperCase() }}</span>
          </div>
          <div class="flex flex-col items-center">
            <h3 class="text-xl font-bold text-[var(--tg-text-white)]">{{ user.display_name || user.username }}</h3>
            <p v-if="user.display_name" class="text-sm text-[var(--tg-text-gray)]">@{{ user.username }}</p>
          </div>
        </div>

        <!-- Info -->
        <div class="flex flex-col gap-4 border-b border-t border-[var(--tg-border)] py-4">
          <div class="flex flex-col gap-1">
            <label class="text-xs text-[var(--tg-text-blue)] font-bold uppercase">Bio</label>
            <p class="text-[var(--tg-text-white)] text-sm leading-relaxed whitespace-pre-wrap">{{ user.bio || "No bio available." }}</p>
          </div>
        </div>

        <!-- Actions -->
        <div v-if="shouldShowMessageButton" class="flex flex-col gap-3 mt-2">
          <button 
            class="tg-btn-primary w-full flex items-center justify-center gap-2"
            @click="sendMessage"
          >
            <IconSend class="w-5 h-5" />
            <span>Send Message</span>
          </button>
        </div>
      </div>

    <!-- Error State -->
    <div v-else class="p-10 flex flex-col items-center text-[var(--tg-text-gray)] gap-4">
      <span>User not found</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue"
import { useAuthStore } from "../stores/auth"
import { useChatStore } from "../stores/chat"
import type { UserProfile } from "../types/chat"
import IconBack from "../components/icons/IconBack.vue"
import IconSend from "../components/icons/IconSend.vue"

const auth = useAuthStore()
const chat = useChatStore()

const shouldShowMessageButton = computed(() => {
  const current = chat.activeConversation
  const profileId = Number(chat.rightSidebarPayload)
  if (!current) return true
  // If we are already in a private chat with THIS user, don't show the button
  return !(current.chatType === 'private' && current.targetId === profileId)
})



const userId = computed(() => Number(chat.rightSidebarPayload))
const user = computed<UserProfile | null>(() => {
  if (!userId.value) {
    return null
  }
  if (chat.userProfileCache[userId.value]) {
    return chat.userProfileCache[userId.value]
  }

  // Return a dummy profile immediately from conversation info, will be updated when background fetch completes
  const conv = chat.conversations.find(c => c.chatType === 'private' && c.targetId === userId.value)
  if (conv) {
    return {
      id: userId.value,
      username: conv.title || `User ${userId.value}`,
      display_name: conv.title || `User ${userId.value}`,
      avatar: conv.avatar || '',
      bio: '',
      created_at: ''
    } as UserProfile
  }
  
  return {
    id: userId.value,
    username: `User ${userId.value}`,
    display_name: `User ${userId.value}`,
    avatar: '',
    bio: '',
    created_at: ''
  } as UserProfile
})

watch(userId, async (newId) => {
  if (!newId || !auth.token) {
    return
  }

  // If viewing self, redirect to edit
  if (auth.user && auth.user.id === newId) {
    chat.openRightSidebar('profile')
    return
  }

  try {
    await chat.getUserProfile(newId)
  } catch (error) {
    console.error("Failed to load profile", error)
  }
}, { immediate: true })

async function sendMessage() {
  if (!user.value) return
  
  try {
    await chat.openPrivateByUserID(user.value.id)
    chat.closeRightSidebar()
  } catch (error) {
    console.error("Failed to start chat", error)
    alert("Failed to start chat")
  }
}
</script>
