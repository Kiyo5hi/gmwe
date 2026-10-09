<template>
  <span class="member-avatar" aria-hidden="true">
    <img v-if="userId && !failed" :src="`/api/auth/avatar/${userId}`" alt="" referrerpolicy="no-referrer" @error="failed = true">
    <span v-else>{{ name.charAt(0).toUpperCase() }}</span>
  </span>
</template>
<script setup lang="ts">
const props = defineProps<{ userId?: number, name: string }>()
const failed = ref(false)
watch(() => props.userId, () => { failed.value = false })
</script>
<style scoped>
.member-avatar { display: inline-grid; place-items: center; width: 36px; height: 36px; flex-shrink: 0; border-radius: 50%; overflow: hidden; background: var(--tint); color: var(--accent); font-weight: 650; }
.member-avatar img { width: 100%; height: 100%; object-fit: cover; }
</style>
