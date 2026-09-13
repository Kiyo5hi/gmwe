<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
<style>
:root { --ink: #262c30; --muted: #737b80; --accent: #23775f; --line: #e4e8e5; --surface: #ffffff; --wash: #f6f8f7; --tint: #edf5f1; }
html[data-theme=dark] { --ink: #e4e8e6; --muted: #979fa3; --accent: #8bcdb5; --line: #343a3d; --surface: #202528; --wash: #1a1e21; --tint: #293c35; }
html, body { background: var(--surface); color: var(--ink); font-family: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; letter-spacing: 0; }
* { box-sizing: border-box; }
.app-content .btn { min-height: 42px; height: 42px; font-size: 13px; font-weight: 600; border-radius: 6px; text-transform: none; box-shadow: none; }
.app-content .btn-primary { background: var(--accent); border-color: var(--accent); color: var(--surface); }
.app-content .btn-ghost { color: var(--muted); }
.app-content .btn-outline { color: var(--ink); border-color: var(--line); }
.app-content .input, .app-content .select, .app-content .textarea { background: var(--surface); color: var(--ink); border-color: var(--line); box-shadow: none; font-weight: 400; }
button, input, textarea, select { font: inherit; }
button, a { -webkit-tap-highlight-color: transparent; }
button:focus-visible, a:focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; }
.app-content h1 { font-size: 30px; font-weight: 650; line-height: 1.25; }
@media (max-width: 767px) { .app-content h1 { font-size: 26px; } }
</style>
