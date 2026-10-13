import type { WordStatus } from '@/api/client'

const STATUS: Record<WordStatus, { label: string; badge: string }> = {
  new: { label: 'ใหม่', badge: 'badge-new' },
  learning: { label: 'กำลังเรียน', badge: 'badge-learning' },
  mastered: { label: 'จำได้แล้ว', badge: 'badge-mastered' },
}

export function statusLabel(status: WordStatus): string {
  return STATUS[status].label
}

export function statusClass(status: WordStatus): string {
  return STATUS[status].badge
}

const shortDate = new Intl.DateTimeFormat('th-TH', { day: 'numeric', month: 'short', year: '2-digit' })

// "12 ต.ค. 69"
export function formatDate(iso: string | null): string {
  return iso ? shortDate.format(new Date(iso)) : '-'
}

const DAY_MS = 24 * 60 * 60 * 1000

// Days are counted between calendar dates, so 23:00 today and 01:00 tomorrow is "พรุ่งนี้".
export function relativeDue(iso: string | null, now: Date = new Date()): string {
  if (!iso) return 'ยังไม่กำหนด'
  const due = new Date(iso)
  if (due.getTime() <= now.getTime()) return 'ถึงกำหนดแล้ว'
  const startOf = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()
  const days = Math.round((startOf(due) - startOf(now)) / DAY_MS)
  if (days === 0) return 'วันนี้'
  if (days === 1) return 'พรุ่งนี้'
  return `อีก ${days} วัน`
}
