<template>
  <div class="stack-lg">
    <h1>ลืมรหัสผ่าน</h1>

    <p v-if="sent" class="alert-success" role="status">ถ้ามีบัญชีนี้ เราได้ส่งลิงก์ไปที่อีเมลแล้ว</p>

    <form v-else class="stack" @submit.prevent="submit">
      <p class="caption">กรอกอีเมลที่ใช้สมัคร เราจะส่งลิงก์สำหรับตั้งรหัสผ่านใหม่ให้</p>
      <div class="field">
        <label for="forgot-email">อีเมล</label>
        <input id="forgot-email" v-model.trim="email" class="input" type="email" autocomplete="email" required />
      </div>
      <p v-if="error" class="alert-error" role="alert">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="pending">
        {{ pending ? 'กำลังส่ง…' : 'ส่งลิงก์ตั้งรหัสผ่าน' }}
      </button>
    </form>

    <p class="caption center">
      <RouterLink :to="{ name: 'login' }">กลับไปหน้าเข้าสู่ระบบ</RouterLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ApiError, auth } from '@/api/client'

const email = ref('')
const pending = ref(false)
const error = ref('')
const sent = ref(false)

async function submit() {
  pending.value = true
  error.value = ''
  try {
    await auth.forgotPassword(email.value)
    sent.value = true
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.status === 503 ? 'ระบบส่งอีเมลยังไม่เปิดใช้งาน' : e.message
    } else {
      error.value = 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
    }
  } finally {
    pending.value = false
  }
}
</script>

<style scoped>
.center {
  text-align: center;
}
</style>
