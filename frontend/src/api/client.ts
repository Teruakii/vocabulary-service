// Every call to the Go backend goes through here. Responses arrive as
// {"data": ...} on success and {"error": "..."} on failure.

export type Role = 'user' | 'teacher' | 'admin'
export type WordStatus = 'new' | 'learning' | 'mastered'
export type QuestionType = 'typed' | 'multiple_choice'
export type QuizSource = 'due' | 'all'

export interface UserResponse {
  id: number
  email: string
  role: Role
  createdAt: string
  updatedAt: string
}

export interface Identity {
  id: number
  userId: number
  provider: 'google' | 'local'
  identifier: string
  createdAt: string
  updatedAt: string
}

export interface Word {
  id: number
  word: string
  meaning: string
  // null means a global word shared by every user.
  ownerUserId: number | null
  createdAt: string
  updatedAt: string
}

export interface UserWord {
  id: number
  userId: number
  wordId: number
  status: WordStatus
  easeFactor: number
  intervalDays: number
  repetitions: number
  nextReviewAt: string | null
  lastReviewedAt: string | null
  createdAt: string
}

export interface UserWordDetail extends UserWord {
  word: string
  meaning: string
}

export interface CategoryView {
  id: number
  name: string
  isGlobal: boolean
  wordCount: number
  maxWords: number
}

export interface QuestionView {
  id: number
  wordId: number
  word: string
  questionType: QuestionType
  options?: string[]
  answered: boolean
  userAnswer?: string
  correctAnswer?: string
  isCorrect?: boolean
}

export interface QuizSummary {
  id: number
  score: number
  totalQuestions: number
  createdAt: string
  completedAt: string | null
}

export interface QuizDetail extends QuizSummary {
  questions: QuestionView[]
}

export interface UserStats {
  userId: number
  currentStreak: number
  longestStreak: number
  totalWordsLearned: number
  totalQuizTaken: number
  lastActivityAt: string | null
  updatedAt: string
}

export interface ActivityPoint {
  date: string
  count: number
}

export interface Page<T> {
  items: T[]
  hasMore: boolean
}

export const REVIEW_QUALITY = { again: 1, hard: 3, good: 4, easy: 5 } as const
export type ReviewQuality = (typeof REVIEW_QUALITY)[keyof typeof REVIEW_QUALITY]

// Every failure a page sees is an ApiError: HTTP errors carry their status,
// network failures carry NETWORK_ERROR_STATUS.
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message)
  }
}

const BASE = '/api/v1'

// The access token lives only in memory and expires after 15 minutes. The
// backend keeps the session in an httpOnly refresh cookie, which is how a
// reload or a Google sign-in gets a new access token.
let accessToken: string | null = null
let refreshing: Promise<boolean> | null = null

type Method = 'GET' | 'POST' | 'PUT' | 'DELETE'

// Status 0: the request never got an HTTP answer (server down, offline).
export const NETWORK_ERROR_STATUS = 0

async function send(path: string, method: Method, body: unknown, token: string | null) {
  const headers: Record<string, string> = {}
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`
  try {
    return await fetch(BASE + path, {
      method,
      headers,
      credentials: 'same-origin',
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch {
    throw new ApiError(NETWORK_ERROR_STATUS, 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ ลองใหม่อีกครั้ง')
  }
}

// The backend answers in English; the UI is Thai. Unknown messages pass
// through unchanged so nothing is hidden.
const thaiMessages: Record<string, string> = {
  'email already registered': 'อีเมลนี้ถูกใช้สมัครแล้ว',
  'an account with this email already exists': 'อีเมลนี้มีบัญชีอยู่แล้ว',
  'invalid username or password': 'อีเมลหรือรหัสผ่านไม่ถูกต้อง',
  'too many login attempts': 'ใส่รหัสผ่านผิดหลายครั้งเกินไป กรุณารอ 15 นาที',
  'too many password reset requests': 'ขอลิงก์รีเซ็ตบ่อยเกินไป กรุณารอสักครู่',
  'current password is incorrect': 'รหัสผ่านปัจจุบันไม่ถูกต้อง',
  'password is too short': 'รหัสผ่านต้องยาวอย่างน้อย 8 ตัวอักษร',
  'password must be at most 72 bytes': 'รหัสผ่านยาวเกินไป',
  'email must not be empty': 'กรุณากรอกอีเมล',
  'invalid or expired reset link': 'ลิงก์รีเซ็ตไม่ถูกต้องหรือหมดอายุแล้ว',
  'password reset is not available': 'ระบบส่งอีเมลยังไม่เปิดใช้งาน',
  'invalid or expired session': 'เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่',
  'missing refresh token': 'กรุณาเข้าสู่ระบบ',
  unauthorized: 'กรุณาเข้าสู่ระบบ',
  'admin only': 'เฉพาะผู้ดูแลระบบ',
  'cannot unlink your only sign-in method': 'ไม่สามารถลบวิธีเข้าสู่ระบบเดียวที่เหลืออยู่ได้',
  'identity already linked': 'บัญชีนี้เชื่อมต่อแล้ว',
  'word not found': 'ไม่พบคำนี้',
  'word already exists': 'มีคำนี้อยู่แล้ว',
  'word is in use and cannot be deleted': 'คำนี้ยังถูกใช้อยู่ จึงลบไม่ได้',
  "you don't have permission to modify this word": 'คุณไม่มีสิทธิ์แก้ไขคำนี้',
  'word must not be empty': 'กรุณากรอกคำศัพท์',
  'meaning must not be empty': 'กรุณากรอกความหมาย',
  'user word not found': 'ไม่พบคำนี้ในรายการของคุณ',
  'word already added': 'คำนี้อยู่ในรายการของคุณแล้ว',
  'you already have a word with this text': 'คุณมีคำนี้อยู่แล้ว',
  'quality must be between 0 and 5': 'คะแนนต้องอยู่ระหว่าง 0 ถึง 5',
  'category not found': 'ไม่พบหมวดหมู่นี้',
  'vocab entry not found': 'ไม่พบคำนี้ในรายการของคุณ',
  'cannot modify this category': 'คุณไม่มีสิทธิ์แก้ไขหมวดหมู่นี้',
  'word already tagged with this category': 'คำนี้อยู่ในหมวดหมู่นี้แล้ว',
  'tag not found': 'ไม่พบหมวดหมู่ของคำนี้',
  'category name must not be empty': 'กรุณากรอกชื่อหมวดหมู่',
  'no words available to quiz on': 'ยังไม่มีคำให้ทำแบบทดสอบ',
  'quiz already completed': 'แบบทดสอบนี้ทำเสร็จแล้ว',
  'question already answered': 'ตอบข้อนี้ไปแล้ว',
  'answer must not be empty': 'กรุณาใส่คำตอบ',
  'not found': 'ไม่พบข้อมูล',
  'internal server error': 'เซิร์ฟเวอร์ขัดข้อง กรุณาลองใหม่',
}

function thaiMessage(message: string): string {
  const known = thaiMessages[message]
  if (known) return known
  if (message.startsWith('this category already has')) return 'หมวดหมู่นี้มีคำครบ 20 คำแล้ว'
  // Gin binding errors ("Key: 'X' Error:Field validation ...").
  if (message.startsWith('Key:')) return 'ข้อมูลที่กรอกไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง'
  return message
}

async function parse<T>(res: Response): Promise<T> {
  if (res.status === 204) return undefined as T
  const payload = (await res.json().catch(() => null)) as { data?: T; error?: string } | null
  if (!res.ok) {
    throw new ApiError(res.status, thaiMessage(payload?.error ?? `คำขอล้มเหลว (${res.status})`))
  }
  return payload?.data as T
}

async function request<T>(path: string, method: Method = 'GET', body?: unknown): Promise<T> {
  let res = await send(path, method, body, accessToken)
  if (res.status === 401 && accessToken && (await refreshSession())) {
    res = await send(path, method, body, accessToken)
  }
  return parse<T>(res)
}

// Concurrent 401s share one refresh call.
function refreshSession(): Promise<boolean> {
  refreshing ??= (async () => {
    try {
      const { accessToken: token } = await parse<{ accessToken: string }>(
        await send('/auth/refresh', 'POST', undefined, null),
      )
      accessToken = token
      return true
    } catch {
      accessToken = null
      return false
    } finally {
      refreshing = null
    }
  })()
  return refreshing
}

// Go encodes an empty slice as null.
async function page<T>(path: string): Promise<Page<T>> {
  const res = await request<{ items: T[] | null; hasMore: boolean }>(path)
  return { items: res.items ?? [], hasMore: res.hasMore }
}

async function items<T>(path: string): Promise<T[]> {
  const res = await request<{ items: T[] | null }>(path)
  return res.items ?? []
}

function query(params: Record<string, string | number | undefined>): string {
  const q = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') q.set(key, String(value))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

export const auth = {
  async login(email: string, password: string): Promise<void> {
    const { accessToken: token } = await request<{ accessToken: string }>('/auth/login', 'POST', {
      username: email,
      password,
    })
    accessToken = token
  },

  register(email: string, password: string): Promise<UserResponse> {
    return request<UserResponse>('/auth/register', 'POST', { email, password })
  },

  async logout(): Promise<void> {
    try {
      await request('/auth/logout', 'POST')
    } finally {
      accessToken = null
    }
  },

  // True when there is a usable session, restoring it from the refresh cookie
  // after a reload or a Google sign-in.
  async hasSession(): Promise<boolean> {
    return accessToken !== null || refreshSession()
  },

  async me(): Promise<UserResponse> {
    const res = await request<{ user: UserResponse; accessToken?: string }>('/auth/me')
    if (res.accessToken) accessToken = res.accessToken
    return res.user
  },

  updateEmail(email: string): Promise<UserResponse> {
    return request<UserResponse>('/auth/me', 'PUT', { email })
  },

  async deleteAccount(): Promise<void> {
    await request('/auth/me', 'DELETE')
    accessToken = null
  },

  // The backend ends every session, so the caller must send the user to login.
  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await request('/auth/change-password', 'POST', { currentPassword, newPassword })
    accessToken = null
  },

  identities(): Promise<Identity[]> {
    return request<Identity[] | null>('/auth/identities').then((list) => list ?? [])
  },

  async forgotPassword(email: string): Promise<void> {
    await request('/auth/forgot-password', 'POST', { email })
  },

  async resetPassword(token: string, newPassword: string): Promise<void> {
    await request('/auth/reset-password', 'POST', { token, newPassword })
  },

  googleLoginUrl: `${BASE}/auth/google`,
}

export const words = {
  list(limit = 20, offset = 0): Promise<Page<Word>> {
    return page<Word>(`/words${query({ limit, offset })}`)
  },

  search(word: string): Promise<Word> {
    return request<Word>(`/words/search${query({ word })}`)
  },

  get(id: number): Promise<Word> {
    return request<Word>(`/words/${id}`)
  },

  add(word: string, meaning: string, global?: boolean): Promise<Word> {
    return request<Word>('/words', 'POST', { word, meaning, global })
  },

  update(id: number, patch: { word?: string; meaning?: string }): Promise<Word> {
    return request<Word>(`/words/${id}`, 'PUT', patch)
  },

  async remove(id: number): Promise<void> {
    await request(`/words/${id}`, 'DELETE')
  },
}

export interface UserWordFilter {
  status?: WordStatus
  categoryId?: number
  limit?: number
  offset?: number
}

export type AddUserWord = { wordId: number } | { word: string; meaning: string }

export const userWords = {
  list(filter: UserWordFilter = {}): Promise<Page<UserWordDetail>> {
    const { status, categoryId, limit = 20, offset = 0 } = filter
    return page<UserWordDetail>(`/user-words${query({ status, categoryId, limit, offset })}`)
  },

  due(limit?: number): Promise<UserWordDetail[]> {
    return items<UserWordDetail>(`/user-words/due${query({ limit })}`)
  },

  add(input: AddUserWord): Promise<UserWord> {
    return request<UserWord>('/user-words', 'POST', input)
  },

  get(id: number): Promise<UserWord> {
    return request<UserWord>(`/user-words/${id}`)
  },

  async remove(id: number): Promise<void> {
    await request(`/user-words/${id}`, 'DELETE')
  },

  review(id: number, quality: ReviewQuality): Promise<UserWord> {
    return request<UserWord>(`/user-words/${id}/review`, 'POST', { quality })
  },

  reviewByWord(wordId: number, quality: ReviewQuality): Promise<UserWord> {
    return request<UserWord>(`/user-words/by-word/${wordId}/review`, 'POST', { quality })
  },
}

export const categories = {
  list(): Promise<CategoryView[]> {
    return items<CategoryView>('/categories')
  },

  create(name: string): Promise<CategoryView> {
    return request<CategoryView>('/categories', 'POST', { name })
  },

  async remove(id: number): Promise<void> {
    await request(`/categories/${id}`, 'DELETE')
  },

  tagsOf(userWordId: number): Promise<CategoryView[]> {
    return items<CategoryView>(`/user-words/${userWordId}/tags`)
  },

  async tag(userWordId: number, categoryId: number): Promise<void> {
    await request(`/user-words/${userWordId}/tags`, 'POST', { categoryId })
  },

  async untag(userWordId: number, categoryId: number): Promise<void> {
    await request(`/user-words/${userWordId}/tags/${categoryId}`, 'DELETE')
  },
}

export const quizzes = {
  create(totalQuestions: number, source: QuizSource = 'all'): Promise<QuizDetail> {
    return request<QuizDetail>('/quizzes', 'POST', { totalQuestions, source })
  },

  list(limit = 20, offset = 0): Promise<Page<QuizSummary>> {
    return page<QuizSummary>(`/quizzes${query({ limit, offset })}`)
  },

  get(id: number): Promise<QuizDetail> {
    return request<QuizDetail>(`/quizzes/${id}`)
  },

  answer(quizId: number, questionId: number, answer: string): Promise<QuestionView> {
    return request<QuestionView>(`/quizzes/${quizId}/questions/${questionId}/answer`, 'POST', {
      answer,
    })
  },

  complete(quizId: number): Promise<QuizDetail> {
    return request<QuizDetail>(`/quizzes/${quizId}/complete`, 'POST')
  },
}

export const stats = {
  me(): Promise<UserStats> {
    return request<UserStats>('/stats/me')
  },

  activity(days = 7): Promise<ActivityPoint[]> {
    return items<ActivityPoint>(`/stats/activity${query({ days })}`)
  },
}

export function messageOf(error: unknown): string {
  return error instanceof ApiError ? error.message : 'เกิดข้อผิดพลาดที่ไม่คาดคิด'
}

export function isStatus(error: unknown, status: number): boolean {
  return error instanceof ApiError && error.status === status
}
