# vocabulary-service

Backend API ของเว็บจำศัพท์ภาษาอังกฤษ **LexiMaster** เขียนด้วย Go (Gin + GORM + MySQL)
รองรับการสมัครสมาชิก, login ด้วยรหัสผ่านหรือ Google, ทบทวนคำศัพท์แบบ spaced repetition (SM-2),
แบบทดสอบ, หมวดหมู่คำ และสถิติการเรียนรายวัน

## ฟีเจอร์

- **บัญชีผู้ใช้**: สมัคร, login, refresh token ใน httpOnly cookie, logout ที่ตัด session ฝั่ง server จริง, เปลี่ยนอีเมล/รหัสผ่าน, ลบบัญชี
- **Google OAuth**: login ด้วย Google; ถ้าอีเมลตรงกับบัญชีที่ยังไม่ยืนยัน ระบบจะยึดบัญชีให้เจ้าของ Google และตัดรหัสผ่านเดิมทิ้ง
- **ลืมรหัสผ่าน**: ส่งลิงก์รีเซ็ตทางอีเมล (token เก็บเป็น SHA-256, ใช้ได้ครั้งเดียว, หมดอายุ 1 ชั่วโมง)
- **คำศัพท์**: คลังคำส่วนกลาง (admin ดูแล) + คำส่วนตัวของแต่ละคน
- **ทบทวนแบบ SM-2**: ให้คะแนน 0-5 แล้วระบบคำนวณวันทบทวนครั้งถัดไป สถานะคำ `new` → `learning` → `mastered`
- **แบบทดสอบ**: แบบพิมพ์คำตอบและแบบเลือกตอบ 4 ตัวเลือก คำตอบป้อนกลับเข้า SM-2 อัตโนมัติ
- **หมวดหมู่**: จัดคำของตัวเองเข้าหมวด หมวดละไม่เกิน 20 คำต่อผู้ใช้
- **สถิติ**: streak รายวัน, คำที่จำได้, จำนวนแบบทดสอบ, กราฟกิจกรรมย้อนหลัง

## ความปลอดภัยที่ออกแบบไว้

- ไม่ยอม start ถ้าไม่ได้ตั้ง `JWT_SECRET_KEY`
- access token กับ refresh token แยกกันด้วย claim `typ`; refresh token ผูกกับ session ใน DB จึง revoke ได้
- เปลี่ยนรหัสผ่าน / รีเซ็ตรหัสผ่าน = ตัดทุก session
- จำกัดการเดารหัส (5 ครั้ง / 15 นาที ต่อ IP + อีเมล) และจำกัดการขอลิงก์รีเซ็ต
- สิทธิ์ admin ได้จาก `ADMIN_EMAILS` **เฉพาะเมื่ออีเมลยืนยันแล้ว** (ผ่าน Google หรือลิงก์รีเซ็ต)
- ไม่ log query string (กัน OAuth code หลุดลง log)
- ไม่ trust `X-Forwarded-For` (กันปลอม IP หลบ rate limit)

## โครงสร้างโปรเจกต์

```
cmd/main.go                  จุดเริ่มต้น: config, migrate, router, graceful shutdown
internal/config              อ่าน environment variables
internal/constants           sentinel errors, role, provider, ขนาดหน้า
internal/infra               jwt, mysql, oauth (Google), mail (SMTP), logs, response envelope
internal/middleware          RequireAuth / OptionalAuth / RequireAdmin
internal/modules/<module>    auth, word, userword, category, quiz, stats
    domain.go                  model ของ GORM + กฎของ domain
    dto.go                     request / response
    interface.go               Repository + Service interface
    service.go                 business logic (ทดสอบด้วย mock)
    repo_gorm.go               query จริงกับ MySQL
    handler.go / router.go     HTTP layer (Gin)
```

## เริ่มใช้งาน

ต้องมี Go 1.26+ และ MySQL 8 (หรือ Docker)

```bash
cp .env.example .env          # แล้วใส่ JWT_SECRET_KEY อย่างน้อย
docker compose up -d db       # MySQL ที่พอร์ต 3308
go run ./cmd                  # API ที่ http://localhost:8085
```

ตารางทั้งหมดสร้างเองด้วย GORM AutoMigrate ตอน start

รันทั้งระบบด้วย Docker:

```bash
docker compose up -d --build  # API ที่ http://localhost:8005, phpMyAdmin ที่ http://localhost:8083
```

container อ่านค่าจาก `.env` ตอนรัน (ไม่ฝังลง image) และตั้ง `TZ=Asia/Bangkok` ให้นับวันของ streak ตามเวลาไทย

ทดสอบ:

```bash
go test ./...
```

## Environment variables

| ตัวแปร | จำเป็น | ค่าเริ่มต้น | คำอธิบาย |
|---|---|---|---|
| `JWT_SECRET_KEY` | ✓ | | ใช้ sign JWT ถ้าว่าง server จะไม่ start |
| `APP_ENV` | | `production` | `development` เปิด debug และ log ลิงก์รีเซ็ตแทนการส่งอีเมล |
| `PORT` | | `8080` | พอร์ตของ API |
| `DB_HOST` `DB_PORT` `DB_USER` `DB_PASSWORD` `DB_NAME` | | `127.0.0.1` `3306` | การเชื่อมต่อ MySQL |
| `FRONTEND_URL` | | | ปลายทางหลัง Google login และลิงก์ในอีเมลรีเซ็ต |
| `ADMIN_EMAILS` | | | อีเมลที่ได้สิทธิ์ admin (คั่นด้วย comma) เมื่อยืนยันอีเมลแล้ว |
| `GOOGLE_OAUTH_CLIENT_ID` `GOOGLE_OAUTH_CLIENT_SECRET` `GOOGLE_OAUTH_REDIRECT_URL` | | | ตั้งครบสามตัวเพื่อเปิด Google login |
| `SMTP_HOST` `SMTP_PORT` `SMTP_USERNAME` `SMTP_PASSWORD` `SMTP_FROM` | | port `587` | ส่งอีเมลรีเซ็ตรหัสผ่าน (STARTTLS อัตโนมัติ) |

## API

Base path `/api/v1` ทุก response เป็น JSON:
สำเร็จ `{"data": ...}` (204 ไม่มี body), ผิดพลาด `{"error": "ข้อความ"}`
endpoint ที่มี 🔒 ต้องส่ง `Authorization: Bearer <accessToken>` (อายุ 15 นาที ขอใหม่ด้วย `POST /auth/refresh`)

### Auth `/auth`

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | `/register` | `{email, password}` สมัครสมาชิก (รหัส 8-72 bytes) |
| POST | `/login` | `{username: email, password}` → `{accessToken}` + refresh cookie |
| POST | `/refresh` | ขอ access token ใหม่จาก refresh cookie |
| POST | `/logout` | ตัด session ปัจจุบัน |
| GET | `/me` | ข้อมูลผู้ใช้ (ใช้ access token หรือ refresh cookie ก็ได้) |
| PUT / DELETE | `/me` 🔒 | เปลี่ยนอีเมล / ลบบัญชีพร้อมข้อมูลทั้งหมด |
| POST | `/change-password` 🔒 | `{currentPassword, newPassword}` |
| POST | `/forgot-password` | `{email}` ส่งลิงก์รีเซ็ต (ตอบเหมือนกันทุกกรณี) |
| POST | `/reset-password` | `{token, newPassword}` |
| GET | `/google`, `/google/callback` | Google OAuth |
| GET / POST / DELETE | `/identities` 🔒 | วิธี login ที่ผูกไว้ |
| POST | `/admin/users/:id/reset-password` 🔒 admin | ตั้งรหัสผ่านให้ผู้ใช้ |

### คำศัพท์ `/words` 🔒

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/words?limit&offset` | คำส่วนกลาง + คำของตัวเอง → `{items, hasMore}` (limit สูงสุด 100) |
| GET | `/words/search?word=` | หาคำตรงตัว |
| POST | `/words` | `{word, meaning, global?}` (`global` ใช้ได้เฉพาะ admin) |
| GET / PUT / DELETE | `/words/:id` | ลบไม่ได้ถ้ายังมีคนใช้คำนั้นอยู่ (409) |

### คำของฉัน + ทบทวน `/user-words` 🔒

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/user-words?status&categoryId&limit&offset` | รายการคำของฉันพร้อมความหมาย |
| GET | `/user-words/due?limit` | คำที่ถึงกำหนดทบทวน |
| POST | `/user-words` | `{wordId}` หรือ `{word, meaning}` (สร้างคำส่วนตัวแล้วเพิ่มให้) |
| GET / DELETE | `/user-words/:id` | |
| POST | `/user-words/:id/review` | `{quality: 0-5}` ให้คะแนนการทบทวน (SM-2) |
| POST | `/user-words/by-word/:wordId/review` | เหมือนด้านบนแต่อ้างด้วย wordId |
| GET / POST | `/user-words/:id/tags` | หมวดหมู่ของคำ / `{categoryId}` ใส่หมวด |
| DELETE | `/user-words/:id/tags/:categoryId` | เอาออกจากหมวด |

### หมวดหมู่ `/categories` 🔒

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/categories` | หมวดส่วนกลาง + ของตัวเอง พร้อม `wordCount/maxWords` |
| POST | `/categories` | `{name}` |
| DELETE | `/categories/:id` | ลบได้เฉพาะหมวดของตัวเอง |

### แบบทดสอบ `/quizzes` 🔒

| Method | Path | คำอธิบาย |
|---|---|---|
| POST | `/quizzes` | `{totalQuestions: 1-50, source: "due" \| "all"}` |
| GET | `/quizzes`, `/quizzes/:id` | ประวัติ / รายละเอียดพร้อมคำถาม |
| POST | `/quizzes/:id/questions/:qid/answer` | `{answer}` ตอบข้อสุดท้ายแล้วจบอัตโนมัติ |
| POST | `/quizzes/:id/complete` | จบแบบทดสอบ |

### สถิติ `/stats` 🔒

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/stats/me` | streak, คำที่จำได้, จำนวนแบบทดสอบ |
| GET | `/stats/activity?days=7` | จำนวนกิจกรรมรายวัน (1-90 วัน) |

`GET /health` (ไม่มี prefix) ใช้เช็กว่า server ทำงานอยู่

## Frontend

เว็บ frontend (Vue 3 + TypeScript + Vite) เรียก API นี้ผ่าน proxy `/api` ไปที่ `http://localhost:8085`
ตั้ง `FRONTEND_URL` และ `GOOGLE_OAUTH_REDIRECT_URL` ให้ชี้ไปที่ origin ของ frontend (dev: `http://localhost:5173`)
