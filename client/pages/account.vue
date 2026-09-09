<template>
  <section class="account-page">
    <h1>Account</h1>
    <p v-if="loading" role="status">
      加载中…
    </p>
    <p v-if="error" role="alert" class="text-error">
      {{ error }}
    </p>
    <div v-if="account" class="identity">
      <UserRound :size="40" aria-hidden="true" />
      <div><h2>{{ fullName }}</h2><p>Pocket-ID</p></div>
    </div>
    <dl v-if="account" class="profile-fields">
      <div><dt>First name</dt><dd>{{ account.profile?.first_name || (account.profile?.loaded ? '未填写' : '待同步') }}</dd></div>
      <div><dt>Last name</dt><dd>{{ account.profile?.last_name || (account.profile?.loaded ? '未填写' : '待同步') }}</dd></div>
    </dl>
    <p v-if="account?.profile?.state === 'consent_required'" role="status">
      当前登录尚未授权姓名资料。
    </p>
    <p v-if="account?.profile?.state === 'unavailable'" role="status">
      Pocket-ID 资料暂时无法读取。
    </p>
    <a v-if="account && !account.profile?.loaded" href="/api/auth/login" class="btn btn-ghost">同步姓名</a>
    <button v-if="account" class="btn btn-outline" :disabled="busy" @click="logout">
      <LogOut :size="18" aria-hidden="true" />{{ busy ? '退出中…' : '退出登录' }}
    </button>
    <a v-if="expired" href="/api/auth/login" class="btn btn-outline">重新登录</a>
  </section>
</template>
<script setup lang="ts">
import { LogOut, UserRound } from '@lucide/vue'
type Account = { user: { name: string, subject: string }, profile?: { first_name: string, last_name: string, loaded: boolean, state?: string }, csrf: string }
const account = ref<Account | null>(null)
const fullName = computed(() => [account.value?.profile?.first_name, account.value?.profile?.last_name].filter(Boolean).join(' ') || account.value?.user.name)
const loading = ref(true)
const busy = ref(false)
const expired = ref(false)
const error = ref('')
useHead({ title: 'Account | GMWE' })
onMounted(async () => {
  try {
    const response = await fetch('/api/auth/me', { cache: 'no-store' })
    if (response.status === 401) { expired.value = true; return }
    if (!response.ok) { throw new Error('Account unavailable') }
    account.value = await response.json()
  } catch { error.value = '无法加载账户，请刷新重试。' } finally { loading.value = false }
})
async function logout () {
  if (busy.value || !account.value) { return }
  busy.value = true; error.value = ''
  try {
    const response = await fetch('/api/auth/logout', { method: 'POST', cache: 'no-store', headers: { 'X-CSRF-Token': account.value.csrf } })
    if (!response.ok && response.status !== 401) { throw new Error('Logout failed') }
    try { sessionStorage.removeItem('gmwe:draft:' + account.value.user.subject) } catch { /* Storage may be unavailable. */ }
    window.location.replace('/login')
  } catch { error.value = '退出失败，请重试。' } finally { busy.value = false }
}
</script>
<style scoped>
.account-page { max-width: 42rem; margin: 0 auto; }
h1 { font-size: 1.5rem; font-weight: 600; }
h2 { font-size: 1.25rem; overflow-wrap: anywhere; }
.identity { display: flex; align-items: center; gap: 1rem; margin: 2rem 0; }
.identity > div { min-width: 0; }
.identity > svg { flex-shrink: 0; }
.profile-fields > div { min-width: 0; }
.identity p { opacity: 0.65; margin-top: 0.25rem; }
.profile-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; margin-bottom: 2rem; }
dt { font-size: 0.875rem; opacity: 0.65; }
dd { margin-top: 0.25rem; overflow-wrap: anywhere; }
.btn { border-radius: 6px; gap: 0.5rem; }
</style>
