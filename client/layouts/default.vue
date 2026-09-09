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
.app-shell { min-height: 100dvh; max-width: 68rem; margin: 0 auto; }
.app-content { padding: 1.5rem; min-width: 0; }
@media (max-width: 767px) { .app-content { padding: 1.25rem 1rem calc(6rem + env(safe-area-inset-bottom)); } }
</style>
