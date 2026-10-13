<template>
  <dialog ref="dialog" class="words-dialog" aria-labelledby="words-dialog-title" @close="emit('close')">
    <header class="dialog-header">
      <div class="stack-xs">
        <h2 id="words-dialog-title" class="dialog-title">จัดคำในหมวด "{{ category.name }}"</h2>
        <span class="caption">เลือกแล้ว {{ selected.size }} / {{ category.maxWords }} คำ</span>
      </div>
      <button type="button" class="icon-btn" aria-label="ปิด" @click="dialog?.close()">
        <Icon name="close" />
      </button>
    </header>

    <input
      v-model.trim="search"
      class="input"
      type="search"
      placeholder="ค้นหาคำหรือความหมาย"
      aria-label="ค้นหาคำ"
    />

    <p v-if="state.phase === 'loading'" class="caption" role="status">กำลังโหลดคำของคุณ…</p>
    <p v-else-if="state.phase === 'error'" class="alert-error" role="alert">{{ state.message }}</p>
    <div v-else-if="state.words.length === 0" class="empty-state">
      <Icon name="bookmarks" />
      <p>ยังไม่มีคำในรายการของคุณ</p>
      <RouterLink :to="{ name: 'library' }" class="btn-primary">ไปเพิ่มคำจากคลัง</RouterLink>
    </div>
    <p v-else-if="visible.length === 0" class="caption">ไม่พบคำที่ตรงกับ "{{ search }}"</p>
    <ul v-else class="word-list" role="list">
      <li v-for="w in visible" :key="w.id">
        <label class="word-option" :class="{ checked: selected.has(w.id) }">
          <input
            type="checkbox"
            :checked="selected.has(w.id)"
            :disabled="busy.has(w.id) || (!selected.has(w.id) && full)"
            @change="toggle(w)"
          />
          <span class="word-text">
            <span class="word">{{ w.word }}</span>
            <span class="caption">{{ w.meaning }}</span>
          </span>
        </label>
      </li>
    </ul>

    <p v-if="full" class="caption">หมวดนี้เต็มแล้ว เอาคำออกก่อนถึงจะเพิ่มคำใหม่ได้</p>

    <footer class="dialog-footer">
      <button type="button" class="btn-primary" @click="dialog?.close()">เสร็จแล้ว</button>
    </footer>
  </dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/Icon.vue'
import { categories, messageOf, userWords, type CategoryView, type UserWordDetail } from '@/api/client'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ category: CategoryView }>()
const emit = defineEmits<{ close: []; count: [number] }>()

const LIMIT = 100

type State =
  | { phase: 'loading' }
  | { phase: 'error'; message: string }
  | { phase: 'ready'; words: UserWordDetail[] }

const { push } = useToast()
const dialog = ref<HTMLDialogElement | null>(null)
const state = ref<State>({ phase: 'loading' })
const selected = reactive(new Set<number>())
const busy = reactive(new Set<number>())
const search = ref('')

const full = computed(() => selected.size >= props.category.maxWords)
const visible = computed(() => {
  if (state.value.phase !== 'ready') return []
  const q = search.value.toLowerCase()
  return q
    ? state.value.words.filter((w) => w.word.toLowerCase().includes(q) || w.meaning.toLowerCase().includes(q))
    : state.value.words
})

async function toggle(w: UserWordDetail) {
  const adding = !selected.has(w.id)
  busy.add(w.id)
  try {
    if (adding) await categories.tag(w.id, props.category.id)
    else await categories.untag(w.id, props.category.id)
    if (adding) selected.add(w.id)
    else selected.delete(w.id)
    emit('count', selected.size)
  } catch (error) {
    push(`${adding ? 'เพิ่ม' : 'เอา'} "${w.word}" ${adding ? 'เข้า' : 'ออกจาก'}หมวดไม่สำเร็จ: ${messageOf(error)}`)
  } finally {
    busy.delete(w.id)
  }
}

onMounted(async () => {
  dialog.value?.showModal()
  try {
    const [all, tagged] = await Promise.all([
      userWords.list({ limit: LIMIT }),
      userWords.list({ categoryId: props.category.id, limit: LIMIT }),
    ])
    for (const w of tagged.items) selected.add(w.id)
    state.value = { phase: 'ready', words: all.items }
  } catch (error) {
    state.value = { phase: 'error', message: messageOf(error) }
  }
})
</script>

<style scoped>
.words-dialog {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  width: min(560px, calc(100vw - 2 * var(--space-md)));
  max-height: min(720px, calc(100vh - 2 * var(--space-lg)));
  margin: auto;
  padding: var(--space-lg);
  border: none;
  border-radius: var(--rounded-xl);
  background: var(--color-surface);
  color: var(--color-on-surface);
  box-shadow: var(--shadow-overlay);
}

.words-dialog:not([open]) {
  display: none;
}

.words-dialog::backdrop {
  background: rgba(40, 46, 62, 0.45);
}

.dialog-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-md);
}

.dialog-title {
  margin: 0;
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
  overflow-wrap: anywhere;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.icon-btn {
  display: grid;
  flex: none;
  place-items: center;
  width: 40px;
  height: 40px;
  border: none;
  border-radius: var(--rounded-full);
  background: var(--color-surface-variant);
  color: var(--color-on-surface-variant);
  cursor: pointer;
}

.word-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  min-height: 0;
  margin: 0;
  padding: 0;
  overflow-y: auto;
  list-style: none;
}

.word-option {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px var(--space-md);
  border: 2px solid var(--color-outline);
  border-radius: var(--rounded-md);
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.word-option.checked {
  border-color: var(--color-primary);
  background: var(--color-primary-container);
}

.word-option input {
  flex: none;
  width: 20px;
  height: 20px;
  accent-color: var(--color-primary);
}

.word-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.word {
  font-weight: 800;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
