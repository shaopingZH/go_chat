<template>
  <div class="tg-auth-page">
    <div class="tg-auth-card">
      <div class="tg-auth-logo">
        <IconSend class="w-12 h-12 text-white ml-1" />
      </div>
      <h1 class="tg-auth-title">Go-Chat</h1>
      <p class="tg-auth-subtitle">Welcome back. Please log in to continue.</p>
      
      <form class="tg-auth-form" @submit.prevent="handleLogin">
        <input 
          v-model="username" 
          class="tg-auth-input" 
          type="text" 
          placeholder="Username" 
          required 
          autofocus
        />
        <input 
          v-model="password" 
          class="tg-auth-input" 
          type="password" 
          placeholder="Password" 
          required 
        />
        
        <div v-if="errorMsg" class="text-red-500 text-sm text-center">{{ errorMsg }}</div>
        
        <button type="submit" class="tg-auth-btn" :disabled="loading">
          {{ loading ? 'Wait...' : 'Log In' }}
        </button>
      </form>
      
      <div class="flex gap-2 text-sm mt-2">
        <span class="text-[var(--tg-text-gray)]">No account?</span>
        <a class="tg-auth-link" @click="$router.push('/register')">Sign up</a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useRouter } from "vue-router"
import { useAuthStore } from "../stores/auth"
import IconSend from "../components/icons/IconSend.vue"

const router = useRouter()
const auth = useAuthStore()

const username = ref("")
const password = ref("")
const loading = ref(false)
const errorMsg = ref("")

async function handleLogin() {
  if (!username.value || !password.value) return
  
  loading.value = true
  errorMsg.value = ""
  
  try {
    await auth.login(username.value, password.value)
    router.push("/chat")
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : "Login failed"
  } finally {
    loading.value = false
  }
}
</script>
