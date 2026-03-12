<template>
  <div class="tg-auth-page">
    <div class="tg-auth-card">
      <div class="tg-auth-logo">
        <IconSend class="w-12 h-12 text-white ml-1" />
      </div>
      <h1 class="tg-auth-title">Sign Up</h1>
      <p class="tg-auth-subtitle">Create a new account to chat with friends.</p>
      
      <form class="tg-auth-form" @submit.prevent="handleRegister">
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
          minlength="6"
        />
        
        <div v-if="errorMsg" class="text-red-500 text-sm text-center">{{ errorMsg }}</div>
        
        <button type="submit" class="tg-auth-btn" :disabled="loading">
          {{ loading ? 'Creating...' : 'Sign Up' }}
        </button>
      </form>
      
      <div class="flex gap-2 text-sm mt-2">
        <span class="text-[var(--tg-text-gray)]">Have an account?</span>
        <a class="tg-auth-link" @click="$router.push('/login')">Log in</a>
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

async function handleRegister() {
  if (!username.value || !password.value) return
  
  loading.value = true
  errorMsg.value = ""
  
  try {
    await auth.register(username.value, password.value)
    // Auto login or redirect to login? 
    // Usually auto login is better, but let's redirect to login for clarity or strict flow
    // But auth store register might not set token. Let's check auth store.
    // Assuming register just registers.
    await auth.login(username.value, password.value)
    router.push("/chat")
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : "Registration failed"
  } finally {
    loading.value = false
  }
}
</script>
