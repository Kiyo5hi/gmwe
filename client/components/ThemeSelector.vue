<template>
  <button class="theme-button" type="button" :title="dark ? '切换浅色' : '切换深色'" :aria-label="dark ? '切换浅色' : '切换深色'" @click="toggle">
    <Sun v-if="dark" :size="19" /><Moon v-else :size="19" />
  </button>
</template>
<script setup lang="ts">
import { Moon, Sun } from '@lucide/vue'
const dark = ref(false)
function apply (value: boolean) { dark.value = value; document.documentElement.dataset.theme = value ? 'dark' : 'light' }
onMounted(() => {
  let saved: string | null = null
  try { saved = localStorage.getItem('theme') } catch { /* Theme preference is optional. */ }
  apply(saved ? saved === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches)
})
function toggle () {
  apply(!dark.value)
  try { localStorage.setItem('theme', dark.value ? 'dark' : 'light') } catch { /* Retain the in-memory choice. */ }
}
</script>
<style scoped>
.theme-button { display: grid; place-items: center; width: 40px; height: 40px; color: var(--muted); border: 1px solid var(--line); border-radius: 50%; background: var(--surface); }
.theme-button:hover { color: var(--ink); border-color: var(--accent); }
</style>
