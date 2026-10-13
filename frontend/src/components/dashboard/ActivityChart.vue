<template>
  <p v-if="total === 0" class="caption">ยังไม่มีกิจกรรมใน {{ points.length }} วันที่ผ่านมา เริ่มทบทวนวันนี้เลย</p>
  <ul class="chart" role="list">
    <li v-for="bar in bars" :key="bar.date" class="bar-row" :title="bar.title">
      <span class="caption day">{{ bar.weekday }}</span>
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: `${bar.percent}%` }" />
      </div>
      <span class="caption count">{{ bar.count }}</span>
    </li>
  </ul>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ActivityPoint } from '@/api/client'

const props = defineProps<{ points: ActivityPoint[] }>()

const weekdayFormat = new Intl.DateTimeFormat('th-TH', { weekday: 'short' })
const fullFormat = new Intl.DateTimeFormat('th-TH', { weekday: 'long', day: 'numeric', month: 'short' })

// "YYYY-MM-DD" parsed by Date() is UTC midnight, which can land on the previous day locally.
function localDate(iso: string): Date {
  const [y, m, d] = iso.split('-').map(Number)
  return new Date(y!, m! - 1, d!)
}

const total = computed(() => props.points.reduce((sum, p) => sum + p.count, 0))

const bars = computed(() => {
  const max = Math.max(1, ...props.points.map((p) => p.count))
  return props.points.map((p) => {
    const day = localDate(p.date)
    return {
      date: p.date,
      count: p.count,
      weekday: weekdayFormat.format(day),
      title: `${fullFormat.format(day)}: ${p.count} ครั้ง`,
      percent: (p.count / max) * 100,
    }
  })
})
</script>

<style scoped>
.chart {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.bar-row {
  display: grid;
  grid-template-columns: 3rem minmax(0, 1fr) 2.5rem;
  align-items: center;
  gap: var(--space-sm);
}

.count {
  text-align: right;
}
</style>
