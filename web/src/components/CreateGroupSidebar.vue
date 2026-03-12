<template>
  <div class="flex flex-col h-full bg-[var(--tg-bg-sidebar)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
      <button class="tg-icon-btn -ml-2" @click="chat.openLeftSidebar('groupOptions')">
        <IconBack />
      </button>
      <h2 class="text-lg font-bold text-[var(--tg-text-white)]">New Group</h2>
    </div>

    <div class="p-6 flex flex-col gap-8">
      <!-- Illustration -->
      <div class="flex flex-col items-center gap-4">
        <div class="w-24 h-24 rounded-full bg-gradient-to-br from-[var(--tg-bg-active)] to-purple-600 flex items-center justify-center text-white text-4xl font-bold shadow-lg shadow-[var(--tg-bg-active)]/20 transition-transform duration-500 hover:scale-105">
           {{ (groupName || "?")[0].toUpperCase() }}
        </div>
        <div class="text-center">
          <h3 class="text-[var(--tg-text-white)] font-bold text-lg">Create a Group</h3>
          <p class="text-xs text-[var(--tg-text-gray)] mt-1">Up to 100 members</p>
        </div>
      </div>

      <!-- Input Field -->
      <div class="flex flex-col gap-2">
        <label class="text-[13px] font-medium text-[var(--tg-text-blue)] ml-1">Group Name</label>
        <div class="relative group">
          <input 
            v-model="groupName"
            class="w-full bg-[var(--tg-bg-chat)] border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] rounded-2xl px-4 py-3.5 text-[var(--tg-text-white)] focus:outline-none placeholder:text-[var(--tg-text-hint)] text-[16px]"
            type="text"
            placeholder="Enter group name"
            @keydown.enter="handleCreate"
          />
        </div>
      </div>

      <!-- Action Button -->
      <button 
        class="tg-btn-primary w-full py-4 mt-4 flex items-center justify-center gap-2"
        :disabled="!groupName.trim() || loading"
        @click="handleCreate"
      >
        <span v-if="loading">Creating Group...</span>
        <span v-else>Create Group</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useChatStore } from "../stores/chat"
import IconBack from "./icons/IconBack.vue"

const chat = useChatStore()
const groupName = ref("")
const loading = ref(false)

async function handleCreate() {
  if (!groupName.value.trim() || loading.value) return
  
  loading.value = true
  try {
    await chat.createGroup(groupName.value.trim())
    chat.closeLeftSidebar()
  } catch (error) {
    alert(error instanceof Error ? error.message : "Failed to create group")
  } finally {
    loading.value = false
  }
}
</script>
