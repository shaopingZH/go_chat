<template>
  <div class="flex flex-col h-full bg-[var(--tg-bg-sidebar)]">
    <!-- Header -->
    <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0 justify-between">
      <div class="flex items-center gap-4">
        <button class="tg-icon-btn -ml-2" @click="chat.closeLeftSidebar()">
          <IconBack />
        </button>
        <h2 class="text-lg font-bold text-[var(--tg-text-white)]">Friends</h2>
      </div>
      <button class="tg-btn-ghost text-sm" @click="chat.openLeftSidebar('addFriend')">Add New</button>
    </div>

    <!-- Tabs -->
    <div class="flex border-b border-[var(--tg-border)] flex-shrink-0">
      <button 
        class="flex-1 py-3 text-sm font-bold transition-colors relative"
        :class="activeTab === 'friends' ? 'text-[var(--tg-text-blue)]' : 'text-[var(--tg-text-gray)] hover:text-[var(--tg-text-white)]'"
        @click="activeTab = 'friends'"
      >
        My Friends
        <span v-if="activeTab === 'friends'" class="absolute bottom-0 left-0 w-full h-[2px] bg-[var(--tg-text-blue)]"></span>
      </button>
      <button 
        class="flex-1 py-3 text-sm font-bold transition-colors relative"
        :class="activeTab === 'requests' ? 'text-[var(--tg-text-blue)]' : 'text-[var(--tg-text-gray)] hover:text-[var(--tg-text-white)]'"
        @click="activeTab = 'requests'"
      >
        Requests
        <span v-if="friendStore.requests.length > 0" class="ml-1 bg-red-500 text-white text-[10px] px-1.5 py-0.5 rounded-full">{{ friendStore.requests.length }}</span>
        <span v-if="activeTab === 'requests'" class="absolute bottom-0 left-0 w-full h-[2px] bg-[var(--tg-text-blue)]"></span>
      </button>
    </div>

    <!-- Content -->
    <div class="flex-1 relative overflow-hidden">
      <Transition :name="tabTransition">
        <!-- Friends Tab -->
        <div v-if="activeTab === 'friends'" key="friends" class="absolute inset-0 overflow-y-auto p-2">
          <div v-if="friendStore.friends.length === 0" class="text-center text-[var(--tg-text-gray)] mt-10 px-4">
            <p class="text-sm">No friends yet.</p>
            <button class="tg-btn-ghost mt-2 text-sm" @click="chat.openLeftSidebar('addFriend')">Find someone!</button>
          </div>
          <div 
            v-for="friend in friendStore.friends" 
            :key="friend.id"
            class="tg-item group"
          >
            <div class="tg-item-avatar" :style="{ background: getAvatarColor(friend.id) }">
              <img v-if="friend.avatar" :src="friend.avatar" class="w-full h-full rounded-full object-cover">
              <span v-else>{{ (friend.display_name || friend.username || "?")[0].toUpperCase() }}</span>
            </div>
            <div class="tg-item-content">
              <div class="tg-item-top">
                <span class="tg-item-title">{{ friend.display_name || friend.username }}</span>
              </div>
              <div class="tg-item-bottom">
                <p class="tg-item-subtitle">@{{ friend.username }}</p>
              </div>
            </div>
            <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <button v-if="confirmingDelete !== friend.id" class="p-2 text-[var(--tg-text-blue)] hover:bg-[var(--tg-bg-hover)] rounded-full transition-colors" @click="openChat(friend.id)">
                <IconSend class="w-5 h-5" />
              </button>
              
              <div v-if="confirmingDelete === friend.id" class="flex items-center bg-red-500 rounded-lg overflow-hidden h-9">
                <button class="px-3 h-full text-white text-[10px] font-bold hover:bg-red-600 transition-colors border-r border-white/20 uppercase tracking-tighter" @click="doDeleteFriend(friend.id)">
                  Confirm
                </button>
                <button class="px-2 h-full text-white text-xs font-bold hover:bg-red-600 transition-colors" @click="confirmingDelete = null">
                  ✕
                </button>
              </div>
              <button v-else class="p-2 text-red-500 hover:bg-red-500/10 rounded-full transition-colors" @click="confirmingDelete = friend.id">
                <IconDelete class="w-5 h-5" />
              </button>
            </div>
          </div>
        </div>

        <!-- Requests Tab -->
        <div v-else key="requests" class="absolute inset-0 overflow-y-auto p-2">
          <div v-if="friendStore.requests.length === 0" class="text-center text-[var(--tg-text-gray)] mt-10 text-sm px-4">No pending requests.</div>
          <div 
            v-for="req in friendStore.requests" 
            :key="req.id"
            class="tg-item cursor-default"
          >
            <div class="tg-item-avatar bg-gray-500">
              <img v-if="req.from_user.avatar" :src="req.from_user.avatar" class="w-full h-full rounded-full object-cover">
              <span v-else>{{ (req.from_user.display_name || req.from_user.username || "?")[0].toUpperCase() }}</span>
            </div>
            <div class="tg-item-content">
              <div class="tg-item-top">
                <span class="tg-item-title">{{ req.from_user.display_name || req.from_user.username }}</span>
              </div>
              <div class="tg-item-bottom">
                <p class="tg-item-subtitle text-[11px]">Sent you a request</p>
              </div>
            </div>
            <div class="flex flex-col gap-1 ml-2">
              <button class="tg-btn-primary px-3 py-1 rounded-lg text-xs" @click="handleRequest(req.id, 'accept')">Accept</button>
              <button class="tg-btn-secondary px-3 py-1 rounded-lg text-xs" @click="handleRequest(req.id, 'reject')">Reject</button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue"
import { useFriendStore } from "../stores/friend"
import { useChatStore } from "../stores/chat"
import IconBack from "./icons/IconBack.vue"
import IconSend from "./icons/IconSend.vue"
import IconDelete from "./icons/IconDelete.vue"

const friendStore = useFriendStore()
const chat = useChatStore()
const activeTab = ref<"friends" | "requests">("friends")
const tabTransition = ref("tab-slide")
const confirmingDelete = ref<number | null>(null)

watch(activeTab, (newVal, oldVal) => {
  if (newVal === "requests" && oldVal === "friends") {
    tabTransition.value = "tab-slide"
  } else {
    tabTransition.value = "tab-slide-reverse"
  }
})

onMounted(() => {
  friendStore.loadFriends()
  friendStore.loadRequests()
})

function getAvatarColor(id: number) {
  const colors = [
    'var(--tg-avatar-1)', 'var(--tg-avatar-2)', 'var(--tg-avatar-3)', 
    'var(--tg-avatar-4)', 'var(--tg-avatar-5)', 'var(--tg-avatar-6)', 'var(--tg-avatar-7)'
  ]
  return colors[id % colors.length]
}

async function handleRequest(id: number, action: "accept" | "reject") {
  try {
    await friendStore.handleRequest(id, action)
    if (action === 'accept') {
      chat.addToast({ title: "已通过", message: "您已成功添加对方为好友", type: "success" })
    }
  } catch (error) {
    chat.addToast({ title: "操作失败", message: error instanceof Error ? error.message : `无法${action}该请求`, type: "error" })
  }
}

async function doDeleteFriend(id: number) {
  try {
    await friendStore.removeFriend(id)
    confirmingDelete.value = null
    chat.addToast({ title: "已删除", message: "好友关系已解除", type: "success" })
  } catch (error) {
    chat.addToast({ title: "删除失败", message: error instanceof Error ? error.message : "无法删除该好友", type: "error" })
  }
}

async function openChat(friendId: number) {
  try {
    await chat.openPrivateByUserID(friendId)
    chat.closeLeftSidebar()
  } catch (error) {
    chat.addToast({ title: "操作失败", message: "无法发起聊天", type: "error" })
  }
}
</script>
