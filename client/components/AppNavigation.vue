<template>
  <nav class="app-navigation bg-base-100" aria-label="主导航">
    <NuxtLink v-for="item in items" :key="item.path" :to="item.path" :aria-current="active(item.path) ? 'page' : undefined" :class="{ selected: active(item.path) }">
      <component :is="item.icon" :size="22" aria-hidden="true" /><span>{{ item.label }}</span>
    </NuxtLink>
  </nav>
</template>
<script setup lang="ts">
import { BookOpen, House, Quote, UserRound } from '@lucide/vue'
const route = useRoute()
const items = [{ path: '/', label: 'Home', icon: House }, { path: '/hitokoto', label: 'Hitokoto', icon: Quote }, { path: '/our-story', label: 'Story', icon: BookOpen }, { path: '/account', label: 'Account', icon: UserRound }]
const active = (path: string) => route.path.replace(/\/$/, '') === path.replace(/\/$/, '')
</script>
<style scoped>
.app-navigation { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); background: transparent; }
a { display: flex; align-items: center; justify-content: center; gap: 8px; min-height: 48px; color: var(--muted); font-size: 14px; border-bottom: 2px solid transparent; }
a.selected { color: var(--accent); border-color: currentColor; font-weight: 650; }
a:focus-visible { outline: 2px solid currentColor; outline-offset: -4px; }
@media (min-width: 768px) {
  .app-navigation { position: absolute; top: 0; left: 50%; transform: translateX(-50%); width: 408px; height: 80px; }
  a :deep(svg) { display: none; }
}
@media (min-width: 768px) and (max-width: 899px) { .app-navigation { width: 360px; } }
@media (max-width: 767px) {
  .app-navigation { grid-row: 3; grid-template-rows: 4rem; background: var(--surface); padding-bottom: env(safe-area-inset-bottom, 0px); border-top: 1px solid var(--line); border-bottom: 0; }
  a { flex-direction: column; gap: 0.2rem; height: 4rem; font-size: 0.75rem; border-bottom: 0; border-top: 3px solid transparent; }
}
</style>
