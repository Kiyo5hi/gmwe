<template>
  <section class="hitokoto-page">
    <header class="page-heading">
      <div><h1>Hitokoto</h1><p>{{ total }} 条</p></div>
      <button class="btn btn-primary" :aria-expanded="composing" aria-controls="entry-composer" @click="composing = !composing">
        <component :is="composing ? X : Plus" :size="18" aria-hidden="true" />{{ composing ? '收起' : '新增' }}
      </button>
    </header>
    <div v-show="composing" id="entry-composer">
      <EntryComposer @saved="entrySaved" />
    </div>
    <form class="search-form" role="search" @submit.prevent="search">
      <label class="sr-only" for="entry-search">搜索 Hitokoto</label>
      <input
        id="entry-search"
        v-model="query"
        type="search"
        maxlength="100"
        class="input input-bordered"
        placeholder="搜索 Hitokoto"
        :disabled="pending"
      >
      <button class="btn btn-square" :disabled="pending" aria-label="搜索" title="搜索">
        <Search :size="20" />
      </button>
    </form>
    <p v-if="pending" role="status" class="status">
      加载中…
    </p>
    <div v-else-if="error" role="alert" class="status">
      <p class="text-error">
        {{ error }}
      </p>
      <a v-if="expired" href="/api/auth/login" class="link">重新登录</a>
      <button v-else class="btn btn-ghost" @click="loadPage(page)">
        重试
      </button>
    </div>
    <p v-else-if="!entries.length" role="status" class="status">
      {{ appliedQuery ? '没有找到匹配的内容。' : '还没有内容。' }}
    </p>
    <ol v-else class="entry-list" :aria-busy="pending">
      <li v-for="entry in entries" :key="entry.ID">
        <p class="entry-content">
          {{ entry.Content }}
        </p>
        <footer><span>{{ entry.User.Name }}</span><time :datetime="entry.CreatedAt">{{ formatDate(entry.CreatedAt) }}</time><span>#{{ entry.ID }}</span></footer>
      </li>
    </ol>
    <nav class="pagination" aria-label="Hitokoto 分页">
      <button class="btn btn-square btn-ghost" :disabled="pending || page <= 1" aria-label="上一页" title="上一页" @click="loadPage(page - 1)">
        <ChevronLeft :size="22" />
      </button>
      <span>{{ page }} / {{ Math.max(1, Math.ceil(total / 20)) }}</span>
      <button class="btn btn-square btn-ghost" :disabled="pending || page * 20 >= total" aria-label="下一页" title="下一页" @click="loadPage(page + 1)">
        <ChevronRight :size="22" />
      </button>
    </nav>
  </section>
</template>
<script setup lang="ts">
import { ChevronLeft, ChevronRight, Plus, Search, X } from '@lucide/vue'
type Entry = { ID: number, Content: string, CreatedAt: string, User: { Name: string } }
const entries = ref<Entry[]>([])
const total = ref(0)
const page = ref(1)
const query = ref('')
const appliedQuery = ref('')
const pending = ref(true)
const error = ref('')
const expired = ref(false)
const composing = ref(false)
let request = 0
useHead({ title: 'Hitokoto | GMWE' })
const formatDate = (value: string) => new Date(value).toLocaleDateString('zh-CN')
onMounted(() => loadPage(1))
onBeforeUnmount(() => { request++ })
async function loadPage (number: number) {
  const current = ++request
  pending.value = true; error.value = ''; expired.value = false
  try {
    const params = new URLSearchParams({ page: String(number), q: appliedQuery.value })
    const response = await fetch('/api/v1/hitokotos?' + params, { cache: 'no-store' })
    if (current !== request) { return }
    if (response.status === 401) { expired.value = true; throw new Error('Login required') }
    if (!response.ok) { throw new Error('List unavailable') }
    const result = await response.json()
    if (current !== request) { return }
    if (!Array.isArray(result.Data) || !Number.isInteger(result.Total)) { throw new TypeError('Invalid list response') }
    entries.value = result.Data; total.value = result.Total; page.value = number
  } catch {
    if (current === request) { error.value = expired.value ? '请重新登录。' : '无法加载内容，请重试。' }
  } finally { if (current === request) { pending.value = false } }
}
function search () { appliedQuery.value = query.value.trim(); loadPage(1) }
function entrySaved () { query.value = ''; appliedQuery.value = ''; loadPage(1) }
</script>
<style scoped>
.hitokoto-page { max-width: 46rem; margin: 0 auto; min-width: 0; }
.page-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
h1 { font-size: 1.5rem; font-weight: 600; }
.page-heading p { margin-top: 0.25rem; opacity: 0.65; }
.btn { border-radius: 6px; gap: 0.5rem; }
.search-form { display: flex; gap: 0.5rem; margin: 1.5rem 0 0.5rem; }
.search-form input { flex: 1; min-width: 0; border-radius: 6px; }
.entry-list { list-style: none; }
.entry-list li { padding: 1.25rem 0; border-bottom: 1px solid #8884; }
.entry-content { white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.75; }
footer { display: flex; flex-wrap: wrap; gap: 0.75rem; font-size: 0.75rem; margin-top: 0.75rem; opacity: 0.65; }
.status { padding: 2rem 0; }
.pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; margin-top: 1rem; }
.pagination span { min-width: 5rem; text-align: center; font-variant-numeric: tabular-nums; }
</style>
