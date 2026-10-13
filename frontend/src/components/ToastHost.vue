<template>
  <div class="toast-host" role="status" aria-live="polite">
    <div v-for="toast in toasts" :key="toast.id" class="toast-error row-between">
      <span>{{ toast.message }}</span>
      <button type="button" class="toast-close" aria-label="ปิด" @click="dismiss(toast.id)">
        <Icon name="close" :size="20" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '@/components/Icon.vue'
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()
</script>

<style scoped>
.toast-host {
  position: fixed;
  z-index: 100;
  right: var(--space-md);
  bottom: var(--space-md);
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  width: min(400px, calc(100vw - 2 * var(--space-md)));
}

.toast-close {
  display: inline-flex;
  padding: 0;
  border: none;
  background: none;
  color: inherit;
}

/* Clears the mobile bottom nav. */
@media (max-width: 639px) {
  .toast-host {
    bottom: calc(72px + var(--space-md));
  }
}
</style>
