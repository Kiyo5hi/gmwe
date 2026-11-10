<template>
  <section class="music-player" aria-label="音乐播放器">
    <img class="cover" :src="cover" :alt="name + ' 专辑封面'" width="256" height="256">
    <div class="music-controls">
      <div class="track">
        <h2>{{ name }}</h2><p>{{ artist }}</p>
      </div>
      <button type="button" :aria-label="isPlaying ? '暂停' : '播放'" :title="isPlaying ? '暂停' : '播放'" @click="toggle">
        <Pause v-if="isPlaying" :size="20" /><Play v-else :size="20" />
      </button>
    </div>
    <p v-if="failed" class="play-error" role="alert">
      无法播放，请重试。
    </p>
    <audio
      ref="audio"
      :src="url"
      preload="none"
      @play="isPlaying = true"
      @pause="isPlaying = false"
      @ended="isPlaying = false"
    />
  </section>
</template>
<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'

import { Pause, Play } from '@lucide/vue'
defineProps<{ name: string, artist: string, url: string, cover: string }>()
const audio = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const failed = ref(false)
async function toggle () {
  failed.value = false
  if (!audio.value) { return }
  if (!audio.value.paused) { audio.value.pause(); return }
  try { await audio.value.play() } catch { failed.value = true }
}
onBeforeUnmount(() => audio.value?.pause())
</script>
<style scoped>
.music-player { min-width: 0; }
.cover { display: block; width: 100%; height: auto; aspect-ratio: 1; object-fit: cover; border-radius: 6px; }
.music-controls { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding-top: 18px; }
.track { min-width: 0; }
.track h2 { font-size: 17px; font-weight: 600; }
.track p { color: var(--muted); font-size: 13px; margin-top: 4px; }
button { display: grid; place-items: center; flex-shrink: 0; width: 44px; height: 44px; border-radius: 50%; color: var(--surface); background: var(--accent); }
.play-error { color: #bf5266; font-size: 12px; margin-top: 10px; }
@media (max-width: 767px) {
  .music-player { display: grid; grid-template-columns: 80px minmax(0, 1fr); gap: 18px; align-items: center; border-top: 1px solid var(--line); padding-top: 24px; }
  .cover { width: 80px; height: 80px; }
  .music-controls { padding: 0; }
  .play-error { grid-column: 1 / -1; }
}
</style>
