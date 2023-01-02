<template>
  <aside class="bg-base-200 w-80">
    <div class="z-20 bg-base-200 bg-opacity-90 backdrop-blur sticky top-0 items-center gap-2 px-4 py-2 hidden lg:flex">
      <AppLogo />
    </div>

    <div class="h-4" />

    <ul v-for="route in routes" :key="route.name" class="menu flex flex-col p-0 px-4">
      <li>
        <NuxtLink :to="route.path" :class="currentRoute.path === route.path ? 'active' : ''">
          <span class="capitalize">
            {{ route.name }}
          </span>
        </NuxtLink>
      </li>
    </ul>
  </aside>
</template>

<script setup>
const routes = getSortedRoutes()
routes.forEach((r) => { r.name = r.name.replaceAll('-', ' ') })
const currentRoute = useRoute()
</script>
