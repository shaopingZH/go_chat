<template>
  <div class="h-full flex flex-col bg-[var(--tg-bg-sidebar)] border-r border-[var(--tg-border)] shadow-xl relative w-full overflow-hidden">
    
    <!-- MAIN VIEW LAYER -->
    <Transition name="slide-view">
      <div v-if="!isEditing" class="absolute inset-0 flex flex-col z-10 bg-[var(--tg-bg-sidebar)]">
        <!-- View Header -->
        <div class="h-14 flex items-center px-4 bg-[var(--tg-bg-header)] border-b border-[var(--tg-border)] gap-4 flex-shrink-0 relative">
          <button class="tg-icon-btn -ml-2" @click="chat.closeLeftSidebar()">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="19" y1="12" x2="5" y2="12"></line>
              <polyline points="12 19 5 12 12 5"></polyline>
            </svg>
          </button>
          <h2 class="text-lg font-bold text-[var(--tg-text-white)] flex-1">My Profile</h2>
          
          <div class="flex items-center gap-1">
            <button class="tg-btn-ghost text-sm text-[var(--tg-text-blue)] font-bold px-3 py-1.5 rounded-lg hover:bg-[var(--tg-text-blue)]/10 transition-colors" @click="enterEditMode">
              Edit
            </button>
            <!-- More Options -->
            <div class="relative">
              <button class="tg-icon-btn -mr-2" @click.stop="showMoreMenu = !showMoreMenu">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="1"></circle>
                  <circle cx="12" cy="5" r="1"></circle>
                  <circle cx="12" cy="19" r="1"></circle>
                </svg>
              </button>
              <Transition name="fade">
                <div v-if="showMoreMenu">
                  <div class="fixed inset-0 z-[100]" @click="showMoreMenu = false"></div>
                  <div class="absolute right-0 top-full mt-1 bg-[var(--tg-bg-menu)] shadow-[0_4px_24px_rgba(0,0,0,0.15)] rounded-[14px] p-1 w-40 flex flex-col z-[101] tg-menu-pop" style="--menu-transform-origin: top right" @click.stop>
                    <button class="flex items-center px-3 py-2 text-[14px] text-red-500 hover:bg-[var(--tg-bg-hover)] rounded-lg transition-colors w-full text-left font-medium" @click="showLogoutConfirm = true; showMoreMenu = false">
                      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-3">
                        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
                        <polyline points="16 17 21 12 16 7"></polyline>
                        <line x1="21" y1="12" x2="9" y2="12"></line>
                      </svg>
                      Logout
                    </button>
                  </div>
                </div>
              </Transition>
            </div>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-6 flex flex-col gap-6 custom-scrollbar">
          <!-- Avatar & Name Info -->
          <div class="flex flex-col items-center gap-3">
            <div class="w-24 h-24 rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-4xl font-bold shadow-lg shadow-purple-500/20 overflow-hidden border-2 border-[var(--tg-border)]">
              <img v-if="form.avatar" :src="form.avatar" class="w-full h-full object-cover">
              <span v-else>{{ (form.display_name || auth.user?.username || "?")[0]?.toUpperCase() }}</span>
            </div>
            <div class="text-center">
              <h3 class="text-xl font-bold text-[var(--tg-text-white)]">{{ form.display_name || auth.user?.username }}</h3>
              <p class="text-sm text-[var(--tg-text-gray)]">@{{ auth.user?.username }}</p>
            </div>
          </div>

          <!-- Bio Section -->
          <div v-if="form.bio" class="bg-[var(--tg-bg-chat)] rounded-2xl p-4 border border-[var(--tg-border)] flex flex-col gap-1">
            <span class="text-[13px] font-bold text-[var(--tg-text-blue)] uppercase tracking-wide">Bio</span>
            <p class="text-[var(--tg-text-white)] whitespace-pre-wrap leading-relaxed">{{ form.bio }}</p>
          </div>
          
          <!-- User Details Card -->
          <div class="bg-[var(--tg-bg-chat)] rounded-2xl p-4 border border-[var(--tg-border)] flex flex-col gap-4">
            <div class="flex flex-col">
              <span class="text-[13px] font-bold text-[var(--tg-text-blue)] uppercase tracking-wide">User ID</span>
              <p class="text-[var(--tg-text-white)] font-mono mt-1 text-sm">{{ auth.user?.id }}</p>
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
          <h2 class="text-lg font-bold text-[var(--tg-text-white)] flex-1">Edit Profile</h2>
          <button class="tg-btn-ghost text-sm text-[var(--tg-text-blue)] font-bold px-3 py-1.5 rounded-lg hover:bg-[var(--tg-text-blue)]/10 transition-colors" :disabled="loading" @click="save">
            {{ loading ? '...' : 'Done' }}
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6 flex flex-col gap-6 custom-scrollbar">
          <!-- Avatar Edit -->
          <div class="flex flex-col items-center gap-4">
            <div class="relative group cursor-pointer w-32 h-32" @click="triggerAvatarUpload">
              <div class="w-full h-full rounded-full bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white text-5xl font-bold shadow-xl overflow-hidden border-2 border-[var(--tg-border)] transition-transform group-hover:scale-[1.02] active:scale-95">
                <img v-if="editForm.avatar" :src="editForm.avatar" class="w-full h-full object-cover">
                <span v-else>{{ (editForm.display_name || auth.user?.username || "?")[0]?.toUpperCase() }}</span>
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
            <label class="text-[13px] text-[var(--tg-text-blue)] font-bold uppercase tracking-wide ml-1">Display Name</label>
            <input
              v-model="editForm.display_name"
              type="text"
              class="w-full bg-[var(--tg-input-bg)] text-[var(--tg-text-white)] px-4 py-3.5 rounded-xl border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] outline-none transition-all shadow-sm focus:shadow-[0_0_0_2px_rgba(51,144,236,0.1)]"
              placeholder="Set a nickname"
              maxlength="50"
            >
          </div>

          <!-- Bio Edit -->
          <div class="flex flex-col gap-1.5">
            <label class="text-[13px] text-[var(--tg-text-blue)] font-bold uppercase tracking-wide ml-1">Bio</label>
            <textarea
              v-model="editForm.bio"
              rows="3"
              class="w-full bg-[var(--tg-input-bg)] text-[var(--tg-text-white)] px-4 py-3.5 rounded-xl border border-[var(--tg-border)] focus:border-[var(--tg-text-blue)] outline-none transition-all resize-none shadow-sm focus:shadow-[0_0_0_2px_rgba(51,144,236,0.1)]"
              placeholder="Tell us about yourself..."
              maxlength="200"
            ></textarea>
            <p class="text-[11px] text-[var(--tg-text-gray)] ml-1 leading-normal">Any details such as age, occupation or city. Example: 23 y.o. designer from San Francisco.</p>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Global Logout Confirmation Modal -->
    <Transition name="fade">
      <div v-if="showLogoutConfirm" class="fixed inset-0 z-[1000] flex items-center justify-center bg-black/20 backdrop-blur-sm px-4" @click="showLogoutConfirm = false">
        <div class="bg-[var(--tg-bg-sidebar)] border border-[var(--tg-border)] shadow-2xl rounded-2xl p-6 w-full max-w-[320px] tg-menu-pop" @click.stop>
          <h3 class="text-lg font-bold text-[var(--tg-text-white)] mb-2">Logout</h3>
          <p class="text-[var(--tg-text-gray)] text-sm mb-6">Are you sure you want to log out?</p>
          <div class="flex gap-3">
            <button class="tg-btn-danger flex-1" @click="confirmLogout">Logout</button>
            <button class="tg-btn-secondary flex-1" @click="showLogoutConfirm = false">Cancel</button>
          </div>
        </div>
      </div>
    </Transition>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { uploadImage } from '../api/http'
import { useChatStore } from '../stores/chat'

const router = useRouter()
const auth = useAuthStore()
const chat = useChatStore()
const fileInput = ref<HTMLInputElement | null>(null)
const loading = ref(false)
const showLogoutConfirm = ref(false)
const showMoreMenu = ref(false)
const isEditing = ref(false)

const form = ref({
  display_name: '',
  bio: '',
  avatar: ''
})

const editForm = ref({
  display_name: '',
  bio: '',
  avatar: ''
})

onMounted(() => {
  if (auth.user) {
    form.value.display_name = auth.user.display_name || auth.user.username
    form.value.bio = auth.user.bio || ''
    form.value.avatar = auth.user.avatar || ''
  }
})

function enterEditMode() {
  editForm.value = { ...form.value }
  isEditing.value = true
}

function exitEditMode() {
  if (loading.value) return
  isEditing.value = false
}

function triggerAvatarUpload() {
  fileInput.value?.click()
}

async function handleFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file || !auth.token) return

  try {
    loading.value = true
    const res = await uploadImage(file, auth.token)
    editForm.value.avatar = res.url
  } catch (error) {
    chat.addToast({ title: "上传失败", message: error instanceof Error ? error.message : "图片上传失败", type: "error" })
  } finally {
    loading.value = false
  }
}

async function save() {
  loading.value = true
  try {
    await auth.updateMyProfile({
      display_name: editForm.value.display_name,
      bio: editForm.value.bio,
      avatar: editForm.value.avatar
    })
    // Update local view state
    form.value = { ...editForm.value }
    isEditing.value = false
    chat.addToast({ title: "保存成功", message: "个人资料已更新", type: "success" })
  } catch (error) {
    chat.addToast({ title: "保存失败", message: error instanceof Error ? error.message : "无法更新个人资料", type: "error" })
  } finally {
    loading.value = false
  }
}

function confirmLogout() {
  auth.clearAuth()
  chat.resetState()
  router.push('/login')
}
</script>

<style scoped>
/* View Layer Animation */
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

/* Edit Layer Animation */
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
