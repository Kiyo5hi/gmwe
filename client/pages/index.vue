<template>
  <section class="home-page">
    <header class="home-heading">
      <h1>Home</h1><p>This is for and only for my Acey.</p>
    </header>
    <div class="home-grid">
      <section class="daily-note">
        <div class="section-label">
          <Quote :size="17" /><NuxtLink to="/hitokoto">
            Hitokoto<ArrowUpRight :size="15" />
          </NuxtLink>
        </div>
        <p v-if="hitokotoPending" class="note-status" role="status">
          加载中…
        </p>
        <p v-else-if="!hitokoto" class="note-status">
          暂时无法加载
        </p>
        <template v-else>
          <blockquote>{{ hitokoto.Data.Content }}</blockquote>
          <p class="note-author">
            {{ hitokoto.Data.User.Name }}
          </p>
        </template>
        <div class="together">
          <div class="portraits" aria-hidden="true">
            <span>A</span><span>K</span>
          </div>
          <span>Acey &amp; Kiyoshi</span>
          <NuxtLink to="/our-story">
            Our story<ArrowUpRight :size="16" />
          </NuxtLink>
        </div>
      </section>
      <MusicPlayer name="暗號" artist="Jay Chou" url="/music.mp3" cover="/cover.webp" />
    </div>
  </section>
</template>
<script lang="ts" setup>
import { ArrowUpRight, Quote } from '@lucide/vue'
const { data: hitokoto, pending: hitokotoPending } = useLazyFetch<Hitokoto>('/api/v1/hitokoto')
</script>
<style scoped>
.home-page { max-width: 880px; margin: 0 auto; }
.home-heading p { font-size: 14px; margin-top: 10px; color: var(--muted); }
.home-grid { display: grid; grid-template-columns: minmax(0, 1fr) 256px; gap: 64px; align-items: start; margin-top: 42px; }
.daily-note { min-width: 0; }
.section-label { display: flex; align-items: center; gap: 10px; color: var(--accent); font-size: 12px; }
.section-label a { display: inline-flex; align-items: center; gap: 6px; }
blockquote { margin: 24px 0 18px; font-family: Georgia, 'Songti SC', serif; font-size: 26px; line-height: 1.75; white-space: pre-wrap; overflow-wrap: anywhere; }
.note-author, .note-status { font-size: 13px; color: var(--muted); }
.note-status { min-height: 100px; padding-top: 24px; }
.together { display: flex; align-items: center; gap: 12px; border-top: 1px solid var(--line); padding-top: 28px; margin-top: 32px; font-size: 12px; flex-wrap: wrap; }
.portraits { display: flex; padding-right: 8px; }
.portraits span { display: grid; place-items: center; border-radius: 50%; border: 3px solid var(--surface); margin-right: -8px; width: 44px; height: 44px; background: var(--tint); color: var(--accent); font-weight: 600; }
.portraits span:first-child { background: #bc72851a; color: #ae677b; }
.together a { display: flex; align-items: center; gap: 4px; color: var(--muted); margin-left: auto; }
@media (max-width: 767px) { .home-grid { grid-template-columns: minmax(0, 1fr); gap: 32px; margin-top: 30px; } blockquote { font-size: 23px; } }
</style>
