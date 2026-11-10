import { ref } from 'vue'
import { createRouter, createWebHistory, START_LOCATION } from 'vue-router'
import { routes } from './routes'

export const navigationError = ref('')
export const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
})

router.beforeEach(async (_to, from) => {
  navigationError.value = ''
  // Caddy authenticates every initial document and all private static assets.
  if (from === START_LOCATION) { return true }
  try {
    const response = await fetch('/api/auth/session', { cache: 'no-store', credentials: 'same-origin' })
    if (response.status === 401) {
      window.location.replace('/login')
      return false
    }
    if (!response.ok) {
      navigationError.value = '账户暂时不可用，请稍后重试。'
      return false
    }
    return true
  } catch {
    navigationError.value = '无法连接，请检查网络后重试。'
    return false
  }
})

router.afterEach((to, _from, failure) => {
  if (!failure) { document.title = typeof to.meta.title === 'string' ? to.meta.title : 'GMWE' }
})
