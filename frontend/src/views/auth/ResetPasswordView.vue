<template>
  <div class="stack-lg">
    <h1>ตั้งรหัสผ่านใหม่</h1>

    <template v-if="!token">
      <p class="alert-error">ลิงก์นี้ไม่สมบูรณ์ กรุณาขอลิงก์ตั้งรหัสผ่านใหม่อีกครั้ง</p>
      <RouterLink :to="{ name: 'forgot-password' }" class="btn-primary">ขอลิงก์ใหม่</RouterLink>
    </template>

    <template v-else-if="done">
      <p class="alert-success" role="status">ตั้งรหัสผ่านใหม่แล้ว เข้าสู่ระบบด้วยรหัสผ่านใหม่ได้เลย</p>
      <RouterLink :to="{ name: 'login' }" class="btn-primary">ไปหน้าเข้าสู่ระบบ</RouterLink>
    </template>

    <form v-else class="stack" @submit.prevent="submit">
      <div class="field">
        <label for="reset-password">รหัสผ่านใหม่ (อย่างน้อย 8 ตัวอักษร)</label>
        <input
          id="reset-password"
          v-model="password"
          class="input"
          type="password"
          autocomplete="new-password"
          minlength="8"
          required
        />
      </div>
      <div class="field">
        <label for="reset-confirm">ยืนยันรหัสผ่านใหม่</label>
        <input
          id="reset-confirm"
          v-model="confirm"
          class="input"
          type="password"
          autocomplete="new-password"
          required
        />
      </div>
      <p v-if="error" class="alert-error" role="alert">{{ error }}</p>
      <p v-if="expired" class="caption">
        <RouterLink :to="{ name: 'forgot-password' }">ขอลิงก์ตั้งรหัสผ่านใหม่</RouterLink>
      </p>
      <button type="submit" class="btn-primary" :disabled="pending">
        {{ pending ? 'กำลังบันทึก…' : 'ตั้งรหัสผ่านใหม่' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { ApiError, auth } from '@/api/client'

const route = useRoute()
const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''))

const password = ref('')
const confirm = ref('')
const pending = ref(false)
const error = ref('')
const expired = ref(false)
const done = ref(false)

async function submit() {
  error.value = ''
  expired.value = false
  if (password.value.length < 8) {
    error.value = 'รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร'
    return
  }
  if (password.value !== confirm.value) {
    error.value = 'รหัสผ่านทั้งสองช่องไม่ตรงกัน'
    return
  }
  pending.value = true
  try {
    await auth.resetPassword(token.value, password.value)
    done.value = true
  } catch (e) {
    if (e instanceof ApiError) {
      error.value = e.message
      expired.value = e.status === 400
    } else {
      error.value = 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
    }
  } finally {
    pending.value = false
  }
}
</script>

