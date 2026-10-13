<template>
  <div class="stack-lg">
    <h1>เข้าสู่ระบบ</h1>

    <p v-if="oauthError" class="alert-error">{{ oauthError }}</p>

    <form class="stack" @submit.prevent="submit">
      <div class="field">
        <label for="login-email">อีเมล</label>
        <input id="login-email" v-model.trim="email" class="input" type="email" autocomplete="email" required />
      </div>
      <div class="field">
        <div class="row-between">
          <label for="login-password" class="label">รหัสผ่าน</label>
          <RouterLink :to="{ name: 'forgot-password' }" class="caption">ลืมรหัสผ่าน?</RouterLink>
        </div>
        <input
          id="login-password"
          v-model="password"
          class="input"
          type="password"
          autocomplete="current-password"
          required
        />
      </div>
      <p v-if="error" class="alert-error" role="alert">{{ error }}</p>
      <button type="submit" class="btn-primary" :disabled="pending">
        {{ pending ? 'กำลังเข้าสู่ระบบ…' : 'เข้าสู่ระบบ' }}
      </button>
    </form>

    <div class="or caption">หรือ</div>

    <button type="button" class="btn-ghost google" :disabled="redirecting" @click="loginWithGoogle">
      <Icon name="account_circle" :size="20" />
      {{ redirecting ? 'กำลังไปที่ Google…' : 'เข้าสู่ระบบด้วย Google' }}
    </button>

    <p class="caption center">
      ยังไม่มีบัญชี?
      <RouterLink :to="{ name: 'register' }">สมัครสมาชิก</RouterLink>
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import Icon from '@/components/Icon.vue'
import { ApiError, auth } from '@/api/client'

const OAUTH_ERRORS: Record<string, string> = {
  email_not_verified: 'อีเมลของบัญชี Google นี้ยังไม่ได้ยืนยัน กรุณายืนยันอีเมลกับ Google ก่อน',
  email_exists: 'อีเมลนี้มีบัญชีอยู่แล้ว กรุณาเข้าสู่ระบบด้วยอีเมลและรหัสผ่าน',
  cancelled: 'ยกเลิกการเข้าสู่ระบบด้วย Google แล้ว',
}

const route = useRoute()
const router = useRouter()

const email = ref('')
const password = ref('')
const pending = ref(false)
const error = ref('')
const redirecting = ref(false)

const oauthError = computed(() => {
  const code = route.query.oauth_error
  if (typeof code !== 'string' || !code) return ''
  return OAUTH_ERRORS[code] ?? 'เข้าสู่ระบบด้วย Google ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง'
})

// Only same-app paths; "//host" would be read as protocol-relative by the browser.
function redirectTarget() {
  const target = route.query.redirect
  if (typeof target === 'string' && target.startsWith('/') && !target.startsWith('//')) return target
  return { name: 'dashboard' }
}

async function submit() {
  pending.value = true
  error.value = ''
  try {
    await auth.login(email.value, password.value)
    await router.replace(redirectTarget())
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
  } finally {
    pending.value = false
  }
}

function loginWithGoogle() {
  redirecting.value = true
  window.location.href = auth.googleLoginUrl
}
</script>

<style scoped>
.google {
  width: 100%;
  border: 1px solid var(--color-outline);
}

.or {
  text-align: center;
}

.center {
  text-align: center;
}
</style>
