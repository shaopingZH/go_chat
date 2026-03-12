import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router"
import { useAuthStore } from "../stores/auth"
import LoginView from "../views/LoginView.vue"
import RegisterView from "../views/RegisterView.vue"
import ChatView from "../views/ChatView.vue"

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/chat",
  },
  {
    path: "/login",
    name: "login",
    component: LoginView,
    meta: { guestOnly: true },
  },
  {
    path: "/register",
    name: "register",
    component: RegisterView,
    meta: { guestOnly: true },
  },
  {
    path: "/chat",
    name: "chat",
    component: ChatView,
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthed) {
    return { name: "login" }
  }

  if (to.meta.guestOnly && auth.isAuthed) {
    return { name: "chat" }
  }

  if (to.meta.requiresAuth && auth.isAuthed && !auth.user) {
    try {
      await auth.fetchMe()
    } catch {
      auth.clearAuth()
      return { name: "login" }
    }
  }

  return true
})

export default router
