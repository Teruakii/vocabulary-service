<template>
  <div class="stack-lg">
    <h1>แบบทดสอบ</h1>

    <form class="card stack" @submit.prevent="create">
      <h2 class="section-title">สร้างแบบทดสอบใหม่</h2>

      <fieldset class="stack-sm" :disabled="creating">
        <legend class="label">จำนวนข้อ</legend>
        <div class="row">
          <button
            v-for="preset in PRESETS"
            :key="preset"
            type="button"
            class="btn-ghost choice"
            :class="{ selected: count === preset }"
            :aria-pressed="count === preset"
            @click="count = preset"
          >
            {{ preset }} ข้อ
          </button>
          <div class="count-input">
            <label for="quiz-count" class="visually-hidden">กำหนดจำนวนข้อเอง (1-50)</label>
            <input
              id="quiz-count"
              v-model.number="count"
              class="input"
              type="number"
              min="1"
              max="50"
              step="1"
              inputmode="numeric"
            />
          </div>
        </div>
        <p class="caption">เลือกได้ตั้งแต่ 1 ถึง 50 ข้อ</p>
      </fieldset>

      <fieldset class="stack-sm" :disabled="creating">
        <legend class="label">คำที่ใช้ออกข้อสอบ</legend>
        <div class="row">
          <button
            v-for="option in SOURCES"
            :key="option.value"
            type="button"
            class="btn-ghost choice"
            :class="{ selected: source === option.value }"
            :aria-pressed="source === option.value"
            @click="source = option.value"
          >
            {{ option.label }}
          </button>
        </div>
        <p class="caption">{{ sourceHint }}</p>
      </fieldset>

      <div class="row">
        <button type="submit" class="btn-secondary" :disabled="creating || !countValid">
          <Icon name="quiz" :size="20" />
          {{ creating ? 'กำลังสร้าง…' : 'สร้างแบบทดสอบ' }}
        </button>
      </div>

      <div v-if="createError" class="alert-error stack-sm" role="alert">
        <p>{{ createError.message }}</p>
        <div v-if="createError.noWords" class="row">
          <button
            v-if="source === 'due'"
            type="button"
            class="btn-ghost"
            @click="switchToAll"
          >
            เปลี่ยนเป็น "ทั้งหมด"
          </button>
          <RouterLink :to="{ name: 'library' }" class="btn-ghost">ไปเพิ่มคำจากคลัง</RouterLink>
        </div>
      </div>
    </form>

    <section class="stack">
      <h2 class="section-title">ประวัติแบบทดสอบ</h2>

      <p v-if="history.loading && history.items.length === 0" class="caption">กำลังโหลดประวัติ…</p>

      <div v-else-if="history.error && history.items.length === 0" class="alert-error row-between" role="alert">
        <span>{{ history.error }}</span>
        <button type="button" class="btn-ghost" @click="loadMore">ลองอีกครั้ง</button>
      </div>

      <div v-else-if="history.items.length === 0" class="card empty-state">
        <Icon name="quiz" />
        <p>ยังไม่เคยทำแบบทดสอบ เลือกจำนวนข้อด้านบนแล้วเริ่มข้อแรกได้เลย</p>
      </div>

      <template v-else>
        <ul class="card history">
          <li v-for="quiz in history.items" :key="quiz.id" class="history-item row-between">
            <div class="stack-xs">
              <span class="history-date">{{ formatDate(quiz.createdAt) }}</span>
              <span class="row">
                <span class="score-badge">{{ quiz.score }} / {{ quiz.totalQuestions }}</span>
                <span class="caption">ถูก {{ percent(quiz) }}%</span>
              </span>
            </div>
            <RouterLink
              v-if="quiz.completedAt"
              :to="{ name: 'quiz-play', params: { id: quiz.id } }"
              class="btn-ghost"
            >
              <Icon name="check_circle" :size="20" />
              ทำเสร็จแล้ว
            </RouterLink>
            <RouterLink v-else :to="{ name: 'quiz-play', params: { id: quiz.id } }" class="btn-ghost resume">
              <Icon name="play_arrow" :size="20" />
              ทำต่อ
            </RouterLink>
          </li>
        </ul>

        <p v-if="history.error" class="alert-error" role="alert">{{ history.error }}</p>

        <div v-if="history.hasMore" class="row load-more">
          <button type="button" class="btn-ghost" :disabled="history.loading" @click="loadMore">
            {{ history.loading ? 'กำลังโหลด…' : 'โหลดเพิ่ม' }}
          </button>
        </div>
      </template>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import Icon from '@/components/Icon.vue'
import { ApiError, quizzes, type QuizSource, type QuizSummary } from '@/api/client'
import { formatDate } from '@/utils/format'

const PAGE_SIZE = 20
const PRESETS = [5, 10, 20] as const
const SOURCES: readonly { value: QuizSource; label: string; hint: string }[] = [
  { value: 'due', label: 'ถึงกำหนด', hint: 'ใช้เฉพาะคำที่ถึงกำหนดทบทวนแล้ว' },
  { value: 'all', label: 'ทั้งหมด', hint: 'สุ่มจากคำทั้งหมดในรายการของคุณ' },
]

const router = useRouter()

const count = ref<number | ''>(10)
const source = ref<QuizSource>('all')
const creating = ref(false)
const createError = ref<{ message: string; noWords: boolean } | null>(null)

const history = reactive({
  items: [] as QuizSummary[],
  hasMore: false,
  loading: false,
  error: '',
})

const countValid = computed(
  () => typeof count.value === 'number' && Number.isInteger(count.value) && count.value >= 1 && count.value <= 50,
)
const sourceHint = computed(() => SOURCES.find((s) => s.value === source.value)?.hint ?? '')

function errorMessage(err: unknown): string {
  return err instanceof ApiError ? err.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
}

function percent(quiz: QuizSummary): number {
  return quiz.totalQuestions > 0 ? Math.round((quiz.score / quiz.totalQuestions) * 100) : 0
}

async function create() {
  if (!countValid.value || typeof count.value !== 'number') return
  creating.value = true
  createError.value = null
  try {
    const quiz = await quizzes.create(count.value, source.value)
    await router.push({ name: 'quiz-play', params: { id: quiz.id } })
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      createError.value = {
        noWords: true,
        message:
          source.value === 'due'
            ? 'ตอนนี้ยังไม่มีคำที่ถึงกำหนดทบทวน ลองเปลี่ยนเป็น "ทั้งหมด" หรือเพิ่มคำใหม่จากคลังคำศัพท์'
            : 'ยังไม่มีคำในรายการของคุณให้ออกข้อสอบ ไปเพิ่มคำจากคลังคำศัพท์ก่อนนะ',
      }
    } else {
      createError.value = { noWords: false, message: errorMessage(err) }
    }
  } finally {
    creating.value = false
  }
}

function switchToAll() {
  source.value = 'all'
  createError.value = null
}

async function loadMore() {
  history.loading = true
  history.error = ''
  try {
    const page = await quizzes.list(PAGE_SIZE, history.items.length)
    history.items.push(...page.items)
    history.hasMore = page.hasMore
  } catch (err) {
    history.error = errorMessage(err)
  } finally {
    history.loading = false
  }
}

onMounted(loadMore)
</script>

<style scoped>
.section-title {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
}

fieldset {
  margin: 0;
  padding: 0;
  border: none;
  min-width: 0;
}

legend {
  padding: 0;
  margin-bottom: var(--space-xs);
}

.stack-sm {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.choice {
  border: 1px solid var(--color-outline);
}

.choice.selected {
  background: var(--color-primary-container);
  color: var(--color-on-primary-container);
  border-color: var(--color-primary);
}

.count-input {
  width: 96px;
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  white-space: nowrap;
}

.history {
  margin: 0;
  padding: var(--space-sm) var(--space-lg);
  list-style: none;
}

.history-item {
  padding: var(--space-md) 0;
}

.history-item + .history-item {
  border-top: 1px solid var(--color-surface-variant);
}

.history-date {
  font-weight: var(--label-md-weight);
}

.resume {
  color: var(--color-primary);
}

.load-more {
  justify-content: center;
}
</style>
