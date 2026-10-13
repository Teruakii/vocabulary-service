<template>
  <div class="stack-lg">
    <section class="card stack summary-head">
      <Icon name="emoji_events" :size="40" />
      <h1>สรุปผลแบบทดสอบ</h1>
      <p class="score">{{ quiz.score }} / {{ quiz.totalQuestions }}</p>
      <div class="progress-track score-track">
        <div class="progress-fill" :style="{ width: `${percent}%` }" />
      </div>
      <p class="caption">ตอบถูก {{ percent }}% · {{ formatDate(quiz.completedAt ?? quiz.createdAt) }}</p>
      <div class="row actions">
        <RouterLink :to="{ name: 'quizzes' }" class="btn-primary">แบบทดสอบใหม่</RouterLink>
        <RouterLink :to="{ name: 'dashboard' }" class="btn-ghost">กลับแดชบอร์ด</RouterLink>
      </div>
    </section>

    <section class="stack">
      <h2 class="section-title">คำตอบของคุณ</h2>
      <ol class="stack answers">
        <li
          v-for="(question, i) in quiz.questions"
          :key="question.id"
          :class="question.isCorrect ? 'quiz-option-correct' : question.answered ? 'quiz-option-wrong' : 'quiz-option'"
        >
          <div class="row-between">
            <span class="word">{{ i + 1 }}. {{ question.word }}</span>
            <Icon :name="question.isCorrect ? 'check_circle' : 'cancel'" :size="20" />
          </div>
          <p>
            คำตอบของคุณ:
            <strong>{{ question.answered ? question.userAnswer || '(เว้นว่าง)' : 'ไม่ได้ตอบ' }}</strong>
          </p>
          <p v-if="!question.isCorrect && question.correctAnswer">
            เฉลย: <strong>{{ question.correctAnswer }}</strong>
          </p>
        </li>
      </ol>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/Icon.vue'
import type { QuizDetail } from '@/api/client'
import { formatDate } from '@/utils/format'

const props = defineProps<{ quiz: QuizDetail }>()

const percent = computed(() =>
  props.quiz.totalQuestions > 0 ? Math.round((props.quiz.score / props.quiz.totalQuestions) * 100) : 0,
)
</script>

<style scoped>
.summary-head {
  align-items: center;
  text-align: center;
  color: var(--color-primary);
}

.summary-head h1 {
  color: var(--color-on-surface);
}

.score {
  color: var(--color-on-surface);
  font-size: var(--display-size);
  font-weight: var(--display-weight);
  line-height: var(--display-line-height);
}

.score-track {
  width: 100%;
  max-width: 320px;
}

.actions {
  justify-content: center;
}

.section-title {
  font-size: var(--title-md-size);
  line-height: var(--title-md-line-height);
}

.answers {
  margin: 0;
  padding: 0;
  list-style: none;
}

.word {
  font-weight: var(--title-md-weight);
  overflow-wrap: anywhere;
}
</style>
