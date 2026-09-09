<template>
  <main class="account-page">
    <header class="account-heading">
      <img src="/gmwe.webp" alt="GMWE" width="56" height="56">
      <div>
        <h1>New entry</h1><p v-if="account">
          {{ account.user.name }}
        </p>
      </div>
      <button class="btn btn-ghost" :disabled="busy || !account" @click="logout">
        Sign out
      </button>
    </header>
    <p v-if="loading" role="status">
      Loading...
    </p>
    <p v-if="error" class="text-error" role="alert">
      {{ error }}
    </p>
    <a v-if="expired" class="btn btn-outline" href="/api/auth/login">Sign in again</a>
    <form class="entry-form" @submit.prevent="saveEntry">
      <label for="entry">Content</label>
      <textarea
        id="entry"
        ref="input"
        v-model="content"
        class="textarea textarea-bordered"
        rows="7"
        maxlength="4000"
        required
        :disabled="busy || loading || expired || !account"
        aria-describedby="entry-count"
      />
      <div class="account-actions">
        <span id="entry-count" :class="{ 'text-error': count > 2000 }">{{ count }} / 2000</span>
        <button class="btn btn-primary" :disabled="busy || loading || expired || !account || !content.trim() || count > 2000">
          {{ busy ? 'Saving...' : 'Add entry' }}
        </button>
      </div>
    </form>
    <section v-if="saved" class="saved-entry" role="status" aria-live="polite">
      <h2>Entry added</h2>
      <blockquote>{{ saved.Content }}</blockquote>
      <p>{{ account?.user.name }} <span class="entry-id">#{{ saved.ID }}</span></p>
      <a href="/">Back to home</a>
    </section>
  </main>
</template>

<script setup lang="ts">
type Account = { user: { name: string, subject: string }, csrf: string }
type Entry = { ID: number, Content: string }
const account = ref<Account | null>(null)
const content = ref('')
const input = ref<HTMLTextAreaElement | null>(null)
const saved = ref<Entry | null>(null)
const error = ref('')
const loading = ref(true)
const busy = ref(false)
const expired = ref(false)
const count = computed(() => Array.from(content.value).length)
const draftKey = computed(() => account.value ? 'gmwe:draft:' + account.value.user.subject : '')
useHead({ title: 'New entry | GMWE', meta: [{ name: 'referrer', content: 'no-referrer' }] })

watch(content, (value) => {
  if (!draftKey.value) { return }
  try {
    if (value) { sessionStorage.setItem(draftKey.value, JSON.stringify({ text: value, time: Date.now() })) } else { sessionStorage.removeItem(draftKey.value) }
  } catch { /* Private browsing may disable storage; retain the in-memory input. */ }
})

onMounted(async () => {
  try {
    const response = await fetch('/api/auth/me', { cache: 'no-store', credentials: 'same-origin' })
    if (response.status === 401) { expired.value = true; return }
    if (!response.ok) { throw new Error('Account unavailable') }
    account.value = await response.json()
    try {
      const draft = JSON.parse(sessionStorage.getItem(draftKey.value) || 'null')
      if (draft && typeof draft.text === 'string' && draft.text.length <= 4000 && Date.now() - draft.time < 86400000) { content.value = draft.text }
    } catch { /* An unavailable or malformed draft must not block the form. */ }
  } catch { error.value = 'Could not load your account. Reload to try again.' } finally { loading.value = false }
})

async function command (path: string, body?: object) {
  return await fetch(path, { method: 'POST', credentials: 'same-origin', cache: 'no-store', headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': account.value?.csrf || '' }, body: body ? JSON.stringify(body) : undefined })
}

async function saveEntry () {
  if (busy.value || !account.value || !content.value.trim() || count.value > 2000) { return }
  busy.value = true; error.value = ''; saved.value = null
  try {
    const response = await command('/api/v1/hitokoto', { Content: content.value.trim() })
    if (response.status === 401 || response.status === 403) { expired.value = true; error.value = 'Please sign in again. Your draft is kept in this tab.'; return }
    if (response.status === 409) { error.value = 'This entry already exists. Your text has been kept.'; return }
    if (response.status === 400 || response.status === 413) { error.value = 'Use 1 to 2000 characters.'; return }
    if (response.status !== 201) { throw new Error('Write not confirmed') }
    const result = await response.json()
    if (!Number.isInteger(result.Data?.ID) || result.Data.ID <= 0 || typeof result.Data.Content !== 'string') { throw new Error('Write not confirmed') }
    saved.value = result.Data
    content.value = ''
  } catch { error.value = 'The save was not confirmed. Your text is kept; check before retrying.' } finally { busy.value = false }
  await nextTick()
  if (!expired.value) { input.value?.focus() }
}

async function logout () {
  if (busy.value) { return }
  busy.value = true; error.value = ''
  try {
    const response = await command('/api/auth/logout')
    if (!response.ok && response.status !== 401) { throw new Error('Sign out failed') }
    try { sessionStorage.removeItem(draftKey.value) } catch { /* Storage may be disabled. */ }
    account.value = null; content.value = ''; saved.value = null
    window.location.replace('/login')
  } catch { error.value = 'Could not sign out. Please try again.' } finally { busy.value = false }
}
</script>

<style scoped>
.account-page { width: 100%; max-width: 42rem; margin: 0 auto; padding: 1.5rem; letter-spacing: 0; }
.account-heading { display: flex; align-items: center; gap: 1rem; margin-bottom: 2rem; flex-wrap: wrap; }
.account-heading > div { flex: 1; min-width: 7rem; }
h1 { font-size: 1.5rem; font-weight: 600; }
h2 { font-size: 1.125rem; font-weight: 600; }
.entry-form { display: flex; flex-direction: column; gap: 0.75rem; margin-top: 1rem; }
.account-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.75rem; }
.account-page .btn { white-space: normal; height: auto; min-height: 3rem; padding: 0.65rem 1rem; border-radius: 6px; }
.account-page textarea { width: 100%; resize: vertical; min-height: 12rem; }
.saved-entry { margin-top: 2rem; padding-top: 1.5rem; border-top: 1px solid currentColor; }
.saved-entry blockquote { white-space: pre-wrap; overflow-wrap: anywhere; margin: 1rem 0; }
.entry-id { opacity: 0.65; margin-left: 0.5rem; }
.saved-entry a { display: inline-block; margin-top: 1rem; text-decoration: underline; }
</style>
