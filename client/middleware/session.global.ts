export default defineNuxtRouteMiddleware(async () => {
  if (import.meta.server) { return }
  try {
    const response = await fetch('/api/auth/me', { cache: 'no-store', credentials: 'same-origin' })
    if (response.status === 401) {
      window.location.replace('/login')
      return abortNavigation()
    }
    if (!response.ok) { return abortNavigation('Account is unavailable. Please reload.') }
  } catch { return abortNavigation('Could not connect. Please reload.') }
})
