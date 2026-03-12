<template>
  <div class="flex flex-col h-full bg-[var(--tg-bg-sidebar)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
      <button class="tg-icon-btn -ml-2" @click="chat.openLeftSidebar('groupOptions')">
        <IconBack />
      </button>
      <h2 class="text-lg font-bold text-[var(--tg-text-white)]">Join Group</h2>
    </div>

    <div class="p-6 flex flex-col gap-8">
      <!-- Illustration -->
      <div class="flex flex-col items-center gap-4">
        <div class="w-24 h-24 rounded-3xl bg-gradient-to-br from-orange-400 to-orange-600 flex items-center justify-center text-white text-4xl font-bold shadow-lg shadow-orange-500/20">
           #
        </div>
        <div class="text-center">
          <h3 class="text-[var(--tg-text-white)] font-bold text-lg">Enter Group ID</h3>
          <p class="text-xs text-[var(--tg-text-gray)] mt-1">Ask the owner for the numeric ID</p>
        </div>
      </div>

      <!-- Input Field -->
      <div class="flex flex-col gap-2">
        <label class="text-[13px] font-medium text-[var(--tg-text-blue)] ml-1">Group Identifier</label>
        <div class="relative group">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <span class="text-[var(--tg-text-gray)] font-bold text-lg">#</span>
          </div>
          <input 
            v-model="groupId"
            class="w-full bg-[var(--tg-bg-chat)] border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] rounded-2xl pl-10 pr-4 py-3.5 text-[var(--tg-text-white)] focus:outline-none placeholder:text-[var(--tg-text-hint)] text-[16px]"
            type="number"
            placeholder=""
            @keydown.enter="handleJoin"
          />
        </div>
      </div>

      <!-- Action Button -->
      <button 
        class="tg-btn-primary w-full py-4 mt-4 flex items-center justify-center gap-2 shadow-lg shadow-[var(--tg-primary)]/20"
        :disabled="!groupId || loading"
        @click="handleJoin"
      >
        <span v-if="loading">Joining Group...</span>
        <span v-else>Join Group</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useChatStore } from "../stores/chat"
import IconBack from "./icons/IconBack.vue"

const chat = useChatStore()
const groupId = ref<number | "">("")
const loading = ref(false)

async function handleJoin() {
  if (!groupId.value || loading.value) return
  
  loading.value = true
  try {
    await chat.joinGroup(Number(groupId.value))
    chat.closeLeftSidebar()
  } catch (error) {
    alert(error instanceof Error ? error.message : "Failed to join group")
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* Hide spin buttons for numeric input */
input::-webkit-outer-spin-button,
input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type=number] {
  -moz-appearance: textfield;
}
</style>
