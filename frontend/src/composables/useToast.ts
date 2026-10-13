import { reactive } from 'vue'

export interface Toast {
  id: number
  message: string
}

const DISMISS_MS = 4000
const toasts = reactive<Toast[]>([])
let nextId = 1

function dismiss(id: number) {
  const index = toasts.findIndex((t) => t.id === id)
  if (index !== -1) toasts.splice(index, 1)
}

function push(message: string) {
  const id = nextId++
  toasts.push({ id, message })
  setTimeout(() => dismiss(id), DISMISS_MS)
}

export function useToast() {
  return { toasts, push, dismiss }
}
