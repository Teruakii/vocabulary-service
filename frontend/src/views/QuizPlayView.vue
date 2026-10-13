<template>
  <p v-if="state.kind === 'loading'" class="caption center">กำลังโหลดแบบทดสอบ…</p>

  <div v-else-if="state.kind === 'error'" class="card stack center">
    <p class="alert-error" role="alert">{{ state.message }}</p>
    <div class="row actions">
      <button type="button" class="btn-primary" @click="load">ลองอีกครั้ง</button>
      <RouterLink :to="{ name: 'quizzes' }" class="btn-ghost">กลับไปหน้าแบบทดสอบ</RouterLink>
    </div>
  </div>

  <QuizSummary v-else-if="state.kind === 'finished'" :quiz="state.quiz" />

  <div v-else-if="question" class="stack-lg">
    <div class="stack-xs">
      <div class="row-between caption">
        <span>ข้อ {{ state.index + 1 }} / {{ state.quiz.totalQuestions }}</span>
        <span>ตอบแล้ว {{ answeredCount }} ข้อ</span>
      </div>
      <div
        class="progress-track"
        role="progressbar"
        :aria-valuenow="answeredCount"
        aria-valuemin="0"
        :aria-valuemax="state.quiz.totalQuestions"
      >
        <div class="progress-fill" :style="{ width: `${progressPercent}%` }" />
      </div>
    </div>

    <section class="card stack question-card">
      <p class="caption">
        {{ question.questionType === 'typed' ? 'พิมพ์ความหมายภาษาไทยของคำนี้' : 'เลือกความหมายที่ถูกต้องของคำนี้' }}
      </p>
      <p class="word" lang="en">{{ question.word }}</p>
    </section>

    <div v-if="question.questionType === 'multiple_choice'" class="stack options">
      <button
        v-for="(option, i) in question.options ?? []"
        :key="option"
        type="button"
        :class="optionClass(option)"
        :disabled="state.kind !== 'answering' || state.submitting"
        @click="submit(option)"
      >
        <span class="option-key">{{ i + 1 }}</span>
        <span>{{ option }}</span>
      </button>
      <p class="caption hint">กดปุ่ม 1-{{ question.options?.length ?? 4 }} บนแป้นพิมพ์เพื่อเลือกได้</p>
    </div>

    <form v-else class="stack" @submit.prevent="submit(typed)">
      <div class="field">
        <label for="typed-answer">ความหมายภาษาไทย</label>
        <input
          id="typed-answer"
          ref="typedInput"
          v-model="typed"
          class="input"
          type="text"
          autocomplete="off"
          :disabled="state.kind !== 'answering' || state.submitting"
        />
      </div>
      <button
        v-if="state.kind === 'answering'"
        type="submit"
        class="btn-primary"
        :disabled="state.submitting || !typed.trim()"
      >
        {{ state.submitting ? 'กำลังบันทึก…' : 'ส่งคำตอบ' }}
      </button>
    </form>

    <p v-if="state.kind === 'answering' && state.error" class="alert-error" role="alert">{{ state.error }}</p>

    <div v-if="state.kind === 'showingResult'" class="stack" aria-live="polite">
      <div :class="state.result.isCorrect ? 'quiz-option-correct' : 'quiz-option-wrong'" class="feedback">
        <Icon :name="state.result.isCorrect ? 'check_circle' : 'cancel'" :size="20" />
        <div>
          <p class="feedback-title">{{ state.result.isCorrect ? 'ถูกต้อง' : 'ยังไม่ถูก' }}</p>
          <p v-if="!state.result.isCorrect">
            เฉลย: <strong>{{ state.result.correctAnswer }}</strong>
          </p>
        </div>
      </div>
      <button ref="nextButton" type="button" class="btn-primary" @click="next">
        {{ hasUnanswered ? 'ข้อต่อไป' : 'ดูผลสรุป' }}
        <Icon name="arrow_forward" :size="20" />
      </button>
      <p class="caption hint">กด Enter เพื่อไปต่อ</p>
    </div>

    <hr class="divider" />

    <div v-if="confirmingEnd" class="row-between end-confirm">
      <span class="caption">ข้อที่ยังไม่ได้ตอบจะนับเป็นผิด จบเลยไหม?</span>
      <div class="row">
        <button type="button" class="btn-secondary" :disabled="completing" @click="finishEarly">
          {{ completing ? 'กำลังบันทึก…' : 'จบเลย' }}
        </button>
        <button type="button" class="btn-ghost" :disabled="completing" @click="confirmingEnd = false">
          ทำต่อ
        </button>
      </div>
    </div>
    <div v-else class="row end-row">
      <button type="button" class="btn-ghost" @click="confirmingEnd = true">
        <Icon name="flag" :size="20" />
        จบแบบทดสอบ
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import Icon from '@/components/Icon.vue'
import QuizSummary from '@/components/quiz/QuizSummary.vue'
import { ApiError, quizzes, type QuestionView, type QuizDetail } from '@/api/client'
import { useToast } from '@/composables/useToast'

type PlayState =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'answering'; quiz: QuizDetail; index: number; submitting: boolean; error: string }
  | { kind: 'showingResult'; quiz: QuizDetail; index: number; result: QuestionView }
  | { kind: 'finished'; quiz: QuizDetail }

const route = useRoute()
const { push } = useToast()

const state = ref<PlayState>({ kind: 'loading' })
const typed = ref('')
const confirmingEnd = ref(false)
const completing = ref(false)
const typedInput = ref<HTMLInputElement | null>(null)
const nextButton = ref<HTMLButtonElement | null>(null)

const quizId = computed(() => Number(route.params.id))

const question = computed<QuestionView | undefined>(() => {
  const s = state.value
  return s.kind === 'answering' || s.kind === 'showingResult' ? s.quiz.questions[s.index] : undefined
})

const answeredCount = computed(() => {
  const s = state.value
  return s.kind === 'answering' || s.kind === 'showingResult' ? s.quiz.questions.filter((q) => q.answered).length : 0
})

const progressPercent = computed(() => {
  const s = state.value
  if (s.kind !== 'answering' && s.kind !== 'showingResult') return 0
  return s.quiz.totalQuestions > 0 ? (answeredCount.value / s.quiz.totalQuestions) * 100 : 0
})

const hasUnanswered = computed(() => {
  const s = state.value
  return s.kind === 'showingResult' && s.quiz.questions.some((q) => !q.answered)
})

function errorMessage(err: unknown): string {
  return err instanceof ApiError ? err.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
}

function optionClass(option: string): string {
  const s = state.value
  if (s.kind !== 'showingResult') return 'quiz-option'
  if (option === s.result.correctAnswer) return 'quiz-option-correct'
  if (option === s.result.userAnswer) return 'quiz-option-wrong'
  return 'quiz-option'
}

// Resumes at the first unanswered question, so a reload or a 409 lands where the user left off.
function enter(quiz: QuizDetail) {
  confirmingEnd.value = false
  const index = quiz.questions.findIndex((q) => !q.answered)
  if (quiz.completedAt || index === -1) {
    state.value = { kind: 'finished', quiz }
    return
  }
  typed.value = ''
  state.value = { kind: 'answering', quiz, index, submitting: false, error: '' }
  void nextTick(() => typedInput.value?.focus())
}

async function load() {
  state.value = { kind: 'loading' }
  try {
    enter(await quizzes.get(quizId.value))
  } catch (err) {
    state.value = { kind: 'error', message: errorMessage(err) }
  }
}

async function submit(answer: string) {
  const s = state.value
  if (s.kind !== 'answering' || s.submitting) return
  const current = s.quiz.questions[s.index]
  if (!current) return
  s.submitting = true
  s.error = ''
  try {
    const result = { ...current, ...(await quizzes.answer(s.quiz.id, current.id, answer)) }
    s.quiz.questions[s.index] = result
    state.value = { kind: 'showingResult', quiz: s.quiz, index: s.index, result }
    void nextTick(() => nextButton.value?.focus())
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      push('ข้อนี้ตอบไปแล้ว กำลังโหลดแบบทดสอบใหม่')
      await load()
      return
    }
    s.submitting = false
    s.error = errorMessage(err)
  }
}

async function next() {
  const s = state.value
  if (s.kind !== 'showingResult') return
  const index = s.quiz.questions.findIndex((q) => !q.answered)
  if (index !== -1) {
    typed.value = ''
    state.value = { kind: 'answering', quiz: s.quiz, index, submitting: false, error: '' }
    void nextTick(() => typedInput.value?.focus())
    return
  }
  // The backend completes the quiz on the last answer; refetch to get the final score.
  await load()
}

async function finishEarly() {
  const s = state.value
  if (s.kind !== 'answering' && s.kind !== 'showingResult') return
  completing.value = true
  try {
    enter(await quizzes.complete(s.quiz.id))
  } catch (err) {
    push(errorMessage(err))
  } finally {
    completing.value = false
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.altKey || e.ctrlKey || e.metaKey || e.repeat) return
  const s = state.value
  if (s.kind === 'showingResult' && e.key === 'Enter') {
    // preventDefault stops a focused button from also firing its click on the same Enter.
    e.preventDefault()
    void next()
    return
  }
  if (s.kind !== 'answering' || e.target instanceof HTMLInputElement) return
  const options = question.value?.questionType === 'multiple_choice' ? question.value.options : undefined
  const option = options?.[Number(e.key) - 1]
  if (option !== undefined) void submit(option)
}

watch(quizId, load, { immediate: true })
onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>

<style scoped>
.center {
  text-align: center;
}

.actions {
  justify-content: center;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.question-card {
  align-items: center;
  padding: var(--space-xl) var(--space-lg);
  text-align: center;
  box-shadow: var(--shadow-raised);
  border-radius: var(--rounded-xl);
}

.word {
  font-size: var(--display-size);
  font-weight: var(--display-weight);
  line-height: var(--display-line-height);
  letter-spacing: var(--display-letter-spacing);
  overflow-wrap: anywhere;
}

.options .quiz-option,
.options .quiz-option-correct,
.options .quiz-option-wrong {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.option-key {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: none;
  width: 28px;
  height: 28px;
  border-radius: var(--rounded-full);
  background: var(--color-surface-variant);
  color: var(--color-on-surface-variant);
  font-size: var(--label-md-size);
  font-weight: var(--label-md-weight);
}

.hint {
  text-align: center;
}

.feedback {
  display: flex;
  align-items: flex-start;
  gap: var(--space-sm);
}

.feedback-title {
  font-weight: var(--title-md-weight);
}

.end-row {
  justify-content: center;
}
</style>
