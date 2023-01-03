<template>
  <div class="m-auto flex flex-col items-center justify-center text-center">
    <div>
      <p>❤️ This is for and only for my Acey. ❤️</p>
      <img src="/gmwe.webp" alt="GMWE Logo" class="w-56 h-auto m-auto">
    </div>
    <div>
      {{ hitokotoPending ? 'Loading...' : `『${hitokoto.Data.Content}』 —— ${hitokoto.Data.User.Name}` }}
    </div>
    <div class="divider" />
    <div class="h-48 w-5/6">
      <MusicPlayer name="暗號" artist="Jay Chou" url="music.mp3" cover="cover.webp" />
    </div>
  </div>
</template>

<script setup>
const { data: hitokoto, pending: hitokotoPending } = useLazyFetch('/api/v1/hitokoto')
</script>
