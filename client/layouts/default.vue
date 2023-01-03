<template>
  <div class="drawer drawer-mobile">
    <input id="sidebar" type="checkbox" class="drawer-toggle">
    <div class="drawer-content flex flex-col">
      <div class="h-full w-full overflow-y-scroll">
        <AppNavbar />
        <div class="p-4">
          <slot />
        </div>
      </div>
    </div>

    <div class="drawer-side">
      <label for="sidebar" class="drawer-overlay" />
      <AppSidebar />
    </div>
  </div>
</template>
