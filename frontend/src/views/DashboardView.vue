<template>
  <div class="stack-lg">
    <header class="stack">
      <h1>{{ greeting }}</h1>
      <div class="hero row">
        <RouterLink v-if="dueCount" :to="{ name: 'review', query: { deck: 'due' } }" class="btn-primary">
          <Icon name="style" :size="20" />
          เริ่มทบทวน ({{ dueCount }} คำ)
        </RouterLink>
        <RouterLink :to="{ name: 'review' }" :class="dueCount ? 'btn-ghost' : 'btn-primary'">
          <Icon name="style" :size="20" />
          เลือกชุด Flashcard
        </RouterLink>
        <p v-if="due.status === 'ready' && dueCount === 0" class="caption">
          วันนี้ไม่มีคำถึงกำหนดทบทวน เลือกฝึกจากหมวดหรือคำทั้งหมดแทนได้
        </p>
      </div>
    </header>

    <p v-if="stats.status === 'error'" class="alert-error">{{ stats.message }}</p>
    <section v-else class="grid-3" aria-label="สถิติ">
      <div class="stat-tile">
        <span class="caption">วันติดต่อกัน (streak)</span>
        <span>{{ stats.status === 'ready' ? stats.data.currentStreak : '…' }}</span>
      </div>
      <div class="stat-tile">
        <span class="caption">คำที่จำได้</span>
        <span>{{ stats.status === 'ready' ? stats.data.totalWordsLearned : '…' }}</span>
      </div>
      <div class="stat-tile">
        <span class="caption">แบบทดสอบที่ทำ</span>
        <span>{{ stats.status === 'ready' ? stats.data.totalQuizTaken : '…' }}</span>
      </div>
    </section>

    <div class="sections">
      <section class="card stack">
        <h3>กิจกรรม 7 วันล่าสุด</h3>
        <p v-if="activity.status === 'loading'" class="caption">กำลังโหลด…</p>
        <p v-else-if="activity.status === 'error'" class="alert-error">{{ activity.message }}</p>
        <ActivityChart v-else :points="activity.data" />
      </section>

      <section class="card stack">
        <h3>คำที่ถึงกำหนดถัดไป</h3>
        <p v-if="due.status === 'loading'" class="caption">กำลังโหลด…</p>
        <p v-else-if="due.status === 'error'" class="alert-error">{{ due.message }}</p>
        <div v-else-if="due.data.length === 0" class="empty-state">
          <Icon name="task_alt" />
          <p>ยังไม่มีคำที่ต้องทบทวน ไปเพิ่มคำจากคลังได้เลย</p>
          <RouterLink :to="{ name: 'library' }" class="btn-ghost">ไปที่คลังคำศัพท์</RouterLink>
        </div>
        <ul v-else class="due-list" role="list">
          <li v-for="item in due.data.slice(0, 5)" :key="item.id" class="row-between">
            <span class="word">{{ item.word }}</span>
            <span :class="statusClass(item.status)">{{ statusLabel(item.status) }}</span>
          </li>
        </ul>
      </section>
    </div>

    <nav class="row" aria-label="ทางลัด">
      <RouterLink :to="{ name: 'quizzes' }" class="btn-secondary">
        <Icon name="quiz" :size="20" />
        ทำแบบทดสอบ
      </RouterLink>
      <RouterLink :to="{ name: 'my-words' }" class="btn-ghost">
        <Icon name="bookmarks" :size="20" />
        คำของฉัน
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, type Ref } from 'vue'
import { RouterLink } from 'vue-router'
import Icon from '@/components/Icon.vue'
import ActivityChart from '@/components/dashboard/ActivityChart.vue'
import { ApiError, stats as statsApi, userWords } from '@/api/client'
import { statusClass, statusLabel } from '@/utils/format'

type Section<T> =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ready'; data: T }

// Each section loads on its own so one failing endpoint does not blank the page.
function load<T>(fetcher: () => Promise<T>): Ref<Section<T>> {
  const section = ref({ status: 'loading' }) as Ref<Section<T>>
  fetcher().then(
    (data) => (section.value = { status: 'ready', data }),
    (e: unknown) =>
      (section.value = {
        status: 'error',
        message: e instanceof ApiError ? e.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้',
      }),
  )
  return section
}

const stats = load(() => statsApi.me())
const activity = load(() => statsApi.activity(7))
const due = load(() => userWords.due(100))

const dueCount = computed(() => (due.value.status === 'ready' ? due.value.data.length : 0))

const greeting = (() => {
  const hour = new Date().getHours()
  if (hour < 12) return 'สวัสดีตอนเช้า'
  if (hour < 17) return 'สวัสดีตอนบ่าย'
  return 'สวัสดีตอนเย็น'
})()
</script>

<style scoped>
.hero {
  gap: var(--space-md);
}

.chip {
  align-self: flex-start;
}

.sections {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-lg);
}

.due-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.word {
  font-weight: var(--title-md-weight);
  overflow-wrap: anywhere;
}

@media (max-width: 1023px) {
  .sections {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
