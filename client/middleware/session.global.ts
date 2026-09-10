export default defineNuxtRouteMiddleware(async () => {
  if (import.meta.server) { return }
  // Caddy has already authenticated the initial document and its assets.
  if (useNuxtApp().isHydrating) { return }
  try {
    const response = await fetch('/api/auth/session', { cache: 'no-store', credentials: 'same-origin' })
    if (response.status === 401) {
      window.location.replace('/login')
      return abortNavigation()
    }
    if (!response.ok) { return abortNavigation('Account is unavailable. Please reload.') }
  } catch { return abortNavigation('Could not connect. Please reload.') }
})
