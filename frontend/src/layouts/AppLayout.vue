<template>
  <div class="app-shell">
    <aside class="sidebar">
      <RouterLink :to="{ name: 'dashboard' }" class="brand">
        <Icon name="school" />
        <span class="brand-name">LexiMaster</span>
      </RouterLink>

      <nav class="sidebar-nav" aria-label="เมนูหลัก">
        <RouterLink
          v-for="item in NAV_ITEMS"
          :key="item.name"
          :to="{ name: item.name }"
          :class="isActive(item.name) ? 'nav-item-active' : 'nav-item'"
          :title="item.label"
        >
          <Icon :name="item.icon" />
          <span class="nav-label">{{ item.label }}</span>
        </RouterLink>
      </nav>
    </aside>

    <header class="topbar">
      <RouterLink :to="{ name: 'dashboard' }" class="topbar-brand" title="LexiMaster">
        <Icon name="school" />
        <span>LexiMaster</span>
      </RouterLink>
      <span class="topbar-spacer" />

      <span v-if="streak !== null" class="streak-chip" :title="streakText">
        <Icon name="local_fire_department" :size="20" />
        <span class="streak-text">{{ streakText }}</span>
      </span>

      <div ref="menuRoot" class="settings">
        <button
          type="button"
          class="settings-trigger"
          :aria-expanded="menuOpen"
          aria-haspopup="menu"
          aria-label="การตั้งค่า"
          @click="menuOpen = !menuOpen"
        >
          <span class="avatar">{{ initial }}</span>
          <Icon name="settings" :size="22" />
        </button>

        <div v-if="menuOpen" class="settings-menu" role="menu">
          <div v-if="user" class="menu-user">
            <span class="avatar avatar-lg">{{ initial }}</span>
            <div class="menu-user-text">
              <span class="menu-email">{{ user.email }}</span>
              <span class="menu-role">{{ user.role === 'admin' ? 'ผู้ดูแลระบบ' : 'ผู้เรียน' }}</span>
            </div>
          </div>
          <RouterLink :to="{ name: 'profile' }" class="menu-item" role="menuitem">
            <Icon name="person" />
            โปรไฟล์และบัญชี
          </RouterLink>
          <hr class="menu-divider" />
          <button type="button" class="menu-item menu-item-danger" role="menuitem" @click="logout">
            <Icon name="logout" />
            ออกจากระบบ
          </button>
        </div>
      </div>
    </header>

    <main class="content">
      <RouterView />
    </main>

    <nav class="bottom-nav" aria-label="เมนูหลัก">
      <RouterLink
        v-for="item in BOTTOM_ITEMS"
        :key="item.name"
        :to="{ name: item.name }"
        class="bottom-item"
        :class="{ active: isActive(item.name) }"
      >
        <Icon :name="item.icon" />
        <span>{{ item.short ?? item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import Icon from '@/components/Icon.vue'
import { auth, stats, type UserResponse } from '@/api/client'

interface NavItem {
  name: string
  label: string
  short?: string
  icon: string
  bottom: boolean
}

const NAV_ITEMS: readonly NavItem[] = [
  { name: 'dashboard', label: 'แดชบอร์ด', icon: 'dashboard', bottom: true },
  { name: 'review', label: 'Flashcard', icon: 'style', bottom: true },
  { name: 'my-words', label: 'คำของฉัน', icon: 'bookmarks', bottom: true },
  { name: 'library', label: 'คลังคำศัพท์', short: 'คลังคำ', icon: 'local_library', bottom: true },
  { name: 'quizzes', label: 'แบบทดสอบ', icon: 'quiz', bottom: true },
  { name: 'categories', label: 'หมวดหมู่', icon: 'label', bottom: false },
]
const BOTTOM_ITEMS = NAV_ITEMS.filter((item) => item.bottom)

const route = useRoute()
const router = useRouter()
const streak = ref<number | null>(null)
const user = ref<UserResponse | null>(null)
const menuOpen = ref(false)
const menuRoot = ref<HTMLElement | null>(null)

const streakText = computed(() =>
  streak.value ? `${streak.value} วันติดต่อกัน` : 'เริ่ม streak วันนี้',
)
const initial = computed(() => user.value?.email.charAt(0).toUpperCase() ?? '?')

function isActive(name: string) {
  return route.name === name
}

async function loadStreak() {
  try {
    streak.value = (await stats.me()).currentStreak
  } catch {
    // The streak is decorative; the page itself reports real failures.
  }
}

watch(
  () => route.fullPath,
  () => {
    menuOpen.value = false
    loadStreak()
  },
  { immediate: true },
)

function onPointerDown(event: PointerEvent) {
  if (menuOpen.value && !menuRoot.value?.contains(event.target as Node)) menuOpen.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') menuOpen.value = false
}

onMounted(async () => {
  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('keydown', onKeydown)
  try {
    user.value = await auth.me()
  } catch {
    // Without the user the menu still offers profile and logout.
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onPointerDown)
  document.removeEventListener('keydown', onKeydown)
})

async function logout() {
  menuOpen.value = false
  await auth.logout().catch(() => {})
  await router.push({ name: 'login' })
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  padding-left: var(--sidebar-width);
}

.sidebar {
  position: fixed;
  inset: 0 auto 0 0;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
  width: var(--sidebar-width);
  padding: var(--space-lg);
  overflow-y: auto;
  background: var(--color-surface);
  color: var(--color-on-surface-variant);
  box-shadow: var(--shadow-card);
}

.brand {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: 40px;
  color: var(--color-primary);
  font-size: var(--title-md-size);
  font-weight: 800;
  line-height: var(--title-md-line-height);
  text-decoration: none;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 9;
  display: flex;
  align-items: center;
  gap: var(--space-md);
  height: 72px;
  padding: 0 var(--space-lg);
  background: var(--color-surface);
  border-bottom: 2px solid var(--color-outline);
}

.topbar-brand {
  display: none;
  align-items: center;
  gap: var(--space-sm);
  color: var(--color-primary);
  font-weight: 900;
  text-decoration: none;
}

.topbar-spacer {
  flex: 1;
}

.settings {
  position: relative;
}

.settings-trigger {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  height: 44px;
  padding: 0 10px 0 4px;
  border: 2px solid var(--color-outline);
  border-radius: var(--rounded-full);
  background: var(--color-surface);
  color: var(--color-on-surface-variant);
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.settings-trigger:hover,
.settings-trigger[aria-expanded="true"] {
  border-color: var(--color-primary);
  background: var(--color-primary-container);
  color: var(--color-on-primary-container);
}

.avatar {
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: var(--rounded-full);
  background: var(--color-primary);
  color: var(--color-on-primary);
  font-weight: 900;
}

.avatar-lg {
  flex: none;
  width: 44px;
  height: 44px;
  font-size: var(--title-md-size);
}

.settings-menu {
  position: absolute;
  top: calc(100% + var(--space-sm));
  right: 0;
  width: 280px;
  padding: var(--space-sm);
  border: 2px solid var(--color-outline);
  border-radius: var(--rounded-lg);
  background: var(--color-surface);
  box-shadow: var(--shadow-overlay);
}

.menu-user {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: var(--space-sm) var(--space-sm) 12px;
}

.menu-user-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.menu-email {
  overflow: hidden;
  color: var(--color-on-surface);
  font-weight: 800;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.menu-role {
  color: var(--color-muted);
  font-size: var(--body-sm-size);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: var(--rounded-md);
  background: none;
  color: var(--color-on-surface-variant);
  font: inherit;
  font-weight: 700;
  text-align: left;
  text-decoration: none;
  cursor: pointer;
}

.menu-item:hover {
  background: var(--color-surface-variant);
}

.menu-item-danger {
  color: var(--color-error);
}

.menu-item-danger:hover {
  background: var(--color-error-container);
  color: var(--color-on-error-container);
}

.menu-divider {
  margin: var(--space-xs) 0;
  border: none;
  border-top: 2px solid var(--color-outline);
}

.content {
  max-width: var(--content-max-width);
  margin: 0 auto;
  padding: var(--space-xl) var(--space-lg);
}

.bottom-nav {
  display: none;
}

@media (max-width: 1023px) {
  .app-shell {
    padding-left: var(--rail-width);
  }

  .sidebar {
    width: var(--rail-width);
    align-items: center;
    padding: var(--space-lg) var(--space-sm);
  }

  .brand-name,
  .nav-label {
    display: none;
  }

  .nav-item,
  .nav-item-active {
    justify-content: center;
  }
}

@media (max-width: 639px) {
  .app-shell {
    padding-left: 0;
    padding-bottom: 72px;
  }

  .sidebar {
    display: none;
  }

  .topbar {
    height: 64px;
    padding: 0 var(--space-md);
    gap: var(--space-sm);
  }

  .topbar-brand {
    display: flex;
  }

  .streak-text {
    display: none;
  }

  .content {
    padding: var(--space-lg) var(--space-md);
  }

  .settings-menu {
    width: min(280px, calc(100vw - 2 * var(--space-md)));
  }

  .bottom-nav {
    position: fixed;
    inset: auto 0 0 0;
    z-index: 10;
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    height: 72px;
    padding: var(--space-xs);
    background: var(--color-surface);
    box-shadow: 0 -1px 4px rgba(17, 28, 45, 0.06);
  }

  .bottom-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    border-radius: var(--rounded-md);
    color: var(--color-on-surface-variant);
    font-size: 0.75rem;
    font-weight: var(--label-md-weight);
    line-height: 1.5;
    text-decoration: none;
  }

  .bottom-item.active {
    background: var(--color-primary-container);
    color: var(--color-on-primary-container);
  }
}
</style>
