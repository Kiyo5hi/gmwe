<template>
  <div
    class="
  sticky top-0 z-30 flex h-16 w-full justify-center bg-opacity-90 backdrop-blur transition-all duration-100
  bg-base-300 text-base-content shadow-sm
  "
  >
    <nav class="navbar w-full">
      <div class="flex-none lg:hidden">
        <label for="sidebar" class="btn btn-square btn-ghost">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            class="inline-block w-6 h-6 stroke-current"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </label>
      </div>
      <div class="flex-1 px-2 mx-2">
        <div class="lg:hidden">
          <AppLogo />
        </div>
        <div class="ml-2">
          🐕 🌼 🦆
        </div>
      </div>
    </nav>
  </div>
</template>
