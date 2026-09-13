<template>
  <header class="app-header">
    <AppLogo />
    <ThemeSelector />
  </header>
</template>
<style scoped>
.app-header { display: flex; align-items: center; justify-content: space-between; height: 80px; padding: 0 32px; }
@media (max-width: 767px) { .app-header { height: 64px; padding: 0 20px; } }
</style>
