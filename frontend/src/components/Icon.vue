<template>
  <span class="material-symbols-outlined" :style="{ fontSize: `${size}px` }" aria-hidden="true">{{ name }}</span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ name: string; size?: number }>(), { size: 24 })
</script>
