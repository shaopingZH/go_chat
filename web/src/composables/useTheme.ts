import { ref, nextTick } from 'vue'

const isDark = ref(true)

export function useTheme() {
  async function toggleTheme(event?: MouseEvent) {
    if (!document.startViewTransition) {
      isDark.value = !isDark.value
      applyTheme()
      return
    }

    const x = event?.clientX ?? window.innerWidth / 2
    const y = event?.clientY ?? window.innerHeight / 2
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y)
    )

    const transition = document.startViewTransition(async () => {
      document.documentElement.classList.add('theme-transitioning')
      isDark.value = !isDark.value
      applyTheme()
      await nextTick()
    })

    transition.finished.finally(() => {
      document.documentElement.classList.remove('theme-transitioning')
    })

    await transition.ready

    const isNowDark = isDark.value
    
    if (isNowDark) {
      // White -> Black: New layer (dark) expands on top
      document.documentElement.animate(
        {
          clipPath: [
            `circle(0px at ${x}px ${y}px)`,
            `circle(${endRadius}px at ${x}px ${y}px)`
          ],
        },
        {
          duration: 450,
          easing: 'ease-in-out',
          pseudoElement: '::view-transition-new(root)',
          fill: 'forwards',
        }
      )
    } else {
      // Black -> White: Old layer (dark) shrinks on top
      document.documentElement.animate(
        {
          clipPath: [
            `circle(${endRadius}px at ${x}px ${y}px)`,
            `circle(0px at ${x}px ${y}px)`
          ],
        },
        {
          duration: 450,
          easing: 'ease-in-out',
          pseudoElement: '::view-transition-old(root)',
          fill: 'forwards',
        }
      )
    }
  }

  function applyTheme() {
    if (isDark.value) {
      document.documentElement.removeAttribute('data-theme')
      localStorage.setItem('tg-theme', 'dark')
    } else {
      document.documentElement.setAttribute('data-theme', 'light')
      localStorage.setItem('tg-theme', 'light')
    }
  }

  const saved = localStorage.getItem('tg-theme')
  if (saved === 'light') {
    isDark.value = false
    applyTheme()
  } else {
    isDark.value = true
    applyTheme()
  }

  return { isDark, toggleTheme }
}

declare global {
  interface Document {
    startViewTransition(callback: () => void): {
      ready: Promise<void>
      finished: Promise<void>
      updateCallbackDone: Promise<void>
    }
  }
}
