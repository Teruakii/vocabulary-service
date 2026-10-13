---
version: alpha
name: LexiMaster
description: เว็บจำศัพท์ภาษาอังกฤษสำหรับคนไทย สดใสแบบเกม สีอินดิโก้กับเหลือง ปุ่มนูนกดได้ การ์ดใหญ่ ให้กำลังใจทุกครั้งที่ตอบ
colors:
  primary: "#4255ff"
  on-primary: "#ffffff"
  primary-hover: "#2e3ad6"
  primary-container: "#e6e9ff"
  on-primary-container: "#1b2499"
  accent: "#ffcd1f"
  on-accent: "#2b2200"
  secondary: "#7b4dff"
  on-secondary: "#ffffff"
  secondary-container: "#efe8ff"
  on-secondary-container: "#3d1a99"
  tertiary-container: "#fff0db"
  on-tertiary-container: "#8a3c00"
  success-container: "#d7f7ea"
  on-success-container: "#064d3b"
  background: "#f6f7fb"
  surface: "#ffffff"
  on-surface: "#282e3e"
  surface-variant: "#edeff4"
  on-surface-variant: "#39405a"
  muted: "#586380"
  outline: "#d9dde8"
  error: "#d1243f"
  on-error: "#ffffff"
  error-container: "#ffe3e8"
  on-error-container: "#7a0019"
typography:
  display:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 3.25rem
    fontWeight: 900
    lineHeight: 1.15
    letterSpacing: -0.01em
  headline-lg:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 2rem
    fontWeight: 800
    lineHeight: 1.25
    letterSpacing: -0.01em
  headline-md:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 1.5rem
    fontWeight: 800
    lineHeight: 1.3
  title-md:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 1.125rem
    fontWeight: 700
    lineHeight: 1.4
  body-lg:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 1.375rem
    fontWeight: 600
    lineHeight: 1.6
  body-md:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 1rem
    fontWeight: 500
    lineHeight: 1.6
  body-sm:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 0.875rem
    fontWeight: 500
    lineHeight: 1.5
  label-md:
    fontFamily: Nunito, Prompt, system-ui, sans-serif
    fontSize: 0.9375rem
    fontWeight: 800
    lineHeight: 1.2
    letterSpacing: 0.01em
rounded:
  sm: 10px
  md: 14px
  lg: 20px
  xl: 28px
  full: 9999px
spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 40px
  xxl: 64px
components:
  page:
    backgroundColor: "{colors.background}"
    textColor: "{colors.on-surface}"
    typography: "{typography.body-md}"
  sidebar:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface-variant}"
    width: 280px
    padding: 24px
  nav-item:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface-variant}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  nav-item-active:
    backgroundColor: "{colors.primary-container}"
    textColor: "{colors.on-primary-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
    textColor: "{colors.on-primary}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  button-primary-disabled:
    backgroundColor: "{colors.surface-variant}"
    textColor: "{colors.on-surface-variant}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.on-secondary}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  button-ghost:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface-variant}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  button-danger:
    backgroundColor: "{colors.error}"
    textColor: "{colors.on-error}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.lg}"
    padding: 24px
  stat-tile:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    typography: "{typography.headline-lg}"
    rounded: "{rounded.lg}"
    padding: 24px
  caption:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.muted}"
    typography: "{typography.body-sm}"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 12px
    height: 48px
  divider:
    backgroundColor: "{colors.outline}"
    height: 2px
  progress-track:
    backgroundColor: "{colors.surface-variant}"
    height: 12px
    rounded: "{rounded.full}"
  progress-fill:
    backgroundColor: "{colors.accent}"
    height: 12px
    rounded: "{rounded.full}"
  score-badge:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 8px
  flashcard-front:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    typography: "{typography.display}"
    rounded: "{rounded.xl}"
    padding: 40px
    height: 320px
  flashcard-back:
    backgroundColor: "{colors.primary-container}"
    textColor: "{colors.on-primary-container}"
    typography: "{typography.body-lg}"
    rounded: "{rounded.xl}"
    padding: 40px
    height: 320px
  rating-again:
    backgroundColor: "{colors.error-container}"
    textColor: "{colors.on-error-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  rating-hard:
    backgroundColor: "{colors.tertiary-container}"
    textColor: "{colors.on-tertiary-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  rating-good:
    backgroundColor: "{colors.primary-container}"
    textColor: "{colors.on-primary-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  rating-easy:
    backgroundColor: "{colors.success-container}"
    textColor: "{colors.on-success-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.md}"
    padding: 12px
  quiz-option:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.on-surface}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 16px
  quiz-option-correct:
    backgroundColor: "{colors.success-container}"
    textColor: "{colors.on-success-container}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 16px
  quiz-option-wrong:
    backgroundColor: "{colors.error-container}"
    textColor: "{colors.on-error-container}"
    typography: "{typography.body-md}"
    rounded: "{rounded.md}"
    padding: 16px
  badge-new:
    backgroundColor: "{colors.surface-variant}"
    textColor: "{colors.on-surface-variant}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 4px
  badge-learning:
    backgroundColor: "{colors.secondary-container}"
    textColor: "{colors.on-secondary-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 4px
  badge-mastered:
    backgroundColor: "{colors.success-container}"
    textColor: "{colors.on-success-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 4px
  streak-chip:
    backgroundColor: "{colors.tertiary-container}"
    textColor: "{colors.on-tertiary-container}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 8px
  category-chip:
    backgroundColor: "{colors.surface-variant}"
    textColor: "{colors.on-surface-variant}"
    typography: "{typography.label-md}"
    rounded: "{rounded.full}"
    padding: 8px
  alert-success:
    backgroundColor: "{colors.success-container}"
    textColor: "{colors.on-success-container}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.md}"
    padding: 12px
  alert-error:
    backgroundColor: "{colors.error-container}"
    textColor: "{colors.on-error-container}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.md}"
    padding: 12px
  toast-error:
    backgroundColor: "{colors.error}"
    textColor: "{colors.on-error}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.md}"
    padding: 12px
---

## Overview

LexiMaster เป็นเว็บจำศัพท์ภาษาอังกฤษสำหรับคนไทย ได้แรงบันดาลใจจากแอปเรียนแนวเกมอย่าง Quizlet แต่ใช้แบรนด์และหน้าตาของตัวเอง ความรู้สึกที่ต้องการคือสดใส สนุก และให้กำลังใจ ทุกการกดต้องรู้สึกว่า "ได้ทำอะไรบางอย่าง" ผ่านปุ่มนูนที่ยุบเมื่อกด สีสดที่บอกผลทันที และแถบความคืบหน้าสีเหลืองที่ขยับต่อหน้า

สร้างด้วย Vue 3 + TypeScript (`<script setup lang="ts">`) ทุกหน้าเรียก backend ผ่าน `src/api/client.ts` ตัวเดียว (Bearer token, refresh cookie, แกะ `{"data"}`, แปลข้อความ error เป็นไทย) ห้ามเรียก `fetch` ตรงจากหน้า สไตล์ทั้งหมดอยู่ใน `src/styles/tokens.css` และ `src/styles/components.css` หน้าใช้ class ตามชื่อ component ด้านล่าง

แผนผังหน้าตามเส้น API ที่มีอยู่ (base `/api/v1`):

| หน้า | เส้นทาง | API |
|---|---|---|
| เข้าสู่ระบบ / สมัคร | `/login`, `/register` | `POST /auth/login`, `POST /auth/register`, `GET /auth/google` |
| ลืม / ตั้งรหัสผ่านใหม่ | `/forgot-password`, `/reset-password?token=` | `POST /auth/forgot-password`, `POST /auth/reset-password` |
| แดชบอร์ด | `/` | `GET /stats/me`, `GET /stats/activity?days=7`, `GET /user-words/due` |
| ทบทวนการ์ด | `/review` | `GET /user-words/due`, `POST /user-words/:id/review` (quality) |
| คำของฉัน | `/my-words` | `GET/POST /user-words`, `DELETE /user-words/:id`, `GET/POST/DELETE /user-words/:id/tags` |
| คลังคำศัพท์ | `/library` | `GET /words`, `GET /words/search?word=`, `POST /words`, `PUT/DELETE /words/:id` |
| แบบทดสอบ | `/quizzes`, `/quizzes/:id` | `POST/GET /quizzes`, `POST /quizzes/:id/questions/:qid/answer`, `POST /quizzes/:id/complete` |
| หมวดหมู่ | `/categories` | `GET/POST/DELETE /categories` |
| โปรไฟล์ | `/profile` | `GET/PUT/DELETE /auth/me`, `POST /auth/change-password`, `GET /auth/identities` |

หลักการ:
- งานหลักของแต่ละหน้าต้องเห็นได้ใน 1 วินาที หน้าแดชบอร์ดมีปุ่ม "เริ่มทบทวน (N คำ)" เด่นที่สุด
- ตอบแล้วต้องรู้ผลทันทีด้วยสี (เขียว = ถูก, แดง = ผิด) ไม่ต้องรอข้อความ
- ความคืบหน้าเห็นได้เสมอ (streak, แถบเหลือง, คะแนน)
- ข้อความ UI เป็นภาษาไทย คำศัพท์ภาษาอังกฤษเป็นตัวเอกของการ์ด

## Colors

- `primary` (อินดิโก้) ใช้กับปุ่มหลักและสิ่งที่ "เลือกอยู่" `primary-container` ใช้กับเมนูที่เลือก ด้านหลังการ์ด และ rating ดี
- `accent` (เหลือง) คือสีของความคืบหน้าและรางวัล ใช้กับ `progress-fill` และ `score-badge` เท่านั้น ถ้ามีตัวอักษรบนเหลืองต้องใช้ `on-accent` ห้ามใช้ขาว
- `secondary` (ม่วง) ใช้กับการกระทำรอง เช่น "ทำแบบทดสอบ" และสถานะ "กำลังเรียน"
- `success-container` (เขียว) แปลว่า "ถูก/จำได้" เท่านั้น ใช้กับตัวเลือกที่ถูก rating ง่าย badge จำได้แล้ว และ `alert-success`
- `tertiary-container` (ส้ม) ใช้กับ streak และ rating ยาก
- `error` / `error-container` ใช้กับข้อผิดพลาด ตัวเลือกที่ผิด และ rating ลืม ห้ามใช้ตกแต่ง
- `background` คือพื้นหน้า `surface` คือพื้นการ์ด `surface-variant` คือพื้นชิปและแถบที่ยังว่าง `outline` คือเส้นขอบหนา 2px ของช่องกรอกและตัวเลือก
- ทุกปุ่มและตัวเลือกมี "ขอบล่าง" สีเข้มกว่าพื้นของมันหนึ่งระดับ หนา 4px เป็นตัวแปร `--edge-*` ใน `tokens.css`

## Typography

ใช้ Nunito สำหรับอักษรละติน (กลม หนา เป็นมิตร) และ Prompt สำหรับภาษาไทย (ทรงเรขาคณิต อ่านง่ายเมื่อหนา) โหลดจาก Google Fonts ทั้งคู่ ภาษาไทยมีสระบนล่าง จึงตั้ง `lineHeight` ของเนื้อหาไว้ที่ 1.5-1.6 น้ำหนักตัวอักษรโดยรวมหนากว่าเว็บทั่วไป เพื่อให้รู้สึกเป็นเกม

- `display` (900) ใช้กับคำศัพท์บนหน้าการ์ดและคำถามในแบบทดสอบ
- `headline-lg` ใช้กับตัวเลขใน `stat-tile` และหัวหน้าเพจ
- `headline-md` ใช้กับหัวข้อส่วน
- `title-md` ใช้กับหัวการ์ดและชื่อแบบทดสอบ
- `body-lg` ใช้กับความหมายด้านหลังการ์ด
- `body-md` ใช้กับเนื้อหาทั่วไป ช่องกรอก และตัวเลือก
- `body-sm` ใช้กับคำอธิบายและข้อความแจ้งเตือน
- `label-md` (800) ใช้กับปุ่ม badge ชิป และเมนู

## Layout

- จอกว้าง (≥ 1024px): sidebar ซ้ายกว้าง 280px + เนื้อหากว้างสูงสุด 1280px
- แท็บเล็ต (640-1023px): sidebar ย่อเหลือไอคอน 80px
- มือถือ (< 640px): ไม่มี sidebar ใช้แถบนำทางล่าง 5 ปุ่ม
- ระยะห่างใช้สเกล 4/8 จาก `spacing` เท่านั้น การ์ดเว้นระหว่างกันด้วย `lg` ภายในการ์ดใช้ `md`
- หน้าทบทวนและหน้าทำแบบทดสอบเป็นโหมดโฟกัส ซ่อนเมนู การ์ดอยู่กลางจอ กว้างสูงสุด 640px
- รายการยาวใช้ปุ่ม "โหลดเพิ่ม" ตาม `hasMore` ขนาดหน้าละ 20

## Elevation & Depth

ความลึกแบบเกมมาจาก "ขอบล่างทึบ" ไม่ใช่เงาฟุ้ง:
- ปุ่ม, rating และตัวเลือก: เงาทึบ `0 4px 0` สีขอบของตัวเอง เมื่อกดเลื่อนลง 4px และเงาหายไป เหมือนปุ่มจริงถูกกด
- การ์ด: ขอบล่างทึบ 2px สีเทาอ่อน + เงาฟุ้งเบามาก
- การ์ดคำศัพท์: ขอบล่างทึบ 4px + เงาฟุ้งเล็กน้อย ให้ลอยเด่นกว่าการ์ดทั่วไป
- modal / toast: เงาฟุ้ง `0 24px 48px` อย่างเดียว

## Shapes

- `rounded.md` สำหรับปุ่ม ช่องกรอก และตัวเลือกแบบทดสอบ
- `rounded.lg` สำหรับการ์ดทั่วไปและ stat tile
- `rounded.xl` สำหรับการ์ดคำศัพท์ให้รู้สึกเป็นการ์ดจริง
- `rounded.full` สำหรับ badge ชิป แถบความคืบหน้า และ `score-badge`
- ไอคอนใช้ Material Symbols Outlined ขนาด 20-24px

## Components

- **ปุ่ม**: หนึ่งหน้ามี `button-primary` ได้ปุ่มเดียว สูง 48px ทุกปุ่มมีขอบล่างและยุบเมื่อกด `button-ghost` มีเส้นขอบ 2px ใช้กับการกระทำรองลงมา `button-danger` ใช้กับลบ ต้องยืนยันก่อนเสมอ ขณะรอ API ใช้สถานะ disabled พร้อมข้อความ "กำลังบันทึก…"
- **flashcard**: `flashcard-front` แสดงคำศัพท์ แตะหรือกด Space เพื่อพลิกเป็น `flashcard-back` หลังพลิกจึงแสดงปุ่มให้คะแนน 4 ปุ่ม
- **rating**: `rating-again` = 1, `rating-hard` = 3, `rating-good` = 4, `rating-easy` = 5 (ค่า `quality` ของ API) ปุ่มลัด 1-4 ส่งแล้วไปการ์ดถัดไปทันที
- **progress**: `progress-track` + `progress-fill` สีเหลือง หนา 12px ใช้กับรอบทบทวน แบบทดสอบ และหมวดหมู่
- **quiz-option**: เส้นขอบ 2px + ขอบล่าง hover เปลี่ยนขอบเป็น `primary` หลังตอบใช้ `quiz-option-correct` / `quiz-option-wrong` พร้อมเฉลย ปุ่มลัด 1-4
- **score-badge**: แสดงคะแนนแบบทดสอบ เช่น "8/10" บนพื้นเหลือง
- **badge**: `new` ใช้ `badge-new` ("ใหม่"), `learning` ใช้ `badge-learning` ("กำลังเรียน"), `mastered` ใช้ `badge-mastered` ("จำได้แล้ว")
- **streak-chip**: แสดง `currentStreak` บนแดชบอร์ดและ sidebar ถ้าเป็น 0 แสดง "เริ่ม streak วันนี้"
- **stat-tile**: แดชบอร์ดมี 3 ช่อง streak, คำที่จำได้, แบบทดสอบที่ทำ
- **category-chip**: แสดง `wordCount / maxWords` (เพดาน 20 คำต่อผู้ใช้) เมื่อเต็มให้ปิดปุ่มเพิ่มและบอกเหตุผล
- **input**: เส้นขอบ 2px `outline` เมื่อ focus ขอบเป็น `primary` พร้อมวงแสงสี `primary-container`
- **alert-success / alert-error**: ข้อความผลลัพธ์ใต้ฟอร์ม `toast-error` ใช้กับข้อผิดพลาดที่ไม่ผูกกับฟอร์ม หายเองใน 4 วินาที
- **สถานะว่าง**: ทุกรายการต้องมี empty state ที่บอกขั้นต่อไป

## Do's and Don'ts

- Do ให้ปุ่มและตัวเลือกทุกตัวมีขอบล่างและยุบเมื่อกด
- Do ใช้เขียวเฉพาะความหมาย "ถูก/จำได้" และแดงเฉพาะ "ผิด/ลืม/ผิดพลาด"
- Do ใช้เหลืองกับความคืบหน้าและคะแนนเท่านั้น และใช้ `on-accent` กับตัวอักษรบนเหลือง
- Do ให้หน้าแดชบอร์ดมีปุ่มหลักปุ่มเดียวคือ "เริ่มทบทวน"
- Do รองรับแป้นพิมพ์ในหน้าทบทวนและแบบทดสอบ (Space, 1-4, Enter)
- Don't ใช้ตัวอักษรสีขาวบนเหลือง
- Don't ใช้ชื่อ โลโก้ หรือสีแบรนด์ของแอปอื่นตรงๆ ใช้แค่แนวทาง
- Don't แสดงคำศัพท์กับความหมายพร้อมกันก่อนพลิกการ์ด
- Don't เก็บ access token ใน `localStorage`
- Don't ใช้ `lineHeight` ต่ำกว่า 1.5 กับข้อความภาษาไทย
