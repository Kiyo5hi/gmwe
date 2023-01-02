<template>
  <div class="h-full flex items-center justify-center">
    <div class="text-center">
      <p>❤️ This is for and only for my Acey. ❤️</p>
      <img class="w-64 h-auto" src="gmwe.webp" alt="GMWE Logo">
      <div>
        <p>
          {{ pending ? 'Loading...' : `『${data.Data.Content}』 —— ${data.Data.User.Name}` }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
const { pending, data } = await useLazyFetch('/api/v1/hitokoto')
</script>
