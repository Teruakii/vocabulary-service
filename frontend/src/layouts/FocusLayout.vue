<template>
  <div class="focus-shell">
    <header class="focus-header">
      <RouterLink :to="{ name: 'dashboard' }" class="btn-ghost" aria-label="ปิดและกลับแดชบอร์ด">
        <Icon name="close" />
        <span>ปิด</span>
      </RouterLink>
    </header>
    <main class="focus-main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import Icon from '@/components/Icon.vue'
</script>

<style scoped>
.focus-shell {
  min-height: 100vh;
  max-width: var(--focus-max-width);
  margin: 0 auto;
  padding: var(--space-md);
}

.focus-header {
  display: flex;
  justify-content: flex-start;
  margin-bottom: var(--space-lg);
}

.focus-header .btn-ghost {
  background: transparent;
}
</style>
