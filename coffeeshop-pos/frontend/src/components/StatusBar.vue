<script setup lang="ts">
defineProps<{
  lastSyncTime: string
}>()
</script>

<template>
  <footer class="status-bar">
    <div class="status-right">
      <span class="status-item text-muted text-sm">
        آخر مزامنة: {{ lastSyncTime || '—' }}
      </span>
    </div>
    <div class="status-left">
      <span class="status-item text-muted text-sm">
        نقطة البيع — المقهى v1.0
      </span>
    </div>
  </footer>
</template>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--gap-xs) var(--gap-lg);
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  height: 32px;
  flex-shrink: 0;
}

.status-right,
.status-left {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
}
</style>
