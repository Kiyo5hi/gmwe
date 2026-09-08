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
        <button v-if="!pairing" class="btn btn-outline" :disabled="busy" @click="startPairing">
          Connect Shortcut
        </button>
        <template v-else>
          <label for="device-code">Authorization code</label>
          <input id="device-code" class="input input-bordered device-code" :value="pairing.user_code" readonly>
          <div class="account-actions">
            <a class="btn btn-outline" :href="pairing.verification_uri" target="_blank" rel="noopener noreferrer">Authorize in Pocket-ID</a>
            <button class="btn btn-primary" :disabled="busy" @click="finishPairing">
              {{ busy ? 'Checking...' : 'Download connection' }}
            </button>
          </div>
        </template>
      </section>
    </template>
  </main>
</template>

<script setup lang="ts">
type Account = { user: { name: string }, csrf: string }
type Pairing = { verification_uri: string, user_code: string }
const account = ref<Account | null>(null)
const pairing = ref<Pairing | null>(null)
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
  if (status === 401) { account.value = null; pairing.value = null; error.value = 'Your session has ended. Please sign in again.' } else if (status === 403) { error.value = 'This request was not authorized.' } else if (status === 409) { error.value = 'Already exists or is in progress.' } else { error.value = 'Could not complete the request. Please try again.' }
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
    account.value = null; pairing.value = null
  })
}

async function startPairing () {
  await perform(async () => {
    const response = await command('/api/auth/shortcut/start')
    if (!response.ok) { failed(response.status); return }
    pairing.value = await response.json()
  })
}

async function finishPairing () {
  await perform(async () => {
    const response = await command('/api/auth/shortcut/finish')
    if (response.status === 202) { notice.value = 'Waiting for authorization.'; return }
    if (!response.ok) { pairing.value = null; failed(response.status); return }
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url; link.download = 'gmwe-shortcut.json'; link.click()
    setTimeout(() => URL.revokeObjectURL(url), 1000)
    pairing.value = null; notice.value = 'Connection downloaded.'
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
.account-page .btn { white-space: normal; height: auto; min-height: 3rem; padding: 0.65rem 1rem; border-radius: 0.5rem; }
.device-code { font-family: monospace; width: 100%; }
.account-page textarea { width: 100%; resize: vertical; }
</style>
