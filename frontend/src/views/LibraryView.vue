<template>
  <div class="stack-lg">
    <h1>คลังคำศัพท์</h1>

    <div class="panels">
      <form class="card stack" role="search" @submit.prevent="runSearch">
        <h2 class="form-title">ค้นหาคำ</h2>
        <div class="field">
          <label for="search-word">คำศัพท์ภาษาอังกฤษ (ตรงทั้งคำ)</label>
          <input id="search-word" v-model="query" class="input" type="search" autocomplete="off" />
        </div>
        <div class="row">
          <button type="submit" class="btn-secondary" :disabled="search.kind === 'loading' || !query.trim()">
            {{ search.kind === 'loading' ? 'กำลังค้นหา…' : 'ค้นหา' }}
          </button>
          <button v-if="search.kind !== 'idle'" type="button" class="btn-ghost" @click="clearSearch">ล้างการค้นหา</button>
        </div>
        <p v-if="search.kind === 'missing'" class="caption" role="status">ไม่พบคำนี้ ลองเพิ่มเป็นคำใหม่ได้</p>
        <p v-else-if="search.kind === 'error'" class="alert-error" role="alert">{{ search.message }}</p>
      </form>

      <form class="card stack" @submit.prevent="createWord">
        <h2 class="form-title">เพิ่มคำใหม่เข้าคลัง</h2>
        <div class="field">
          <label for="create-word">คำศัพท์</label>
          <input id="create-word" v-model="newWord" class="input" required autocomplete="off" :disabled="creating" />
        </div>
        <div class="field">
          <label for="create-meaning">ความหมาย</label>
          <input id="create-meaning" v-model="newMeaning" class="input" required autocomplete="off" :disabled="creating" />
        </div>
        <label v-if="isAdmin" class="row checkbox">
          <input v-model="newGlobal" type="checkbox" :disabled="creating" />
          <span>เป็นคำส่วนกลาง (ทุกคนเห็น)</span>
        </label>
        <button type="submit" class="btn-primary" :disabled="creating || !canCreate">
          {{ creating ? 'กำลังบันทึก…' : 'เพิ่มคำ' }}
        </button>
        <p v-if="createError" class="alert-error" role="alert">{{ createError }}</p>
      </form>
    </div>

    <section v-if="search.kind === 'found'" class="stack" aria-label="ผลการค้นหา">
      <h2 class="form-title">ผลการค้นหา</h2>
      <ul role="list" class="stack">
        <LibraryWordItem
          :key="search.word.id"
          :word="search.word"
          :can-edit="canEdit(search.word)"
          @updated="onUpdated"
          @removed="onRemoved"
        />
      </ul>
    </section>

    <section class="stack" aria-label="คำทั้งหมด">
      <h2 class="form-title">คำทั้งหมด</h2>
      <p v-if="listState === 'loading'" class="caption" role="status">กำลังโหลดคำ…</p>
      <div v-else-if="listState === 'error'" class="stack">
        <p class="alert-error" role="alert">{{ listError }}</p>
        <button type="button" class="btn-ghost centered" @click="load(true)">ลองอีกครั้ง</button>
      </div>
      <div v-else-if="items.length === 0" class="card empty-state">
        <Icon name="local_library" />
        <p>คลังยังว่างอยู่ เริ่มเพิ่มคำแรกจากฟอร์มด้านบนได้เลย</p>
      </div>
      <template v-else>
        <ul role="list" class="stack">
          <LibraryWordItem
            v-for="word in items"
            :key="word.id"
            :word="word"
            :can-edit="canEdit(word)"
            @updated="onUpdated"
            @removed="onRemoved"
          />
        </ul>
        <button v-if="hasMore" type="button" class="btn-ghost centered" :disabled="loadingMore" @click="load(false)">
          {{ loadingMore ? 'กำลังโหลด…' : 'โหลดเพิ่ม' }}
        </button>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/Icon.vue'
import LibraryWordItem from '@/components/words/LibraryWordItem.vue'
import { auth, words, type Word } from '@/api/client'
import { useToast } from '@/composables/useToast'
import { isStatus, messageOf } from '@/api/client'

const PAGE_SIZE = 20

const { push } = useToast()

const isAdmin = ref(false)
auth
  .me()
  .then((user) => {
    isAdmin.value = user.role === 'admin'
  })
  .catch((error: unknown) => push(`โหลดข้อมูลผู้ใช้ไม่สำเร็จ: ${messageOf(error)}`))

// Lists hold global words plus the caller's own, so any word with an owner belongs to the caller.
function canEdit(word: Word) {
  return word.ownerUserId !== null || isAdmin.value
}

const items = ref<Word[]>([])
const hasMore = ref(false)
const listState = ref<'loading' | 'ready' | 'error'>('loading')
const listError = ref('')
const loadingMore = ref(false)

async function load(reset: boolean) {
  if (reset) listState.value = 'loading'
  else loadingMore.value = true
  try {
    const page = await words.list(PAGE_SIZE, reset ? 0 : items.value.length)
    if (reset) {
      items.value = page.items
    } else {
      // A word created on this page was prepended locally and can show up again in a later page.
      const seen = new Set(items.value.map((w) => w.id))
      items.value = [...items.value, ...page.items.filter((w) => !seen.has(w.id))]
    }
    hasMore.value = page.hasMore
    listState.value = 'ready'
  } catch (error) {
    if (reset) {
      listError.value = messageOf(error)
      listState.value = 'error'
    } else {
      push(messageOf(error))
    }
  } finally {
    loadingMore.value = false
  }
}

void load(true)

type Search =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'found'; word: Word }
  | { kind: 'missing' }
  | { kind: 'error'; message: string }

const query = ref('')
const search = ref<Search>({ kind: 'idle' })

async function runSearch() {
  const term = query.value.trim()
  if (!term) return
  search.value = { kind: 'loading' }
  try {
    search.value = { kind: 'found', word: await words.search(term) }
  } catch (error) {
    search.value = isStatus(error, 404) ? { kind: 'missing' } : { kind: 'error', message: messageOf(error) }
  }
}

function clearSearch() {
  query.value = ''
  search.value = { kind: 'idle' }
}

function onUpdated(updated: Word) {
  items.value = items.value.map((w) => (w.id === updated.id ? updated : w))
  if (search.value.kind === 'found' && search.value.word.id === updated.id) {
    search.value = { kind: 'found', word: updated }
  }
}

function onRemoved(id: number) {
  items.value = items.value.filter((w) => w.id !== id)
  if (search.value.kind === 'found' && search.value.word.id === id) search.value = { kind: 'idle' }
}

const newWord = ref('')
const newMeaning = ref('')
const newGlobal = ref(false)
const creating = ref(false)
const createError = ref('')
const canCreate = computed(() => newWord.value.trim() !== '' && newMeaning.value.trim() !== '')

async function createWord() {
  if (!canCreate.value) return
  creating.value = true
  createError.value = ''
  try {
    const created = await words.add(
      newWord.value.trim(),
      newMeaning.value.trim(),
      isAdmin.value ? newGlobal.value : undefined,
    )
    items.value = [created, ...items.value]
    newWord.value = ''
    newMeaning.value = ''
    newGlobal.value = false
  } catch (error) {
    createError.value = messageOf(error)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.panels {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-lg);
  align-items: start;
}

.form-title {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
}

.checkbox {
  cursor: pointer;
}

.centered {
  align-self: center;
}

@media (max-width: 1023px) {
  .panels {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
