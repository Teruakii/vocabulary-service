<template>
  <div class="auth-shell">
    <div class="auth-brand">
      <Icon name="school" :size="32" />
      <span>LexiMaster</span>
    </div>
    <main class="card auth-card">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import Icon from '@/components/Icon.vue'
</script>

<style scoped>
.auth-shell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-lg);
  min-height: 100vh;
  padding: var(--space-xl) var(--space-md);
}

.auth-brand {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  color: var(--color-primary);
  font-size: var(--headline-md-size);
  font-weight: 800;
  line-height: var(--headline-md-line-height);
}

.auth-card {
  width: 100%;
  max-width: 440px;
}
</style>
