import { defineStore } from "pinia"
import { apiRequest, authHeaders, patchMyProfile } from "../api/http"
import type { LoginResponse, UserProfile } from "../types/chat"

const TOKEN_KEY = "go_chat_token"
const USER_KEY = "go_chat_user"

function loadUser(): UserProfile | null {
  try {
    const raw = localStorage.getItem(USER_KEY)
    return raw ? (JSON.parse(raw) as UserProfile) : null
  } catch {
    return null
  }
}

export const useAuthStore = defineStore("auth", {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || "",
    user: loadUser() as UserProfile | null,
    loading: false,
  }),
  getters: {
    isAuthed: (state): boolean => Boolean(state.token),
  },
  actions: {
    persist() {
      if (this.token) {
        localStorage.setItem(TOKEN_KEY, this.token)
      } else {
        localStorage.removeItem(TOKEN_KEY)
      }

      if (this.user) {
        localStorage.setItem(USER_KEY, JSON.stringify(this.user))
      } else {
        localStorage.removeItem(USER_KEY)
      }
    },
    setAuth(token: string, user: UserProfile) {
      this.token = token
      this.user = user
      this.persist()
    },
    clearAuth() {
      this.token = ""
      this.user = null
      this.persist()
    },
    async register(username: string, password: string) {
      this.loading = true
      try {
        await apiRequest("/api/auth/register", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
        })
      } finally {
        this.loading = false
      }
    },
    async login(username: string, password: string): Promise<LoginResponse> {
      this.loading = true
      try {
        const data = await apiRequest<LoginResponse>("/api/auth/login", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
        })
        this.setAuth(data.token, data.user)
        return data
      } finally {
        this.loading = false
      }
    },
    async fetchMe(): Promise<UserProfile | null> {
      if (!this.token) {
        return null
      }
      const data = await apiRequest<UserProfile>("/api/users/me", {
        headers: authHeaders(this.token),
      })
      this.user = data
      this.persist()
      return data
    },
    async updateMyProfile(payload: { display_name?: string; bio?: string; avatar?: string }) {
      if (!this.token) return
      this.loading = true
      try {
        const updated = await patchMyProfile(this.token, payload)
        this.user = updated
        this.persist()
        return updated
      } finally {
        this.loading = false
      }
    },
  },
})
