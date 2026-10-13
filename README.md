# vocabulary-service

Backend API for **LexiMaster**, a web app for learning English vocabulary. Written in Go with Gin, GORM and MySQL.

It handles sign-up, login with a password or Google, spaced-repetition review (SM-2), quizzes, word categories and daily learning stats.

## Features

- **Accounts**: sign up, log in, refresh token in an httpOnly cookie, logout that revokes the session on the server, change email or password, delete account.
- **Google OAuth**: log in with Google. If the email matches an unverified account, the Google owner takes the account over and its old password is dropped.
- **Password reset**: a reset link sent by email. The token is stored as a SHA-256 hash, works once, and expires after 1 hour.
- **Words**: a shared library maintained by admins, plus private words for each user.
- **SM-2 review**: rate a word 0-5 and the server schedules its next review. A word moves through `new` → `learning` → `mastered`.
- **Quizzes**: typed-answer and 4-option multiple-choice questions. Answers feed back into SM-2.
- **Categories**: group your words. Each category holds up to 20 words per user.
- **Stats**: daily streak, words mastered, quizzes taken, and recent activity.

## Security

- The server refuses to start without `JWT_SECRET_KEY`.
- Access and refresh tokens are told apart by a `typ` claim. Refresh tokens are bound to a session row, so they can be revoked.
- Changing or resetting a password revokes every session.
- Login attempts are rate limited (5 per 15 minutes per IP and email), and so are reset-link requests.
- `ADMIN_EMAILS` grants admin **only once the email is verified** (through Google or a reset link).
- Query strings are not logged, so OAuth codes never reach the logs.
- `X-Forwarded-For` is not trusted, so a client cannot fake its IP to dodge rate limits.

## Project layout

```
cmd/main.go                  entry point: config, migrations, router, graceful shutdown
internal/config              environment variables
internal/constants           sentinel errors, roles, providers, page sizes
internal/infra               jwt, mysql, oauth (Google), mail (SMTP), logs, response envelope
internal/middleware          RequireAuth / OptionalAuth / RequireAdmin
internal/modules/<module>    auth, word, userword, category, quiz, stats
    domain.go                  GORM models and domain rules
    dto.go                     requests and responses
    interface.go               Repository and Service interfaces
    service.go                 business logic (tested with mocks)
    repo_gorm.go               MySQL queries
    handler.go / router.go     HTTP layer (Gin)
```

## Getting started

Requires Go 1.26+ and MySQL 8 (or Docker).

```bash
cp .env.example .env          # set at least JWT_SECRET_KEY
docker compose up -d db       # MySQL on port 3308
go run ./cmd                  # API on http://localhost:8085
```

GORM AutoMigrate creates every table on startup.

To run the whole stack in Docker:

```bash
docker compose up -d --build  # API on http://localhost:8005, phpMyAdmin on http://localhost:8083
```

The container reads `.env` at runtime (nothing is baked into the image) and sets `TZ=Asia/Bangkok`, so streak days follow Thai time.

Tests:

```bash
go test ./...
```

## Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `JWT_SECRET_KEY` | ✓ | | Signs JWTs. The server will not start without it. |
| `APP_ENV` | | `production` | `development` enables debug mode and logs reset links instead of emailing them. |
| `PORT` | | `8080` | API port. |
| `DB_HOST` `DB_PORT` `DB_USER` `DB_PASSWORD` `DB_NAME` | | `127.0.0.1` `3306` | MySQL connection. |
| `FRONTEND_URL` | | | Where the browser lands after Google login, and the base of reset links. |
| `ADMIN_EMAILS` | | | Comma-separated emails that become admin once verified. |
| `GOOGLE_OAUTH_CLIENT_ID` `GOOGLE_OAUTH_CLIENT_SECRET` `GOOGLE_OAUTH_REDIRECT_URL` | | | Set all three to enable Google login. |
| `SMTP_HOST` `SMTP_PORT` `SMTP_USERNAME` `SMTP_PASSWORD` `SMTP_FROM` | | port `587` | Sends password reset emails (STARTTLS when offered). |

## API

Every path is under `/api/v1` and every response is JSON:
`{"data": ...}` on success (204 has no body), `{"error": "message"}` on failure.
Endpoints marked 🔒 need `Authorization: Bearer <accessToken>`. Access tokens last 15 minutes; get a new one with `POST /auth/refresh`.

### Auth `/auth`

| Method | Path | Description |
|---|---|---|
| POST | `/register` | `{email, password}` creates an account (password 8-72 bytes) |
| POST | `/login` | `{username: email, password}` → `{accessToken}` plus the refresh cookie |
| POST | `/refresh` | New access token from the refresh cookie |
| POST | `/logout` | Revokes the current session |
| GET | `/me` | Current user (works with an access token or the refresh cookie) |
| PUT / DELETE | `/me` 🔒 | Change email / delete the account and all its data |
| POST | `/change-password` 🔒 | `{currentPassword, newPassword}` |
| POST | `/forgot-password` | `{email}` sends a reset link (same response either way) |
| POST | `/reset-password` | `{token, newPassword}` |
| GET | `/google`, `/google/callback` | Google OAuth |
| GET / POST / DELETE | `/identities` 🔒 | Linked login methods |
| POST | `/admin/users/:id/reset-password` 🔒 admin | Set a user's password |

### Words `/words` 🔒

| Method | Path | Description |
|---|---|---|
| GET | `/words?limit&offset` | Shared words plus your own → `{items, hasMore}` (limit up to 100) |
| GET | `/words/search?word=` | Exact match lookup |
| POST | `/words` | `{word, meaning, global?}` (`global` is admin only) |
| GET / PUT / DELETE | `/words/:id` | Delete returns 409 while anyone still uses the word |

### My words and review `/user-words` 🔒

| Method | Path | Description |
|---|---|---|
| GET | `/user-words?status&categoryId&limit&offset` | Your words with meanings |
| GET | `/user-words/due?limit` | Words due for review |
| POST | `/user-words` | `{wordId}`, or `{word, meaning}` to create a private word and add it |
| GET / DELETE | `/user-words/:id` | |
| POST | `/user-words/:id/review` | `{quality: 0-5}` rates a review (SM-2) |
| POST | `/user-words/by-word/:wordId/review` | Same, addressed by wordId |
| GET / POST | `/user-words/:id/tags` | The word's categories / `{categoryId}` adds one |
| DELETE | `/user-words/:id/tags/:categoryId` | Removes the word from a category |

### Categories `/categories` 🔒

| Method | Path | Description |
|---|---|---|
| GET | `/categories` | Shared and your own categories with `wordCount/maxWords` |
| POST | `/categories` | `{name}` |
| DELETE | `/categories/:id` | Only your own categories |

### Quizzes `/quizzes` 🔒

| Method | Path | Description |
|---|---|---|
| POST | `/quizzes` | `{totalQuestions: 1-50, source: "due" \| "all"}` |
| GET | `/quizzes`, `/quizzes/:id` | History / one quiz with its questions |
| POST | `/quizzes/:id/questions/:qid/answer` | `{answer}`; answering the last question completes the quiz |
| POST | `/quizzes/:id/complete` | Completes the quiz |

### Stats `/stats` 🔒

| Method | Path | Description |
|---|---|---|
| GET | `/stats/me` | Streak, words mastered, quizzes taken |
| GET | `/stats/activity?days=7` | Daily activity counts (1-90 days) |

`GET /health` (no prefix) reports whether the server is up.

## Connecting a frontend

A web client calls this API under `/api/v1`. In development, proxy `/api` to `http://localhost:8085`.
Point `FRONTEND_URL` and `GOOGLE_OAUTH_REDIRECT_URL` at the frontend's origin (in development, `http://localhost:5173`) so Google login and reset links return to the web app.
