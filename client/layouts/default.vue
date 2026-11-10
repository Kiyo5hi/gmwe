<template>
  <div class="app-shell">
    <AppNavbar />
    <AppNavigation />
    <main :key="route.path" class="app-content">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import AppNavbar from '../components/AppNavbar.vue'
import AppNavigation from '../components/AppNavigation.vue'

const route = useRoute()
</script>

<style scoped>
.app-shell { box-sizing: border-box; min-height: 100dvh; width: 100%; max-width: 1120px; margin: 0 auto; position: relative; padding: env(safe-area-inset-top, 0px) env(safe-area-inset-right, 0px) 0 env(safe-area-inset-left, 0px); }
.app-content { border-top: 1px solid var(--line); padding: 48px 32px 64px; min-width: 0; overflow-wrap: anywhere; }
@media (max-width: 767px) {
  .app-shell { height: 100dvh; min-height: 0; display: grid; grid-template-columns: minmax(0, 1fr); grid-template-rows: auto minmax(0, 1fr) auto; overflow: hidden; }
  .app-content { grid-row: 2; min-height: 0; overflow-x: hidden; overflow-y: auto; overscroll-behavior-y: contain; padding: 28px 20px; }
  .app-shell .app-content :deep(input), .app-shell .app-content :deep(select), .app-shell .app-content :deep(textarea) { font-size: 16px; min-width: 0; max-width: 100%; }
}
</style>
