<template>
  <div class="drawer drawer-mobile">
    <input id="sidebar" type="checkbox" class="drawer-toggle">
    <div class="drawer-content flex flex-col">
      <div class="sticky top-0 opacity-80">
        <AppNavbar />
      </div>

      <div class="p-4 h-full">
        <slot />
      </div>
    </div>

    <div class="drawer-side">
      <label for="sidebar" class="drawer-overlay" />
      <AppSidebar />
    </div>
  </div>
</template>
