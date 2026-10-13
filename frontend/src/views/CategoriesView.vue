<template>
  <div class="stack-lg">
    <header class="stack">
      <h1>หมวดหมู่</h1>
      <p class="caption">จัดกลุ่มคำเป็นเรื่อง ๆ หมวดละไม่เกิน 20 คำ แล้วเล่น Flashcard ทีละหมวด</p>
    </header>

    <form class="card stack" @submit.prevent="create">
      <h2 class="form-title">สร้างหมวดใหม่</h2>
      <div class="create-row">
        <div class="field create-field">
          <label for="category-name">ชื่อหมวด</label>
          <input
            id="category-name"
            v-model="newName"
            class="input"
            type="text"
            maxlength="100"
            placeholder="เช่น คำศัพท์ที่ทำงาน"
            :disabled="creating"
            required
          />
        </div>
        <button type="submit" class="btn-primary" :disabled="creating || !newName.trim()">
          <Icon name="add" :size="20" />
          {{ creating ? 'กำลังบันทึก…' : 'สร้างหมวด' }}
        </button>
      </div>
      <p v-if="createError" class="alert-error" role="alert">{{ createError }}</p>
    </form>

    <p v-if="loading" class="caption">กำลังโหลดหมวดหมู่…</p>

    <div v-else-if="loadError" class="alert-error row-between" role="alert">
      <span>{{ loadError }}</span>
      <button type="button" class="btn-ghost" @click="load">ลองอีกครั้ง</button>
    </div>

    <div v-else-if="list.length === 0" class="card empty-state">
      <Icon name="label" />
      <p>ยังไม่มีหมวดหมู่ หมวดช่วยจัดกลุ่มคำของคุณได้หมวดละไม่เกิน 20 คำ</p>
      <p>ตั้งชื่อหมวดแรกในฟอร์มด้านบน แล้วกด "จัดคำ" เพื่อเลือกคำเข้าหมวด</p>
    </div>

    <ul v-else class="grid-3 category-list">
      <li v-for="category in list" :key="category.id" class="card stack">
        <div class="row-between">
          <h3 class="category-name">{{ category.name }}</h3>
          <span class="category-chip">{{ category.isGlobal ? 'ส่วนกลาง' : 'ของฉัน' }}</span>
        </div>

        <div class="stack-xs">
          <div class="row-between caption">
            <span>{{ category.wordCount }} / {{ category.maxWords }} คำ</span>
            <span v-if="isFull(category)" class="full-label">เต็มแล้ว</span>
          </div>
          <div
            class="progress-track"
            role="progressbar"
            :aria-valuenow="category.wordCount"
            aria-valuemin="0"
            :aria-valuemax="category.maxWords"
          >
            <div class="progress-fill" :style="{ width: `${fillPercent(category)}%` }" />
          </div>
        </div>

        <div class="card-actions">
          <RouterLink
            v-if="category.wordCount > 0"
            :to="{ name: 'review', query: { deck: `c${category.id}` } }"
            class="btn-primary"
          >
            <Icon name="style" :size="20" />
            Flashcard
          </RouterLink>
          <button type="button" class="btn-ghost" @click="editing = category">
            <Icon name="checklist" :size="20" />
            จัดคำ
          </button>
        </div>

        <div class="row-between">
          <RouterLink
            :to="{ name: 'my-words', query: { categoryId: category.id } }"
            class="link"
          >
            ดูคำในหมวดนี้
          </RouterLink>

          <template v-if="!category.isGlobal">
            <div v-if="confirmingId === category.id" class="row">
              <button
                type="button"
                class="btn-danger"
                :disabled="deletingId === category.id"
                @click="remove(category)"
              >
                {{ deletingId === category.id ? 'กำลังลบ…' : 'ยืนยันลบ' }}
              </button>
              <button
                type="button"
                class="btn-ghost"
                :disabled="deletingId === category.id"
                @click="confirmingId = null"
              >
                ยกเลิก
              </button>
            </div>
            <button
              v-else
              type="button"
              class="btn-ghost"
              :aria-label="`ลบหมวด ${category.name}`"
              @click="confirmingId = category.id"
            >
              <Icon name="delete" :size="20" />
              ลบ
            </button>
          </template>
        </div>
      </li>
    </ul>

    <CategoryWordsDialog
      v-if="editing"
      :category="editing"
      @count="setCount(editing, $event)"
      @close="editing = null"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/Icon.vue'
import CategoryWordsDialog from '@/components/CategoryWordsDialog.vue'
import { ApiError, categories, type CategoryView } from '@/api/client'
import { useToast } from '@/composables/useToast'

const { push } = useToast()

const list = ref<CategoryView[]>([])
const loading = ref(true)
const loadError = ref('')

const newName = ref('')
const creating = ref(false)
const createError = ref('')

const editing = ref<CategoryView | null>(null)
const confirmingId = ref<number | null>(null)
const deletingId = ref<number | null>(null)

function errorMessage(err: unknown): string {
  return err instanceof ApiError ? err.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
}

function setCount(category: CategoryView, wordCount: number) {
  list.value = list.value.map((c) => (c.id === category.id ? { ...c, wordCount } : c))
}

function isFull(category: CategoryView): boolean {
  return category.wordCount >= category.maxWords
}

function fillPercent(category: CategoryView): number {
  if (category.maxWords <= 0) return 0
  return Math.min(100, (category.wordCount / category.maxWords) * 100)
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    list.value = await categories.list()
  } catch (err) {
    loadError.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function create() {
  const name = newName.value.trim()
  if (!name) return
  creating.value = true
  createError.value = ''
  try {
    const created = await categories.create(name)
    list.value = [...list.value, created]
    newName.value = ''
  } catch (err) {
    createError.value = errorMessage(err)
  } finally {
    creating.value = false
  }
}

async function remove(category: CategoryView) {
  deletingId.value = category.id
  try {
    await categories.remove(category.id)
    list.value = list.value.filter((c) => c.id !== category.id)
    confirmingId.value = null
  } catch (err) {
    push(errorMessage(err))
  } finally {
    deletingId.value = null
  }
}

onMounted(load)
</script>

<style scoped>
.form-title {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
}

.create-row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: var(--space-md);
}

.create-field {
  flex: 1 1 240px;
}

.category-list {
  padding: 0;
  margin: 0;
  list-style: none;
}

.category-name {
  overflow-wrap: anywhere;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.card-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: var(--space-sm);
}

.card-actions > * {
  padding-inline: var(--space-sm);
}

.link {
  color: var(--color-primary);
  font-weight: 700;
  text-decoration: none;
}

.link:hover {
  text-decoration: underline;
}

.full-label {
  color: var(--color-on-surface);
  font-weight: var(--label-md-weight);
}
</style>
