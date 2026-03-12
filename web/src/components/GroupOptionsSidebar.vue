<template>
  <div class="flex flex-col h-full bg-[var(--tg-bg-sidebar)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
      <button class="tg-icon-btn -ml-2" @click="chat.closeLeftSidebar()">
        <IconBack />
      </button>
      <h2 class="text-lg font-bold text-[var(--tg-text-white)]">New Action</h2>
    </div>

    <!-- Options List -->
    <div class="flex-1 p-2">
      <div 
        class="tg-item group hover:bg-[var(--tg-bg-hover)] transition-colors rounded-xl"
        @click="chat.openLeftSidebar('createGroup')"
      >
        <div class="tg-item-avatar bg-[var(--tg-text-blue)]">
          <span class="text-white text-xl">+</span>
        </div>
        <div class="tg-item-content py-2">
          <span class="tg-item-title text-[16px]">New Group</span>
          <p class="tg-item-subtitle text-xs">Create a new private or public group</p>
        </div>
      </div>

      <div 
        class="tg-item group hover:bg-[var(--tg-bg-hover)] transition-colors rounded-xl"
        @click="chat.openLeftSidebar('joinGroup')"
      >
        <div class="tg-item-avatar bg-orange-500">
          <span class="text-white text-xl">#</span>
        </div>
        <div class="tg-item-content py-2">
          <span class="tg-item-title text-[16px]">Join Group</span>
          <p class="tg-item-subtitle text-xs">Join an existing group via numeric ID</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useChatStore } from "../stores/chat"
import IconBack from "./icons/IconBack.vue"

const chat = useChatStore()
</script>
