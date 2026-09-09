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
    <form class="filters" @submit.prevent="search">
      <label>成员
        <select v-model="member" class="select select-bordered" :disabled="pending">
          <option value="">全部成员</option>
          <option v-for="person in members" :key="person.ID" :value="String(person.ID)">{{ person.Name }}</option>
        </select>
      </label>
      <label>排序
        <select v-model="sort" class="select select-bordered" :disabled="pending">
          <option value="newest">最新在前</option><option value="oldest">最早在前</option>
        </select>
      </label>
      <label>开始日期 (UTC)<input v-model="from" class="input input-bordered" type="date" :max="to || undefined" :disabled="pending"></label>
      <label>结束日期 (UTC)<input v-model="to" class="input input-bordered" type="date" :min="from || undefined" :disabled="pending"></label>
      <div class="filter-actions">
        <button class="btn btn-outline" :disabled="pending" type="submit">
          <Filter :size="18" />筛选
        </button>
        <button class="btn btn-ghost" :disabled="pending" type="button" @click="resetFilters">
          <RotateCcw :size="18" />重置
        </button>
      </div>
    </form>
    <p v-if="memberError" class="text-error" role="alert">
      {{ memberError }}
    </p>
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
import { ChevronLeft, ChevronRight, Filter, Plus, RotateCcw, Search, X } from '@lucide/vue'
type Entry = { ID: number, Content: string, CreatedAt: string, User: { Name: string } }
const entries = ref<Entry[]>([])
const total = ref(0)
const page = ref(1)
const query = ref('')
const appliedQuery = ref('')
const members = ref<{ ID: number, Name: string }[]>([])
const memberError = ref('')
const member = ref('')
const from = ref('')
const to = ref('')
const sort = ref('newest')
const filters = ref({ user_id: '', from: '', to: '', sort: 'newest' })
const pending = ref(true)
const error = ref('')
const expired = ref(false)
const composing = ref(false)
let request = 0
useHead({ title: 'Hitokoto | GMWE' })
const formatDate = (value: string) => new Date(value).toLocaleDateString('zh-CN')
onMounted(() => loadPage(1))
onMounted(async () => {
  try {
    const response = await fetch('/api/v1/users', { cache: 'no-store' })
    if (!response.ok) { throw new Error('Members unavailable') }
    const result = await response.json()
    if (!Array.isArray(result.Data)) { throw new TypeError('Members unavailable') }
    members.value = result.Data
  } catch { memberError.value = '成员列表加载失败，请刷新重试。' }
})
onBeforeUnmount(() => { request++ })
async function loadPage (number: number) {
  const current = ++request
  pending.value = true; error.value = ''; expired.value = false
  try {
    const params = new URLSearchParams({ page: String(number), q: appliedQuery.value, ...filters.value })
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
function search () {
  if (from.value && to.value && from.value > to.value) { error.value = '开始日期不能晚于结束日期。'; return }
  appliedQuery.value = query.value.trim()
  filters.value = { user_id: member.value, from: from.value, to: to.value, sort: sort.value }
  loadPage(1)
}
function resetFilters () { query.value = ''; member.value = ''; from.value = ''; to.value = ''; sort.value = 'newest'; search() }
function entrySaved () { resetFilters() }
</script>
<style scoped>
.hitokoto-page { max-width: 46rem; margin: 0 auto; min-width: 0; }
.page-heading { display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
h1 { font-size: 1.5rem; font-weight: 600; }
.page-heading p { margin-top: 0.25rem; opacity: 0.65; }
.btn { border-radius: 6px; gap: 0.5rem; }
.search-form { display: flex; gap: 0.5rem; margin: 1.5rem 0 0.5rem; }
.search-form input { flex: 1; min-width: 0; border-radius: 6px; }
.filters { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.75rem; margin: 1rem 0; }
.filters label { display: flex; flex-direction: column; gap: 0.35rem; min-width: 0; font-size: 0.875rem; }
.filters input, .filters select { width: 100%; min-width: 0; max-width: 100%; border-radius: 6px; }
.filter-actions { grid-column: 1 / -1; display: flex; gap: 0.5rem; }
.entry-list { list-style: none; }
.entry-list li { padding: 1.25rem 0; border-bottom: 1px solid #8884; }
.entry-content { white-space: pre-wrap; overflow-wrap: anywhere; line-height: 1.75; }
footer { display: flex; flex-wrap: wrap; gap: 0.75rem; font-size: 0.75rem; margin-top: 0.75rem; opacity: 0.65; }
.status { padding: 2rem 0; }
.pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; margin-top: 1rem; }
.pagination span { min-width: 5rem; text-align: center; font-variant-numeric: tabular-nums; }
</style>
