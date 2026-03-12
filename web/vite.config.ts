import { defineConfig, loadEnv } from "vite"
import vue from "@vitejs/plugin-vue"

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "")
  const target = env.VITE_DEV_PROXY_TARGET || "http://localhost:8080"
  const parsedPort = Number.parseInt(env.VITE_DEV_PORT || "", 10)
  const devPort = Number.isInteger(parsedPort) && parsedPort > 0 ? parsedPort : 4173
  const devHost = env.VITE_DEV_HOST || "127.0.0.1"

  return {
    plugins: [vue()],
    server: {
      host: devHost,
      port: devPort,
      proxy: {
        "/api": {
          target,
          changeOrigin: true,
        },
        "/ws": {
          target,
          ws: true,
          changeOrigin: true,
        },
        "/uploads": {
          target,
          changeOrigin: true,
        },
      },
    },
  }
})
