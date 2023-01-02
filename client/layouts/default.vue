<template>
  <div class="drawer drawer-mobile">
    <input id="sidebar" type="checkbox" class="drawer-toggle">
    <div class="drawer-content flex flex-col">
      <AppNavbar />

      <div class="p-4">
        <!-- Page content here -->
        <slot />
      </div>
    </div>

    <div class="drawer-side">
      <label for="sidebar" class="drawer-overlay" />
      <AppSidebar />
    </div>
  </div>
</template>
