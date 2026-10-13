# Foundation for page agents

Replace only your own view file under `src/views/`. Do not edit the files listed here. If you need a change to them, report it instead.

## Rules

- Pages never call `fetch`. Every backend call goes through `@/api/client`.
- Show `ApiError.message` from a failed call. Form errors go in `.alert-error` under the form. Errors not tied to a form go to a toast.
- One `.btn-primary` per page. While a request is pending, disable it and show "กำลังบันทึก…".
- Danger actions (`.btn-danger`) need a confirm step first.
- Every list has an `.empty-state` that tells the user the next step.
- Long lists page by 20 with a "โหลดเพิ่ม" button driven by `hasMore`.
- Use the tokens in `src/styles/tokens.css` (`var(--space-md)`, `var(--color-muted)`, ...) in scoped styles. Do not hard-code colors.

## API (`src/api/client.ts`)

All ids are numbers. Paged lists return `Page<T> = { items: T[], hasMore }` and `items` is never null. Unpaged lists return `T[]`.

- `auth`: `login(email, password)`, `register(email, password)`, `logout()`, `hasSession()`, `me()`, `updateEmail(email)`, `deleteAccount()`, `changePassword(current, next)` (clears the token, so send the user to `login` afterwards), `identities()`, `forgotPassword(email)`, `resetPassword(token, newPassword)`, `googleLoginUrl` (assign to `window.location.href`).
- `words`: `list(limit?, offset?)`, `search(word)`, `get(id)`, `add(word, meaning, global?)`, `update(id, { word?, meaning? })`, `remove(id)`.
- `userWords`: `list({ status?, categoryId?, limit?, offset? })`, `due(limit?)`, `add({ wordId } | { word, meaning })`, `get(id)`, `remove(id)`, `review(id, quality)`, `reviewByWord(wordId, quality)`.
- `categories`: `list()`, `create(name)`, `remove(id)`, `tagsOf(userWordId)`, `tag(userWordId, categoryId)`, `untag(userWordId, categoryId)`.
- `quizzes`: `create(totalQuestions, source?)`, `list(limit?, offset?)`, `get(id)`, `answer(quizId, questionId, answer)`, `complete(quizId)`.
- `stats`: `me()`, `activity(days?)`.
- `REVIEW_QUALITY = { again: 1, hard: 3, good: 4, easy: 5 }`. `review` only accepts these values.
- Types: `UserResponse`, `Identity`, `Word`, `UserWord`, `UserWordDetail`, `CategoryView`, `QuestionView`, `QuizSummary`, `QuizDetail`, `UserStats`, `ActivityPoint`, `Page`, `WordStatus`, `QuestionType`, `QuizSource`, `ReviewQuality`, `ApiError`.

## Formatting (`src/utils/format.ts`)

`statusLabel(status)` gives "ใหม่" / "กำลังเรียน" / "จำได้แล้ว". `statusClass(status)` gives the badge class. `formatDate(iso)` gives a Thai short date. `relativeDue(iso)` gives "ถึงกำหนดแล้ว", "วันนี้", "พรุ่งนี้" or "อีก N วัน".

## Toasts

```ts
import { useToast } from '@/composables/useToast'
const { push } = useToast()
push('เชื่อมต่อเซิร์ฟเวอร์ไม่ได้')
```

`ToastHost` is already mounted in `App.vue`. Toasts dismiss after 4 seconds.

## Icons

`<Icon name="delete" />` from `@/components/Icon.vue` renders a Material Symbols Outlined glyph. Optional `:size="20"`.

## CSS classes (`src/styles/components.css`)

- Buttons: `.btn-primary`, `.btn-secondary`, `.btn-danger`, `.btn-ghost` (tertiary actions).
- Surfaces: `.card`, `.stat-tile`, `.caption`, `.divider` (use on `<hr>`), `.empty-state`.
- Forms: wrap each control in `.field` with a `<label>` above an `.input`. `.label` styles a standalone label.
- Progress: `.progress-track` containing `.progress-fill` with an inline `width`.
- Review: `.flashcard-front`, `.flashcard-back`, `.rating-again`, `.rating-hard`, `.rating-good`, `.rating-easy`.
- Quiz: `.quiz-option`, `.quiz-option-correct`, `.quiz-option-wrong`.
- Status: `.badge-new`, `.badge-learning`, `.badge-mastered`, `.streak-chip`, `.category-chip`.
- Feedback: `.alert-error`, `.toast-error`.
- Navigation: `.nav-item`, `.nav-item-active`.
- Layout: `.stack` (column, gap md), `.stack-lg` (column, gap lg), `.row` (wrapping row, gap sm), `.row-between`, `.grid-3` (one column under 640px).

Base styles already set `h1` to headline-lg, `h2` to headline-md and `h3` to title-md.

## Layouts and routes

| Name | Path | Layout | View file |
|---|---|---|---|
| `login` | `/login` | AuthLayout (centred card) | `views/auth/LoginView.vue` |
| `register` | `/register` | AuthLayout | `views/auth/RegisterView.vue` |
| `forgot-password` | `/forgot-password` | AuthLayout | `views/auth/ForgotPasswordView.vue` |
| `reset-password` | `/reset-password?token=` | AuthLayout | `views/auth/ResetPasswordView.vue` |
| `dashboard` | `/` | AppLayout (sidebar, rail, bottom nav) | `views/DashboardView.vue` |
| `my-words` | `/my-words` | AppLayout | `views/MyWordsView.vue` |
| `library` | `/library` | AppLayout | `views/LibraryView.vue` |
| `quizzes` | `/quizzes` | AppLayout | `views/QuizListView.vue` |
| `categories` | `/categories` | AppLayout | `views/CategoriesView.vue` |
| `profile` | `/profile` | AppLayout | `views/ProfileView.vue` |
| `review` | `/review` | FocusLayout (640px column, close link) | `views/ReviewView.vue` |
| `quiz-play` | `/quizzes/:id` (digits) | FocusLayout | `views/QuizPlayView.vue` |

Auth pages carry `meta.guest`. The others carry `meta.requiresAuth`. A signed-out visit to a protected page goes to `login` with `?redirect=<path>`, so the login page should `router.replace(redirect ?? { name: 'dashboard' })` after signing in. The layouts already provide page padding and the max width, so views render only their content.
