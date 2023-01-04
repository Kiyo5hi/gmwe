<template>
  <aside class="bg-base-200 w-80">
    <div class="z-20 bg-opacity-90 backdrop-blur sticky top-0 items-center gap-2 px-4 py-2 hidden lg:flex">
      <AppLogo />
    </div>

    <div class="h-4" />

    <ul v-for="route in routes" :key="route.name" class="menu flex flex-col p-0 px-4">
      <li>
        <NuxtLink :to="route.path" :class="{ active: currentRoute.path == route.path }" @click="emit('change')">
          <span class="capitalize">
            {{ route.name }}
          </span>
        </NuxtLink>
      </li>
    </ul>
  </aside>
</template>

<script lang="ts" setup>
const routes = getSortedRoutes()
routes.forEach((r) => { r.name = r.name!.toString().replaceAll('-', ' ') })
const currentRoute = useRoute()

const emit = defineEmits<{(event: 'change'): void}>()
</script>
