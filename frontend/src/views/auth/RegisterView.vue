<template>
  <div class="stack-lg">
    <h1>สมัครสมาชิก</h1>

    <form class="stack" @submit.prevent="submit">
      <div class="field">
        <label for="reg-email">อีเมล</label>
        <input id="reg-email" v-model.trim="email" class="input" type="email" autocomplete="email" required />
      </div>
      <div class="field">
        <label for="reg-password">รหัสผ่าน (อย่างน้อย 8 ตัวอักษร)</label>
        <input
          id="reg-password"
          v-model="password"
          class="input"
          type="password"
          autocomplete="new-password"
          minlength="8"
          required
        />
      </div>
      <div class="field">
        <label for="reg-confirm">ยืนยันรหัสผ่าน</label>
        <input
          id="reg-confirm"
          v-model="confirm"
          class="input"
          type="password"
          autocomplete="new-password"
          required
        />
      </div>
      <p v-if="error" class="alert-error" role="alert">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="pending">
        {{ pending ? 'กำลังบันทึก…' : 'สมัครสมาชิก' }}
      </button>
    </form>

    <p class="caption center">
      มีบัญชีอยู่แล้ว?
      <RouterLink :to="{ name: 'login' }">เข้าสู่ระบบ</RouterLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ApiError, auth } from '@/api/client'

const router = useRouter()

const email = ref('')
const password = ref('')
const confirm = ref('')
const pending = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
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
    await auth.register(email.value, password.value)
    await auth.login(email.value, password.value)
    await router.replace({ name: 'dashboard' })
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
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
