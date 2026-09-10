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
.app-navigation { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); border-bottom: 1px solid #8884; }
a { display: flex; align-items: center; justify-content: center; gap: 0.6rem; min-height: 3.5rem; border-bottom: 3px solid transparent; }
a.selected { color: #21825c; border-color: currentColor; font-weight: 600; }
a:focus-visible { outline: 2px solid currentColor; outline-offset: -4px; }
@media (max-width: 767px) {
  .app-navigation { grid-row: 3; grid-template-rows: 4rem; padding-bottom: env(safe-area-inset-bottom, 0px); border-top: 1px solid #8884; border-bottom: 0; }
  a { flex-direction: column; gap: 0.2rem; height: 4rem; font-size: 0.75rem; border-bottom: 0; border-top: 3px solid transparent; }
}
</style>
