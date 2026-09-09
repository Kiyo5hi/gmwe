<template>
  <div class="app-shell">
    <AppNavbar />
    <AppNavigation />
    <main class="app-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-shell { min-height: 100dvh; width: 100%; max-width: 68rem; margin: 0 auto; }
.app-content { padding: 1.5rem; min-width: 0; overflow-wrap: anywhere; }
@media (max-width: 767px) {
  .app-shell { height: 100dvh; min-height: 0; display: grid; grid-template-columns: minmax(0, 1fr); grid-template-rows: auto minmax(0, 1fr) auto; overflow: hidden; }
  .app-content { grid-row: 2; min-height: 0; overflow-y: auto; overscroll-behavior-y: contain; padding: 1.25rem 1rem; }
  .app-content :deep(input), .app-content :deep(select), .app-content :deep(textarea) { font-size: 16px; min-width: 0; max-width: 100%; }
}
</style>
