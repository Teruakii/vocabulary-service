<template>
  <div class="stack-lg">
    <h1>คำของฉัน</h1>

    <form class="card stack" @submit.prevent="addWord">
      <h2 class="form-title">เพิ่มคำใหม่</h2>
      <div class="add-grid">
        <div class="field">
          <label for="new-word">คำศัพท์</label>
          <input id="new-word" v-model="newWord" class="input" required autocomplete="off" :disabled="adding" />
        </div>
        <div class="field">
          <label for="new-meaning">ความหมาย</label>
          <input id="new-meaning" v-model="newMeaning" class="input" required autocomplete="off" :disabled="adding" />
        </div>
        <button type="submit" class="btn-primary" :disabled="adding || !canAdd">
          {{ adding ? 'กำลังบันทึก…' : 'เพิ่มคำ' }}
        </button>
      </div>
      <p v-if="addError" class="alert-error" role="alert">{{ addError }}</p>
    </form>

    <div class="row-between">
      <div class="row" role="tablist" aria-label="กรองตามสถานะ">
        <button
          v-for="tab in STATUS_TABS"
          :key="tab.label"
          type="button"
          role="tab"
          :aria-selected="status === tab.value"
          :class="status === tab.value ? 'nav-item-active' : 'nav-item'"
          class="tab"
          @click="status = tab.value"
        >
          {{ tab.label }}
        </button>
      </div>
      <div class="field category-filter">
        <label for="category-filter">หมวดหมู่</label>
        <select id="category-filter" v-model="categoryId" class="input">
          <option :value="undefined">ทุกหมวดหมู่</option>
          <option v-for="c in categoryList" :key="c.id" :value="c.id">{{ c.name }} ({{ c.wordCount }})</option>
        </select>
      </div>
    </div>

    <p v-if="listState === 'loading'" class="caption" role="status">กำลังโหลดคำ…</p>
    <div v-else-if="listState === 'error'" class="stack">
      <p class="alert-error" role="alert">{{ listError }}</p>
      <button type="button" class="btn-ghost retry" @click="load(true)">ลองอีกครั้ง</button>
    </div>
    <div v-else-if="items.length === 0" class="card empty-state">
      <Icon name="bookmarks" />
      <template v-if="filtered">
        <p>ไม่มีคำที่ตรงกับตัวกรองนี้</p>
        <button type="button" class="btn-ghost" @click="clearFilters">ล้างตัวกรอง</button>
      </template>
      <template v-else>
        <p>ยังไม่มีคำในรายการ เพิ่มคำด้านบนหรือเลือกจากคลังคำศัพท์</p>
        <RouterLink :to="{ name: 'library' }" class="btn-secondary">ไปที่คลังคำศัพท์</RouterLink>
      </template>
    </div>
    <template v-else>
      <ul role="list" class="stack">
        <UserWordRow
          v-for="item in items"
          :key="item.id"
          :item="item"
          :categories="categoryList"
          @removed="onRemoved"
          @tags-changed="loadCategories"
        />
      </ul>
      <button v-if="hasMore" type="button" class="btn-ghost load-more" :disabled="loadingMore" @click="load(false)">
        {{ loadingMore ? 'กำลังโหลด…' : 'โหลดเพิ่ม' }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/Icon.vue'
import UserWordRow from '@/components/words/UserWordRow.vue'
import { categories, userWords, type CategoryView, type UserWordDetail, type WordStatus } from '@/api/client'
import { statusLabel } from '@/utils/format'
import { useToast } from '@/composables/useToast'
import { messageOf } from '@/api/client'

const PAGE_SIZE = 20
const STATUS_TABS: readonly { value: WordStatus | undefined; label: string }[] = [
  { value: undefined, label: 'ทั้งหมด' },
  { value: 'new', label: statusLabel('new') },
  { value: 'learning', label: statusLabel('learning') },
  { value: 'mastered', label: statusLabel('mastered') },
]

const { push } = useToast()

const status = ref<WordStatus | undefined>()
const categoryId = ref<number | undefined>()
const filtered = computed(() => status.value !== undefined || categoryId.value !== undefined)

const items = ref<UserWordDetail[]>([])
const hasMore = ref(false)
const listState = ref<'loading' | 'ready' | 'error'>('loading')
const listError = ref('')
const loadingMore = ref(false)
// Filters can change while a page is in flight; only the newest request may write the list.
let latestRequest = 0

async function load(reset: boolean) {
  const request = ++latestRequest
  if (reset) listState.value = 'loading'
  else loadingMore.value = true
  try {
    const page = await userWords.list({
      status: status.value,
      categoryId: categoryId.value,
      limit: PAGE_SIZE,
      offset: reset ? 0 : items.value.length,
    })
    if (request !== latestRequest) return
    items.value = reset ? page.items : [...items.value, ...page.items]
    hasMore.value = page.hasMore
    listState.value = 'ready'
  } catch (error) {
    if (request !== latestRequest) return
    if (reset) {
      listError.value = messageOf(error)
      listState.value = 'error'
    } else {
      push(messageOf(error))
    }
  } finally {
    if (request === latestRequest) loadingMore.value = false
  }
}

watch([status, categoryId], () => load(true), { immediate: true })

function clearFilters() {
  status.value = undefined
  categoryId.value = undefined
}

function onRemoved(id: number) {
  items.value = items.value.filter((item) => item.id !== id)
  void loadCategories()
}

const categoryList = ref<CategoryView[]>([])

async function loadCategories() {
  try {
    categoryList.value = await categories.list()
  } catch (error) {
    push(`โหลดหมวดหมู่ไม่สำเร็จ: ${messageOf(error)}`)
  }
}

void loadCategories()

const newWord = ref('')
const newMeaning = ref('')
const adding = ref(false)
const addError = ref('')
const canAdd = computed(() => newWord.value.trim() !== '' && newMeaning.value.trim() !== '')

async function addWord() {
  if (!canAdd.value) return
  adding.value = true
  addError.value = ''
  try {
    await userWords.add({ word: newWord.value.trim(), meaning: newMeaning.value.trim() })
    newWord.value = ''
    newMeaning.value = ''
    await load(true)
  } catch (error) {
    addError.value = messageOf(error)
  } finally {
    adding.value = false
  }
}
</script>

<style scoped>
.form-title {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
}

.add-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 2fr) auto;
  align-items: end;
  gap: var(--space-md);
}

.tab {
  border: none;
}

.category-filter {
  min-width: 220px;
}

.load-more,
.retry {
  align-self: center;
}

@media (max-width: 639px) {
  .add-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .category-filter {
    width: 100%;
  }
}
</style>
