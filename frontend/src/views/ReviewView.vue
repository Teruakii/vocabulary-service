<template>
  <div class="review">
    <section v-if="session.phase === 'pick'" class="stack-lg">
      <div class="stack-sm">
        <h1>Flashcard</h1>
        <p class="caption">เลือกชุดคำที่อยากจำ แล้วพลิกการ์ดทีละใบ</p>
      </div>
      <p v-if="pickError" class="alert-error" role="alert">{{ pickError }}</p>
      <ul role="list" class="decks">
        <li v-for="d in deckOptions" :key="d.key">
          <button type="button" class="deck" :disabled="d.count === 0" @click="choose(d.key)">
            <span class="deck-icon" :class="d.tone"><Icon :name="d.icon" /></span>
            <span class="deck-text">
              <span class="deck-name">{{ d.name }}</span>
              <span class="caption">{{ d.hint }}</span>
            </span>
            <span class="deck-count">{{ d.count ?? '…' }}</span>
          </button>
        </li>
      </ul>
      <p v-if="cats.length === 0 && !pickError" class="caption">
        ยังไม่มีหมวดหมู่
        <RouterLink :to="{ name: 'categories' }">สร้างหมวดหมู่</RouterLink>
        แล้วจัดคำเข้าหมวดเพื่อฝึกเป็นเรื่อง ๆ
      </p>
    </section>

    <p v-else-if="session.phase === 'loading'" class="caption" role="status">กำลังโหลดการ์ด…</p>

    <div v-else-if="session.phase === 'error'" class="stack">
      <p class="alert-error" role="alert">{{ session.message }}</p>
      <div class="row">
        <button type="button" class="btn-ghost" @click="pickAnother">เลือกชุดอื่น</button>
        <button type="button" class="btn-primary" @click="start">ลองอีกครั้ง</button>
      </div>
    </div>

    <div v-else-if="session.phase === 'empty'" class="card empty-state">
      <Icon name="task_alt" />
      <p>{{ deck?.kind === 'due' ? 'ยังไม่มีคำที่ถึงกำหนดทบทวน' : 'ชุดนี้ยังไม่มีคำ' }}</p>
      <div class="row">
        <button type="button" class="btn-ghost" @click="pickAnother">เลือกชุดอื่น</button>
        <RouterLink :to="{ name: 'library' }" class="btn-primary">ไปที่คลังคำศัพท์</RouterLink>
      </div>
    </div>

    <section v-else-if="session.phase === 'done'" class="card stack-lg">
      <div class="stack">
        <h1>จบชุด "{{ deckTitle }}" แล้ว</h1>
        <p class="caption">รอบนี้ฝึกไป {{ reviewed }} คำ</p>
      </div>
      <ul role="list" class="tallies">
        <li v-for="r in RATINGS" :key="r.key" :class="r.cls" class="tally">
          <span>{{ r.label }}</span>
          <strong class="tally-count">{{ tally[r.key] }}</strong>
        </li>
      </ul>
      <div class="row">
        <button type="button" class="btn-ghost" @click="pickAnother">เลือกชุดอื่น</button>
        <button type="button" class="btn-primary" @click="start">เล่นชุดนี้อีกรอบ</button>
      </div>
    </section>

    <div v-else class="stack-lg">
      <div class="stack-sm">
        <div class="row-between caption">
          <span class="deck-chip">{{ deckTitle }}</span>
          <span>{{ reviewed }} / {{ total }}</span>
        </div>
        <div
          class="progress-track"
          role="progressbar"
          aria-label="ความคืบหน้าการทบทวน"
          :aria-valuenow="reviewed"
          aria-valuemin="0"
          :aria-valuemax="total"
        >
          <div class="progress-fill" :style="{ width: `${percent}%` }" />
        </div>
      </div>

      <button
        v-if="session.phase === 'front'"
        :key="`front-${session.card.id}`"
        type="button"
        class="flashcard-front card-face"
        aria-label="พลิกการ์ดเพื่อดูความหมาย"
        @click="flip"
      >
        <span>{{ session.card.word }}</span>
        <span class="caption">แตะหรือกด Space เพื่อดูความหมาย</span>
      </button>
      <div v-else :key="`back-${session.card.id}`" class="flashcard-back card-face flip-in" aria-live="polite">
        <span class="back-word">{{ session.card.word }}</span>
        <p>{{ session.card.meaning }}</p>
      </div>

      <div v-if="session.phase === 'back'" class="ratings" role="group" aria-label="ให้คะแนนว่าจำได้แค่ไหน">
        <button v-for="r in RATINGS" :key="r.key" type="button" :class="r.cls" @click="rate(r.key)">
          <span>{{ r.label }}</span>
          <kbd class="hotkey">{{ r.hotkey }}</kbd>
        </button>
      </div>
      <p v-else class="caption hint">กด 1-4 เพื่อให้คะแนนหลังพลิกการ์ด</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Icon from '@/components/Icon.vue'
import {
  REVIEW_QUALITY,
  categories,
  messageOf,
  userWords,
  type CategoryView,
  type UserWordDetail,
} from '@/api/client'
import { useToast } from '@/composables/useToast'

type RatingKey = keyof typeof REVIEW_QUALITY

const RATINGS: readonly { key: RatingKey; label: string; cls: string; hotkey: string }[] = [
  { key: 'again', label: 'ลืม', cls: 'rating-again', hotkey: '1' },
  { key: 'hard', label: 'ยาก', cls: 'rating-hard', hotkey: '2' },
  { key: 'good', label: 'ดี', cls: 'rating-good', hotkey: '3' },
  { key: 'easy', label: 'ง่าย', cls: 'rating-easy', hotkey: '4' },
]

const DUE_LIMIT = 50
const DECK_LIMIT = 100

// The deck lives in ?deck= so a reload or a link from the dashboard resumes it.
type Deck = { kind: 'due' } | { kind: 'all' } | { kind: 'category'; id: number }

function parseDeck(raw: unknown): Deck | null {
  if (raw === 'due' || raw === 'all') return { kind: raw }
  const m = typeof raw === 'string' ? /^c(\d+)$/.exec(raw) : null
  return m ? { kind: 'category', id: Number(m[1]) } : null
}

function deckKey(d: Deck): string {
  return d.kind === 'category' ? `c${d.id}` : d.kind
}

function shuffled<T>(list: T[]): T[] {
  const a = [...list]
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    const tmp = a[i]!
    a[i] = a[j]!
    a[j] = tmp
  }
  return a
}

type Session =
  | { phase: 'pick' }
  | { phase: 'loading' }
  | { phase: 'error'; message: string }
  | { phase: 'empty' }
  | { phase: 'front' | 'back'; card: UserWordDetail }
  | { phase: 'done' }

const { push } = useToast()
const route = useRoute()
const router = useRouter()
const deck = ref<Deck | null>(parseDeck(route.query.deck))
const session = ref<Session>(deck.value ? { phase: 'loading' } : { phase: 'pick' })
const cats = ref<CategoryView[]>([])
const dueCount = ref<number | null>(null)
const allCount = ref<number | null>(null)
const pickError = ref('')
const total = ref(0)
const tally = reactive<Record<RatingKey, number>>({ again: 0, hard: 0, good: 0, easy: 0 })
let upcoming: UserWordDetail[] = []
// Reviews are sent without waiting, so a failure can land after the user started a new round.
let round = 0

const reviewed = computed(() => tally.again + tally.hard + tally.good + tally.easy)
const percent = computed(() => (total.value ? (reviewed.value / total.value) * 100 : 0))

const deckOptions = computed(() => [
  {
    key: 'due',
    name: 'ถึงกำหนดทบทวน',
    hint: 'คำที่ระบบนัดให้ทบทวนวันนี้',
    icon: 'schedule',
    tone: 'tone-primary',
    count: dueCount.value,
  },
  {
    key: 'all',
    name: 'คำของฉันทั้งหมด',
    hint: 'สุ่มลำดับทุกคำที่บันทึกไว้',
    icon: 'bookmarks',
    tone: 'tone-secondary',
    count: allCount.value,
  },
  ...cats.value.map((c) => ({
    key: `c${c.id}`,
    name: c.name,
    hint: c.isGlobal ? 'หมวดส่วนกลาง' : 'หมวดของฉัน',
    icon: 'label',
    tone: 'tone-accent',
    count: c.wordCount,
  })),
])

const deckTitle = computed(() => {
  const d = deck.value
  if (!d) return ''
  if (d.kind === 'due') return 'ถึงกำหนดทบทวน'
  if (d.kind === 'all') return 'คำของฉันทั้งหมด'
  return cats.value.find((c) => c.id === d.id)?.name ?? 'หมวดหมู่'
})

async function loadDecks() {
  pickError.value = ''
  try {
    const [list, due, all] = await Promise.all([
      categories.list(),
      userWords.due(DUE_LIMIT),
      userWords.list({ limit: DECK_LIMIT }),
    ])
    cats.value = list
    dueCount.value = due.length
    allCount.value = all.items.length
  } catch (error) {
    pickError.value = messageOf(error)
  }
}

async function fetchDeck(d: Deck): Promise<UserWordDetail[]> {
  if (d.kind === 'due') return userWords.due(DUE_LIMIT)
  const categoryId = d.kind === 'category' ? d.id : undefined
  return shuffled((await userWords.list({ categoryId, limit: DECK_LIMIT })).items)
}

function choose(key: string) {
  deck.value = parseDeck(key)
  void router.replace({ query: { deck: key } })
  void start()
}

function pickAnother() {
  round++
  deck.value = null
  session.value = { phase: 'pick' }
  void router.replace({ query: {} })
  void loadDecks()
}

function advance() {
  const next = upcoming.shift()
  session.value = next ? { phase: 'front', card: next } : { phase: 'done' }
}

async function start() {
  const d = deck.value
  if (!d) return pickAnother()
  const current = ++round
  session.value = { phase: 'loading' }
  try {
    const cards = await fetchDeck(d)
    if (current !== round) return
    upcoming = cards
    total.value = cards.length
    Object.assign(tally, { again: 0, hard: 0, good: 0, easy: 0 })
    if (cards.length === 0) session.value = { phase: 'empty' }
    else advance()
  } catch (error) {
    if (current === round) session.value = { phase: 'error', message: messageOf(error) }
  }
}

function flip() {
  const s = session.value
  if (s.phase === 'front') session.value = { phase: 'back', card: s.card }
}

async function rate(key: RatingKey) {
  const s = session.value
  if (s.phase !== 'back') return
  const current = round
  tally[key]++
  advance()
  try {
    await userWords.review(s.card.id, REVIEW_QUALITY[key])
  } catch (error) {
    push(`บันทึกคะแนนของ "${s.card.word}" ไม่สำเร็จ: ${messageOf(error)}`)
    if (current !== round) return
    tally[key]--
    upcoming.push(s.card)
    if (session.value.phase === 'done') advance()
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.repeat || event.ctrlKey || event.metaKey || event.altKey) return
  const phase = session.value.phase
  if (phase === 'front' && event.code === 'Space') {
    event.preventDefault()
    flip()
  } else if (phase === 'back') {
    const rating = RATINGS.find((r) => r.hotkey === event.key)
    if (rating) {
      event.preventDefault()
      void rate(rating.key)
    }
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

void loadDecks()
if (deck.value) {
  void router.replace({ query: { deck: deckKey(deck.value) } })
  void start()
}
</script>

<style scoped>
.decks {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.deck {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  width: 100%;
  padding: var(--space-md);
  border: 2px solid var(--color-outline);
  border-radius: var(--rounded-lg);
  background: var(--color-surface);
  color: var(--color-on-surface);
  font: inherit;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 var(--edge-depth) 0 var(--edge-neutral);
  transition: transform 0.08s, box-shadow 0.08s, border-color 0.15s;
}

.deck:hover:not(:disabled) {
  border-color: var(--color-primary);
  box-shadow: 0 var(--edge-depth) 0 var(--edge-primary-container);
}

.deck:active:not(:disabled) {
  transform: translateY(var(--edge-depth));
  box-shadow: none;
}

.deck:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.deck-icon {
  display: grid;
  flex: none;
  place-items: center;
  width: 48px;
  height: 48px;
  border-radius: var(--rounded-md);
}

.tone-primary {
  background: var(--color-primary-container);
  color: var(--color-on-primary-container);
}

.tone-secondary {
  background: var(--color-secondary-container);
  color: var(--color-on-secondary-container);
}

.tone-accent {
  background: var(--color-tertiary-container);
  color: var(--color-on-tertiary-container);
}

.deck-text {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
}

.deck-name {
  overflow: hidden;
  font-weight: 800;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.deck-count {
  min-width: 40px;
  padding: 4px 12px;
  border-radius: var(--rounded-full);
  background: var(--color-surface-variant);
  color: var(--color-on-surface-variant);
  font-weight: 900;
  text-align: center;
}

.deck-chip {
  color: var(--color-primary);
  font-weight: 800;
}

.stack-sm {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.card-face {
  width: 100%;
}

button.card-face {
  border: none;
}

.back-word {
  color: var(--color-on-primary-container);
  font-size: var(--title-md-size);
  font-weight: var(--title-md-weight);
  line-height: var(--title-md-line-height);
}

.ratings,
.tallies {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-sm);
}

.ratings button {
  min-height: 64px;
}

.hotkey {
  font-family: inherit;
  font-size: var(--body-sm-size);
  font-weight: var(--body-sm-weight);
  opacity: 0.75;
}

.tally {
  cursor: default;
}

.tally-count {
  font-size: var(--headline-md-size);
  line-height: var(--headline-md-line-height);
}

.hint {
  text-align: center;
}

.flip-in {
  animation: flip-in 0.25s ease-out;
}

@keyframes flip-in {
  from {
    transform: perspective(800px) rotateY(90deg);
  }
  to {
    transform: perspective(800px) rotateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .flip-in {
    animation: none;
  }
}

@media (max-width: 639px) {
  .ratings,
  .tallies {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
