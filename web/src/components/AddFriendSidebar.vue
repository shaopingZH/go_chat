<template>
  <div class="flex flex-col h-full bg-[var(--tg-bg-sidebar)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
      <button class="tg-icon-btn -ml-2" @click="chat.openLeftSidebar('friends')">
        <IconBack />
      </button>
      <h2 class="text-lg font-bold text-[var(--tg-text-white)]">Add Friend</h2>
    </div>

    <!-- Search Input -->
    <div class="p-4 border-b border-[var(--tg-border)]">
      <div class="relative group">
        <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
          <IconSearch class="text-[var(--tg-text-gray)] w-5 h-5" />
        </div>
        <input 
          v-model="query" 
          class="w-full bg-[var(--tg-bg-chat)] border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] rounded-2xl pl-11 pr-4 py-2.5 text-[var(--tg-text-white)] focus:outline-none placeholder:text-[var(--tg-text-hint)] text-sm"
          type="text" 
          placeholder="Search by username..." 
          @input="onSearch"
        />
      </div>
    </div>

    <!-- Results -->
    <div class="flex-1 overflow-y-auto p-2">
      <div v-if="friendStore.loading" class="text-center text-[var(--tg-text-gray)] mt-10 text-sm">Searching...</div>
      <div v-else-if="friendStore.searchResults.length === 0 && query" class="text-center text-[var(--tg-text-gray)] mt-10 text-sm">No users found.</div>
      
      <div 
        v-for="user in friendStore.searchResults" 
        :key="user.id"
        class="tg-item group hover:bg-[var(--tg-bg-hover)] transition-colors rounded-xl"
      >
        <div class="tg-item-avatar" :style="{ background: getAvatarColor(user.id) }">
          <img v-if="user.avatar" :src="user.avatar" class="w-full h-full rounded-full object-cover">
          <span v-else>{{ (user.display_name || user.username || "?")[0].toUpperCase() }}</span>
        </div>
        <div class="tg-item-content">
          <div class="tg-item-top">
            <span class="tg-item-title">{{ user.display_name || user.username }}</span>
          </div>
          <div class="tg-item-bottom">
            <p class="tg-item-subtitle text-xs">@{{ user.username }}</p>
          </div>
        </div>
        <button 
          class="px-4 py-1.5 rounded-full text-xs font-bold transition-all shadow-md"
          :class="[
            requestStatus[user.id] === 'sent' 
              ? 'bg-[var(--tg-bg-hover)] text-[var(--tg-text-gray)] shadow-none cursor-default' 
              : 'bg-[var(--tg-primary)] text-white hover:opacity-90 active:scale-[0.98] shadow-[var(--tg-primary)]/10'
          ]"
          :disabled="requestStatus[user.id] === 'sent' || requestStatus[user.id] === 'sending'"
          @click="addFriend(user.id)"
        >
          {{ requestStatus[user.id] === 'sending' ? '...' : (requestStatus[user.id] === 'sent' ? 'Sent' : 'Add') }}
        </button>
      </div>
      
      <!-- Error Toast -->
      <Transition name="fade">
        <div v-if="errorMessage" class="absolute bottom-6 left-1/2 -translate-x-1/2 bg-[var(--tg-bg-chat)] border border-[var(--tg-border)] shadow-xl px-4 py-2 rounded-xl text-sm text-[var(--tg-text-white)] z-50">
          {{ errorMessage }}
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useFriendStore } from "../stores/friend"
import { useChatStore } from "../stores/chat"
import IconBack from "./icons/IconBack.vue"
import IconSearch from "./icons/IconSearch.vue"

const friendStore = useFriendStore()
const chat = useChatStore()
const query = ref("")
const errorMessage = ref("")
const requestStatus = ref<Record<number, 'sending' | 'sent'>>({})
let debounceTimer: ReturnType<typeof setTimeout>
let errorTimer: ReturnType<typeof setTimeout>

function getAvatarColor(id: number) {
  const colors = [
    'var(--tg-avatar-1)', 'var(--tg-avatar-2)', 'var(--tg-avatar-3)', 
    'var(--tg-avatar-4)', 'var(--tg-avatar-5)', 'var(--tg-avatar-6)', 'var(--tg-avatar-7)'
  ]
  return colors[id % colors.length]
}

function onSearch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (query.value.trim()) {
      friendStore.search(query.value)
      // Reset statuses on new search
      requestStatus.value = {}
    }
  }, 300)
}

function showError(msg: string) {
  errorMessage.value = msg
  clearTimeout(errorTimer)
  errorTimer = setTimeout(() => {
    errorMessage.value = ""
  }, 3000)
}

async function addFriend(userId: number) {
  if (requestStatus.value[userId] === 'sent' || requestStatus.value[userId] === 'sending') return
  
  try {
    requestStatus.value[userId] = 'sending'
    await friendStore.sendRequest(userId)
    requestStatus.value[userId] = 'sent'
  } catch (error) {
    delete requestStatus.value[userId]
    const msg = error instanceof Error ? error.message : "Failed to send request"
    showError(msg)
  }
}
</script>
