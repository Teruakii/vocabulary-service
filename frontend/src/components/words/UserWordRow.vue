<template>
  <li class="card word-row">
    <button type="button" class="row-head" :aria-expanded="expanded" @click="toggle">
      <span class="row-text">
        <strong class="row-word">{{ item.word }}</strong>
        <span class="caption">{{ item.meaning }}</span>
      </span>
      <span class="row">
        <span :class="statusClass(item.status)">{{ statusLabel(item.status) }}</span>
        <span class="caption">{{ relativeDue(item.nextReviewAt) }}</span>
        <Icon :name="expanded ? 'expand_less' : 'expand_more'" :size="20" />
      </span>
    </button>

    <div v-if="expanded" class="stack">
      <hr class="divider" />

      <div class="stack">
        <span class="label">หมวดหมู่</span>
        <p v-if="tags.kind === 'loading'" class="caption" role="status">กำลังโหลดหมวดหมู่…</p>
        <p v-else-if="tags.kind === 'error'" class="alert-error" role="alert">{{ tags.message }}</p>
        <template v-else-if="tags.kind === 'ready'">
          <p v-if="tags.list.length === 0" class="caption">ยังไม่ได้จัดหมวดหมู่ เลือกหมวดด้านล่างเพื่อเพิ่ม</p>
          <ul v-else role="list" class="row">
            <li v-for="tag in tags.list" :key="tag.id" class="category-chip">
              <span>{{ tag.name }}</span>
              <button
                type="button"
                class="chip-remove"
                :aria-label="`เอาออกจากหมวด ${tag.name}`"
                :disabled="busy"
                @click="untag(tag.id)"
              >
                <Icon name="close" :size="16" />
              </button>
            </li>
          </ul>

          <form class="row" @submit.prevent="addTag">
            <label class="visually-hidden" :for="`tag-select-${item.id}`">เลือกหมวดหมู่</label>
            <select :id="`tag-select-${item.id}`" v-model="selectedCategory" class="input tag-select" :disabled="busy">
              <option :value="null" disabled>เลือกหมวดหมู่</option>
              <option v-for="c in untagged" :key="c.id" :value="c.id" :disabled="isFull(c)">
                {{ c.name }} ({{ c.wordCount }}/{{ c.maxWords }}){{ isFull(c) ? ' เต็มแล้ว' : '' }}
              </option>
            </select>
            <button type="submit" class="btn-secondary" :disabled="busy || selectedCategory === null">
              {{ busy ? 'กำลังบันทึก…' : 'เพิ่มหมวด' }}
            </button>
          </form>
          <p v-if="tagError" class="alert-error" role="alert">{{ tagError }}</p>
        </template>
      </div>

      <div class="row">
        <template v-if="confirmingRemove">
          <span class="caption">ลบ "{{ item.word }}" ออกจากคำของฉัน?</span>
          <button type="button" class="btn-danger" :disabled="removing" @click="remove">
            {{ removing ? 'กำลังลบ…' : 'ยืนยันลบ' }}
          </button>
          <button type="button" class="btn-ghost" :disabled="removing" @click="confirmingRemove = false">ยกเลิก</button>
        </template>
        <button v-else type="button" class="btn-ghost" @click="confirmingRemove = true">
          <Icon name="delete" :size="20" />
          <span>ลบออกจากรายการ</span>
        </button>
      </div>
    </div>
  </li>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/Icon.vue'
import { categories, userWords, type CategoryView, type UserWordDetail } from '@/api/client'
import { relativeDue, statusClass, statusLabel } from '@/utils/format'
import { useToast } from '@/composables/useToast'
import { messageOf } from '@/api/client'

const props = defineProps<{ item: UserWordDetail; categories: CategoryView[] }>()
const emit = defineEmits<{ removed: [id: number]; tagsChanged: [] }>()

type Tags =
  | { kind: 'idle' }
  | { kind: 'loading' }
  | { kind: 'ready'; list: CategoryView[] }
  | { kind: 'error'; message: string }

const { push } = useToast()
const expanded = ref(false)
const tags = ref<Tags>({ kind: 'idle' })
const selectedCategory = ref<number | null>(null)
const busy = ref(false)
const tagError = ref('')
const confirmingRemove = ref(false)
const removing = ref(false)

const untagged = computed(() => {
  const current = tags.value.kind === 'ready' ? tags.value.list : []
  return props.categories.filter((c) => !current.some((t) => t.id === c.id))
})

function isFull(c: CategoryView) {
  return c.wordCount >= c.maxWords
}

async function loadTags() {
  tags.value = { kind: 'loading' }
  try {
    tags.value = { kind: 'ready', list: await categories.tagsOf(props.item.id) }
  } catch (error) {
    tags.value = { kind: 'error', message: messageOf(error) }
  }
}

function toggle() {
  expanded.value = !expanded.value
  if (expanded.value && tags.value.kind !== 'ready') void loadTags()
}

async function changeTags(action: () => Promise<void>) {
  busy.value = true
  tagError.value = ''
  try {
    await action()
    const list = await categories.tagsOf(props.item.id)
    tags.value = { kind: 'ready', list }
    emit('tagsChanged')
  } catch (error) {
    tagError.value = messageOf(error)
  } finally {
    busy.value = false
  }
}

async function addTag() {
  const categoryId = selectedCategory.value
  if (categoryId === null) return
  await changeTags(() => categories.tag(props.item.id, categoryId))
  if (!tagError.value) selectedCategory.value = null
}

function untag(categoryId: number) {
  return changeTags(() => categories.untag(props.item.id, categoryId))
}

async function remove() {
  removing.value = true
  try {
    await userWords.remove(props.item.id)
    emit('removed', props.item.id)
  } catch (error) {
    push(messageOf(error))
    removing.value = false
  }
}
</script>

<style scoped>
.word-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  padding: var(--space-md) var(--space-lg);
}

.row-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  width: 100%;
  padding: 0;
  border: none;
  background: none;
  text-align: left;
}

.row-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.row-word {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
  overflow-wrap: anywhere;
}

.chip-remove {
  display: inline-flex;
  padding: 0;
  border: none;
  background: none;
  color: inherit;
}

.tag-select {
  flex: 1 1 200px;
  width: auto;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
}
</style>
