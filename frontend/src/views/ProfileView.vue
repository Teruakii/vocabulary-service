<template>
  <div class="stack-lg">
    <h1>โปรไฟล์</h1>

    <section class="card stack">
      <h3>บัญชีของฉัน</h3>
      <p v-if="!user && loadError" class="alert-error">{{ loadError }}</p>
      <p v-else-if="!user" class="caption">กำลังโหลด…</p>
      <dl v-else class="facts">
        <dt class="caption">อีเมล</dt>
        <dd>{{ user.email }}</dd>
        <dt class="caption">บทบาท</dt>
        <dd>{{ ROLE_LABEL[user.role] }}</dd>
        <dt class="caption">สมัครเมื่อ</dt>
        <dd>{{ formatDate(user.createdAt) }}</dd>
      </dl>
    </section>

    <div class="forms">
      <form class="card stack" @submit.prevent="saveEmail">
        <h3>เปลี่ยนอีเมล</h3>
        <div class="field">
          <label for="profile-email">อีเมลใหม่</label>
          <input id="profile-email" v-model.trim="newEmail" class="input" type="email" autocomplete="email" required />
        </div>
        <p class="caption">เมื่อเปลี่ยนอีเมล สถานะยืนยันอีเมลจะถูกยกเลิก และสิทธิ์ผู้ดูแลระบบ (ถ้ามี) จะถูกถอดออก</p>
        <p v-if="emailError" class="alert-error" role="alert">{{ emailError }}</p>
        <p v-if="emailSaved" class="alert-success" role="status">บันทึกอีเมลใหม่แล้ว</p>
        <button type="submit" class="btn-secondary" :disabled="emailPending || !user || newEmail === user.email">
          {{ emailPending ? 'กำลังบันทึก…' : 'บันทึกอีเมล' }}
        </button>
      </form>

      <form class="card stack" @submit.prevent="savePassword">
        <h3>เปลี่ยนรหัสผ่าน</h3>
        <div class="field">
          <label for="pw-current">รหัสผ่านปัจจุบัน</label>
          <input
            id="pw-current"
            v-model="currentPassword"
            class="input"
            type="password"
            autocomplete="current-password"
            required
          />
        </div>
        <div class="field">
          <label for="pw-new">รหัสผ่านใหม่ (อย่างน้อย 8 ตัวอักษร)</label>
          <input
            id="pw-new"
            v-model="nextPassword"
            class="input"
            type="password"
            autocomplete="new-password"
            minlength="8"
            required
          />
        </div>
        <div class="field">
          <label for="pw-confirm">ยืนยันรหัสผ่านใหม่</label>
          <input
            id="pw-confirm"
            v-model="confirmPassword"
            class="input"
            type="password"
            autocomplete="new-password"
            required
          />
        </div>
        <p v-if="passwordError" class="alert-error" role="alert">{{ passwordError }}</p>
        <p v-if="passwordChanged" class="alert-success" role="status">
          เปลี่ยนรหัสผ่านแล้ว ทุกอุปกรณ์ถูกออกจากระบบ กำลังพาไปหน้าเข้าสู่ระบบ…
        </p>
        <button type="submit" class="btn-secondary" :disabled="passwordPending || passwordChanged">
          {{ passwordPending ? 'กำลังบันทึก…' : 'เปลี่ยนรหัสผ่าน' }}
        </button>
      </form>
    </div>

    <section class="card stack">
      <h3>บัญชีที่เชื่อมต่อ</h3>
      <p v-if="!identities && loadError" class="alert-error">{{ loadError }}</p>
      <p v-else-if="!identities" class="caption">กำลังโหลด…</p>
      <div v-else-if="identities.length === 0" class="empty-state">
        <Icon name="link_off" />
        <p>ยังไม่มีบัญชีที่เชื่อมต่อ เข้าสู่ระบบด้วย Google ครั้งหน้าเพื่อเชื่อมบัญชี</p>
      </div>
      <ul v-else class="identities" role="list">
        <li v-for="identity in identities" :key="identity.id" class="row-between">
          <span class="row">
            <Icon :name="PROVIDER[identity.provider].icon" :size="20" />
            <span>{{ PROVIDER[identity.provider].label }}</span>
            <span class="caption">{{ identity.identifier }}</span>
          </span>
          <span class="caption">เชื่อมเมื่อ {{ formatDate(identity.createdAt) }}</span>
        </li>
      </ul>
    </section>

    <section class="card stack danger">
      <h3>ลบบัญชี</h3>
      <p class="caption">ลบบัญชีและข้อมูลทั้งหมดของคุณ เช่น คำศัพท์ ประวัติการทบทวน และแบบทดสอบ กู้คืนไม่ได้</p>
      <p v-if="deleteError" class="alert-error" role="alert">{{ deleteError }}</p>
      <button type="button" class="btn-danger" :disabled="deletePending" @click="removeAccount">
        {{ deletePending ? 'กำลังลบ…' : 'ลบบัญชีถาวร' }}
      </button>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import Icon from '@/components/Icon.vue'
import { ApiError, auth, type Identity, type Role, type UserResponse } from '@/api/client'
import { formatDate } from '@/utils/format'

const ROLE_LABEL: Record<Role, string> = {
  user: 'ผู้ใช้',
  teacher: 'ครู',
  admin: 'ผู้ดูแลระบบ',
}

const PROVIDER: Record<Identity['provider'], { label: string; icon: string }> = {
  google: { label: 'Google', icon: 'account_circle' },
  local: { label: 'อีเมลและรหัสผ่าน', icon: 'mail' },
}

// Long enough to read the success message before the redirect.
const REDIRECT_DELAY_MS = 2000

const router = useRouter()

function messageOf(e: unknown): string {
  return e instanceof ApiError ? e.message : 'เชื่อมต่อเซิร์ฟเวอร์ไม่ได้'
}

const user = ref<UserResponse | null>(null)
const identities = ref<Identity[] | null>(null)
const loadError = ref('')

const newEmail = ref('')
const emailPending = ref(false)
const emailError = ref('')
const emailSaved = ref(false)

const currentPassword = ref('')
const nextPassword = ref('')
const confirmPassword = ref('')
const passwordPending = ref(false)
const passwordError = ref('')
const passwordChanged = ref(false)
let redirectTimer: ReturnType<typeof setTimeout> | undefined

const deletePending = ref(false)
const deleteError = ref('')

async function load() {
  try {
    const [me, linked] = await Promise.all([auth.me(), auth.identities()])
    user.value = me
    newEmail.value = me.email
    identities.value = linked
  } catch (e) {
    loadError.value = messageOf(e)
  }
}

async function saveEmail() {
  emailPending.value = true
  emailError.value = ''
  emailSaved.value = false
  try {
    user.value = await auth.updateEmail(newEmail.value)
    newEmail.value = user.value.email
    emailSaved.value = true
  } catch (e) {
    emailError.value = messageOf(e)
  } finally {
    emailPending.value = false
  }
}

async function savePassword() {
  passwordError.value = ''
  if (nextPassword.value.length < 8) {
    passwordError.value = 'รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร'
    return
  }
  if (nextPassword.value !== confirmPassword.value) {
    passwordError.value = 'รหัสผ่านใหม่ทั้งสองช่องไม่ตรงกัน'
    return
  }
  passwordPending.value = true
  try {
    await auth.changePassword(currentPassword.value, nextPassword.value)
    passwordChanged.value = true
    redirectTimer = setTimeout(() => router.push({ name: 'login' }), REDIRECT_DELAY_MS)
  } catch (e) {
    passwordError.value = messageOf(e)
  } finally {
    passwordPending.value = false
  }
}

async function removeAccount() {
  if (!window.confirm('ลบบัญชีนี้ถาวร? ข้อมูลทั้งหมดจะหายและกู้คืนไม่ได้')) return
  deletePending.value = true
  deleteError.value = ''
  try {
    await auth.deleteAccount()
    await router.push({ name: 'login' })
  } catch (e) {
    deleteError.value = messageOf(e)
  } finally {
    deletePending.value = false
  }
}

onBeforeUnmount(() => clearTimeout(redirectTimer))

load()
</script>

<style scoped>
.facts {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  gap: var(--space-sm) var(--space-md);
  align-items: baseline;
}

.facts dd {
  overflow-wrap: anywhere;
}

.forms {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-lg);
  align-items: start;
}

.forms button {
  align-self: flex-start;
}

.identities {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.danger {
  border: 1px solid var(--color-error-container);
}

.danger button {
  align-self: flex-start;
}

@media (max-width: 1023px) {
  .forms {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
