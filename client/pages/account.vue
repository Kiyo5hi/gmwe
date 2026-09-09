<template>
  <main class="account-page">
    <header class="account-heading">
      <img src="/gmwe.webp" alt="GMWE" width="64" height="64">
      <div>
        <h1>Account</h1><p v-if="account">
          {{ account.user.name }}
        </p>
      </div>
      <button v-if="account" class="btn btn-ghost" :disabled="busy" @click="logout">
        Sign out
      </button>
    </header>
    <p v-if="loading" role="status">
      Loading...
    </p>
    <p v-if="error" class="text-error" role="alert">
      {{ error }}
    </p>
    <p v-if="notice" class="text-success" role="status">
      {{ notice }}
    </p>
    <a v-if="!loading && !account" class="btn btn-primary" href="/api/auth/login">Sign in with Pocket-ID</a>
    <template v-if="account">
      <form class="account-section" @submit.prevent="saveEntry">
        <h2>New entry</h2>
        <label for="entry">Content</label>
        <textarea
          id="entry"
          v-model="content"
          class="textarea textarea-bordered"
          rows="5"
          maxlength="2000"
          required
          :disabled="busy"
        />
        <div class="account-actions">
          <span>{{ content.length }} / 2000</span><button class="btn btn-primary" :disabled="busy || !content.trim()">
            Add entry
          </button>
        </div>
      </form>
      <section class="account-section">
        <h2>Apple Shortcuts</h2>
        <div class="shortcut-actions">
          <a class="btn btn-outline" href="/downloads/GMWE-iPhone.shortcut" download="GMWE-iPhone.shortcut">Download Shortcut</a>
          <button class="btn btn-outline" :disabled="busy" @click="startPairing">
            Download Connection
          </button>
        </div>
      </section>
    </template>
  </main>
</template>

<script setup lang="ts">
type Account = { user: { name: string }, csrf: string }
const account = ref<Account | null>(null)
const content = ref('')
const error = ref('')
const notice = ref('')
const loading = ref(true)
const busy = ref(false)
useHead({ title: 'Account | GMWE', meta: [{ name: 'referrer', content: 'no-referrer' }] })

onMounted(async () => {
  try {
    const response = await fetch('/api/auth/me', { cache: 'no-store', credentials: 'same-origin' })
    if (response.ok) { account.value = await response.json() } else if (response.status !== 401) { error.value = 'Account is unavailable. Please try again.' }
  } catch { error.value = 'Could not connect. Please try again.' } finally { loading.value = false }
})

async function command (path: string, body?: object) {
  return await fetch(path, { method: 'POST', credentials: 'same-origin', cache: 'no-store', headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': account.value?.csrf || '' }, body: body ? JSON.stringify(body) : undefined })
}

async function perform (action: () => Promise<void>) {
  if (busy.value) { return }
  busy.value = true; error.value = ''; notice.value = ''
  try { await action() } catch { error.value = 'Could not complete the request. Please try again.' } finally { busy.value = false }
}

function failed (status: number) {
  if (status === 401) { account.value = null; error.value = 'Your session has ended. Please sign in again.' } else if (status === 403) { error.value = 'This request was not authorized.' } else if (status === 409) { error.value = 'Already exists or is in progress.' } else { error.value = 'Could not complete the request. Please try again.' }
}

async function saveEntry () {
  await perform(async () => {
    const response = await command('/api/v1/hitokoto', { Content: content.value })
    if (!response.ok) { failed(response.status); return }
    content.value = ''; notice.value = 'Entry added.'
  })
}

async function logout () {
  await perform(async () => {
    const response = await command('/api/auth/logout')
    if (!response.ok) { failed(response.status); return }
    account.value = null
  })
}

async function startPairing () {
  await perform(async () => {
    const response = await command('/api/auth/shortcut/start')
    if (!response.ok) { failed(response.status); return }
    const result = await response.json()
    const target = new URL(result.authorization_url)
    if (target.protocol !== 'https:') { throw new Error('Invalid authorization URL') }
    window.location.assign(target.href)
  })
}
</script>

<style scoped>
.account-page { width: 100%; max-width: 42rem; margin: 0 auto; padding: 1.5rem; letter-spacing: 0; }
.account-heading { display: flex; align-items: center; gap: 1rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
.account-heading > div { flex: 1; }
h1 { font-size: 1.5rem; font-weight: 600; }
h2 { font-size: 1.125rem; font-weight: 600; }
.account-section { margin-top: 2rem; padding-top: 1.5rem; border-top: 1px solid currentColor; display: flex; flex-direction: column; gap: 0.75rem; }
.account-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.shortcut-actions { display: flex; flex-wrap: wrap; gap: 0.75rem; }
.shortcut-actions > * { flex: 1 1 13rem; }
.account-page .btn { white-space: normal; height: auto; min-height: 3rem; padding: 0.65rem 1rem; border-radius: 0.5rem; }
.account-page textarea { width: 100%; resize: vertical; }
</style>
