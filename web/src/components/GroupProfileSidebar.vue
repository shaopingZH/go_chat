<template>
  <div class="h-full flex flex-col bg-[var(--tg-bg-sidebar)] border-l border-[var(--tg-border)] shadow-xl relative w-full overflow-hidden">
    
    <!-- MAIN VIEW LAYER -->
    <Transition name="slide-view">
      <div v-if="!isEditing" class="absolute inset-0 flex flex-col z-10 bg-[var(--tg-bg-sidebar)]">
        <!-- View Header -->
        <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0 justify-between">
          <div class="flex items-center gap-4">
            <button class="tg-icon-btn -ml-2" @click="chat.closeRightSidebar()">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
            <h2 class="text-lg font-bold text-[var(--tg-text-white)] truncate max-w-[160px]">Group Info</h2>
          </div>
          <button v-if="isOwner" class="tg-btn-ghost text-sm text-[var(--tg-text-blue)] font-bold px-3 py-1.5 rounded-lg hover:bg-[var(--tg-text-blue)]/10 transition-colors" @click="enterEditMode">
            Edit
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6 flex flex-col gap-6 custom-scrollbar">
          <!-- Group Basic Info -->
          <div class="flex flex-col items-center gap-3">
            <div class="w-24 h-24 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-4xl font-bold shadow-lg shadow-purple-500/20 overflow-hidden border-2 border-[var(--tg-border)]">
              <img v-if="groupAvatar" :src="groupAvatar" class="w-full h-full object-cover">
              <span v-else>{{ groupName[0]?.toUpperCase() }}</span>
            </div>
            <div class="text-center">
              <h3 class="text-xl font-bold text-[var(--tg-text-white)]">{{ groupName }}</h3>
              <p class="text-sm text-[var(--tg-text-gray)]">{{ members.length }} members</p>
            </div>
          </div>

          <!-- Group Details Card -->
          <div class="bg-[var(--tg-bg-chat)] rounded-2xl p-4 border border-[var(--tg-border)] flex flex-col gap-4">
            <div class="flex flex-col">
              <span class="text-[13px] font-medium text-[var(--tg-text-blue)] uppercase tracking-wider">Group ID</span>
              <div class="flex items-center justify-between mt-1">
                <code class="text-[var(--tg-text-white)] font-mono text-sm">{{ groupId }}</code>
                <button class="text-xs text-[var(--tg-text-blue)] font-bold hover:underline" @click="copyGroupId">Copy</button>
              </div>
            </div>
          </div>

          <!-- Members List -->
          <div class="flex flex-col gap-3">
            <h4 class="text-[13px] font-bold text-[var(--tg-text-gray)] uppercase px-1">Members</h4>
            <div v-for="member in members" :key="member.id" class="flex items-center gap-3 p-2 rounded-xl hover:bg-[var(--tg-bg-hover)] transition-colors group">
              <div class="w-10 h-10 rounded-full bg-[var(--tg-bg-active)] flex items-center justify-center text-white font-bold flex-shrink-0 text-sm overflow-hidden border border-[var(--tg-border)]">
                <img v-if="member.avatar" :src="member.avatar" class="w-full h-full object-cover">
                <span v-else>{{ (member.display_name || member.username || "?")[0].toUpperCase() }}</span>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-1">
                  <span class="text-[var(--tg-text-white)] font-bold truncate text-sm">
                    {{ member.display_name || member.username }}
                  </span>
                  <span v-if="member.id === ownerId" class="text-[10px] bg-[var(--tg-text-blue)]/20 text-[var(--tg-text-blue)] px-1.5 py-0.5 rounded font-bold uppercase">Owner</span>
                </div>
                <p class="text-xs text-[var(--tg-text-gray)] truncate">@{{ member.username }}</p>
              </div>
              
              <!-- Kick Action -->
              <div v-if="isOwner && member.id !== selfId" class="ml-2 flex items-center opacity-0 group-hover:opacity-100 transition-opacity">
                <div v-if="confirmingKick === member.id" class="flex items-center bg-red-500 rounded-lg overflow-hidden h-9">
                  <button class="px-3 h-full text-white text-[10px] font-bold hover:bg-red-600 transition-colors border-r border-white/20 uppercase tracking-tighter" @click="doKickMember(member.id)">
                    Confirm
                  </button>
                  <button class="px-2 h-full text-white text-xs font-bold hover:bg-red-600 transition-colors" @click="confirmingKick = null">
                    ✕
                  </button>
                </div>
                <button 
                  v-else
                  class="p-2 text-red-500 hover:bg-red-500/10 rounded-full transition-colors"
                  @click="confirmingKick = member.id"
                >
                  <IconDelete class="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="mt-auto pt-6 border-t border-[var(--tg-border)]">
            <div v-if="memberIds.includes(selfId) && selfId !== ownerId" class="relative overflow-hidden min-h-[48px]">
              <Transition name="fade" mode="out-in">
                <button 
                  v-if="!showLeaveConfirm"
                  class="tg-btn-danger-ghost w-full"
                  @click="showLeaveConfirm = true"
                >
                  Leave Group
                </button>
                <div v-else class="flex gap-2">
                  <button 
                    class="tg-btn-danger flex-1 text-sm uppercase tracking-tight px-0"
                    @click="doLeaveGroup"
                  >
                    Confirm Leave
                  </button>
                  <button 
                    class="tg-btn-secondary px-4 text-sm"
                    @click="showLeaveConfirm = false"
                  >
                    Cancel
                  </button>
                </div>
              </Transition>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- EDIT MODE LAYER -->
    <Transition name="slide-edit">
      <div v-if="isEditing" class="absolute inset-0 flex flex-col z-20 bg-[var(--tg-bg-sidebar)]">
        <!-- Edit Header -->
        <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0">
          <button class="tg-icon-btn -ml-2" @click="exitEditMode">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="19" y1="12" x2="5" y2="12"></line>
              <polyline points="12 19 5 12 12 5"></polyline>
            </svg>
          </button>
          <h2 class="text-lg font-bold text-[var(--tg-text-white)] flex-1">Edit Group</h2>
          <button 
            class="tg-btn-ghost text-sm text-[var(--tg-text-blue)] font-bold px-3 py-1.5 rounded-lg hover:bg-[var(--tg-text-blue)]/10 transition-colors"
            :disabled="isSaving"
            @click="saveGroupInfo"
          >
            {{ isSaving ? '...' : 'Done' }}
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6 flex flex-col gap-6 custom-scrollbar">
          <!-- Avatar Edit -->
          <div class="flex flex-col items-center gap-4">
            <div class="relative group cursor-pointer w-32 h-32" @click="triggerAvatarUpload">
              <div class="w-full h-full rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-5xl font-bold shadow-xl overflow-hidden border-2 border-[var(--tg-border)] transition-transform group-hover:scale-[1.02] active:scale-95">
                <img v-if="editForm.avatar" :src="editForm.avatar" class="w-full h-full object-cover">
                <span v-else>{{ (editForm.name || "G")[0]?.toUpperCase() }}</span>
              </div>
              <div class="absolute inset-0 bg-black/40 rounded-full flex flex-col items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mb-1">
                  <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path>
                  <circle cx="12" cy="13" r="4"></circle>
                </svg>
                <span class="text-white text-xs font-bold uppercase tracking-tight">Change</span>
              </div>
            </div>
            <input ref="fileInput" type="file" class="hidden" accept="image/*" @change="handleFileChange">
          </div>
          
          <!-- Name Edit -->
          <div class="flex flex-col gap-1.5">
            <label class="text-[13px] text-[var(--tg-text-blue)] font-bold uppercase tracking-wide ml-1">Group Name</label>
            <input
              v-model="editForm.name"
              type="text"
              class="w-full bg-[var(--tg-input-bg)] text-[var(--tg-text-white)] px-4 py-3.5 rounded-xl border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] outline-none transition-all shadow-sm focus:shadow-[0_0_0_2px_rgba(51,144,236,0.1)]"
              placeholder="Enter group name"
              maxlength="100"
              @keyup.enter="saveGroupInfo"
            >
            <p class="text-[11px] text-[var(--tg-text-gray)] ml-1">You can provide a group name and optional avatar.</p>
          </div>
        </div>
      </div>
    </Transition>

  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useChatStore } from "../stores/chat"
import { useAuthStore } from "../stores/auth"
import { uploadImage } from "../api/http"
import IconDelete from "./icons/IconDelete.vue"

const auth = useAuthStore()
const chat = useChatStore()
const selfId = computed(() => auth.user?.id || 0)
const showLeaveConfirm = ref(false)
const confirmingKick = ref<number | null>(null)

const isEditing = ref(false)
const isSaving = ref(false)
const editForm = ref({ name: "", avatar: "" })
const fileInput = ref<HTMLInputElement | null>(null)

const groupId = computed(() => chat.rightSidebarPayload as number)
const group = computed(() => chat.conversations.find(c => c.chatType === 'group' && c.targetId === groupId.value))
const groupName = computed(() => group.value?.title || "Group")
const groupAvatar = computed(() => group.value?.avatar || "")
const ownerId = computed(() => (group.value as any)?.owner_id || 0)

const members = computed(() => chat.groupMembersCache[groupId.value] || [])
const memberIds = computed(() => members.value.map(m => m.user_id || m.id))

const isOwner = computed(() => {
  return selfId.value === ownerId.value
})

watch(groupId, async (newId) => {
  if (!newId || !auth.token) {
    return
  }
  showLeaveConfirm.value = false
  confirmingKick.value = null
  isEditing.value = false

  try {
    await chat.getGroupMembers(newId)
  } catch (error) {
    console.error("Failed to load group members", error)
  }
}, { immediate: true })

async function doLeaveGroup() {
  try {
    await chat.leaveCurrentGroup()
    chat.closeRightSidebar()
    chat.addToast({ title: "已退出", message: "您已退出群组", type: "info" })
  } catch (error) {
    chat.addToast({ title: "退出失败", message: error instanceof Error ? error.message : "无法退出群组", type: "error" })
  }
}

async function doKickMember(userId: number) {
  try {
    await chat.kickMember(groupId.value, userId)
    confirmingKick.value = null
  } catch (error) {
    chat.addToast({ title: "移除失败", message: error instanceof Error ? error.message : "无法移除成员", type: "error" })
  }
}

async function copyGroupId() {
  try {
    await navigator.clipboard.writeText(String(groupId.value))
    chat.addToast({ title: "已复制", message: "群组ID已复制到剪贴板！", type: "success" })
  } catch (err) {
    console.error('Failed to copy!', err)
  }
}

function enterEditMode() {
  editForm.value = {
    name: groupName.value,
    avatar: groupAvatar.value
  }
  isEditing.value = true
}

function exitEditMode() {
  if (isSaving.value) return
  isEditing.value = false
}

function triggerAvatarUpload() {
  fileInput.value?.click()
}

async function handleFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  if (file.size > 5 * 1024 * 1024) {
    chat.addToast({ title: "图片过大", message: "群头像不能超过 5MB", type: "warning" })
    target.value = ''
    return
  }

  try {
    isSaving.value = true
    const res = await uploadImage(file, auth.token)
    editForm.value.avatar = res.url
  } catch (error) {
    chat.addToast({ title: "上传失败", message: error instanceof Error ? error.message : "图片上传失败", type: "error" })
  } finally {
    isSaving.value = false
    target.value = ''
  }
}

async function saveGroupInfo() {
  if (!editForm.value.name.trim()) {
    chat.addToast({ title: "无效的群名", message: "群名称不能为空", type: "warning" })
    return
  }
  
  try {
    isSaving.value = true
    await chat.updateGroupInfo(groupId.value, {
      name: editForm.value.name.trim(),
      avatar: editForm.value.avatar
    })
    isEditing.value = false
    chat.addToast({ title: "保存成功", message: "群资料已更新", type: "success" })
  } catch (error) {
    chat.addToast({ title: "保存失败", message: error instanceof Error ? error.message : "无法更新群资料", type: "error" })
  } finally {
    isSaving.value = false
  }
}
</script>

<style scoped>
/* View Layer Animation: Slides slightly to left when edit layer enters */
.slide-view-enter-active, .slide-view-leave-active {
  transition: transform 0.3s cubic-bezier(0.33, 1, 0.68, 1), opacity 0.3s ease;
}
.slide-view-enter-from {
  transform: translateX(-20%);
  opacity: 0.8;
}
.slide-view-leave-to {
  transform: translateX(-20%);
  opacity: 0.8;
}

/* Edit Layer Animation: Slides in from right */
.slide-edit-enter-active, .slide-edit-leave-active {
  transition: transform 0.3s cubic-bezier(0.33, 1, 0.68, 1);
}
.slide-edit-enter-from {
  transform: translateX(100%);
}
.slide-edit-leave-to {
  transform: translateX(100%);
}

.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--tg-border);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--tg-text-gray);
}
</style>
