<template>
  <nav class="app-navigation bg-base-100" aria-label="主导航">
    <NuxtLink v-for="item in items" :key="item.path" :to="item.path" :aria-current="active(item.path) ? 'page' : undefined" :class="{ selected: active(item.path) }">
      <component :is="item.icon" :size="22" aria-hidden="true" /><span>{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>
<script setup lang="ts">
import { House, Quote, UserRound } from '@lucide/vue'
const route = useRoute()
const items = [{ path: '/', label: 'Home', icon: House }, { path: '/hitokoto', label: '一言', icon: Quote }, { path: '/account', label: 'Account', icon: UserRound }]
const active = (path: string) => route.path.replace(/\/$/, '') === path.replace(/\/$/, '')
</script>
<style scoped>
.app-navigation { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); border-bottom: 1px solid #8884; }
a { display: flex; align-items: center; justify-content: center; gap: 0.6rem; min-height: 3.5rem; border-bottom: 3px solid transparent; }
a.selected { color: #21825c; border-color: currentColor; font-weight: 600; }
a:focus-visible { outline: 2px solid currentColor; outline-offset: -4px; }
@media (max-width: 767px) {
  .app-navigation { position: fixed; bottom: 0; left: 0; right: 0; z-index: 40; padding-bottom: env(safe-area-inset-bottom); border-top: 1px solid #8884; border-bottom: 0; }
  a { flex-direction: column; gap: 0.2rem; height: 4rem; font-size: 0.75rem; border-bottom: 0; border-top: 3px solid transparent; }
}
</style>
