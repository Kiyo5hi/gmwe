<template>
  <RouterLink to="/" class="wordmark">
    Gummy<span>We</span><span class="mark" aria-hidden="true">.</span>
  </RouterLink>
</template>
<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
<style scoped>
.wordmark { font-size: 24px; font-weight: 750; text-decoration: none; color: var(--ink); }
.wordmark > span { color: var(--accent); }
.mark { margin-left: 2px; }
@media (max-width: 767px) { .wordmark { font-size: 21px; } }
</style>
