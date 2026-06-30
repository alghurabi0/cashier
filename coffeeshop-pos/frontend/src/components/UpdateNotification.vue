<script setup lang="ts">
import { useUpdater } from '../composables/useUpdater'

const {
  updateAvailable,
  updateVersion,
  releaseNotes,
  isDownloading,
  downloadProgress,
  updateError,
  downloadAndInstall,
  dismiss,
} = useUpdater()
</script>

<template>
  <Transition name="slide">
    <div v-if="updateAvailable" class="update-banner">
      <div class="update-content">
        <div class="update-icon">&#x2B06;</div>
        <div class="update-info">
          <span class="update-title">إصدار جديد متاح: v{{ updateVersion }}</span>
          <span v-if="releaseNotes" class="update-notes">{{ releaseNotes }}</span>
          <span v-if="updateError" class="update-error">{{ updateError }}</span>
        </div>
        <div class="update-actions">
          <div v-if="isDownloading" class="update-progress">
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: downloadProgress + '%' }"></div>
            </div>
            <span class="progress-text">{{ downloadProgress }}%</span>
          </div>
          <template v-else>
            <button class="btn-update" @click="downloadAndInstall">تحديث</button>
            <button class="btn-dismiss" @click="dismiss">لاحقاً</button>
          </template>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.update-banner {
  background: linear-gradient(135deg, rgba(201, 168, 76, 0.15), rgba(201, 168, 76, 0.08));
  border-bottom: 1px solid var(--color-border-light);
  padding: 10px 20px;
  direction: rtl;
}

.update-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.update-icon {
  font-size: 1.3rem;
  flex-shrink: 0;
}

.update-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.update-title {
  font-size: var(--font-size-sm);
  font-weight: 700;
  color: var(--color-accent);
}

.update-notes {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.update-error {
  font-size: var(--font-size-xs);
  color: var(--color-danger, #e74c3c);
}

.update-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.btn-update {
  padding: 6px 18px;
  background: var(--color-accent);
  color: #0d0d0d;
  border: none;
  border-radius: var(--radius-sm);
  font-weight: 700;
  font-size: var(--font-size-xs);
  cursor: pointer;
  font-family: inherit;
}

.btn-update:hover {
  background: var(--color-accent-hover);
}

.btn-dismiss {
  padding: 6px 12px;
  background: transparent;
  color: var(--color-text-muted);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  cursor: pointer;
  font-family: inherit;
}

.btn-dismiss:hover {
  color: var(--color-text);
  border-color: var(--color-text-muted);
}

.update-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 140px;
}

.progress-bar {
  flex: 1;
  height: 6px;
  background: var(--color-surface-2);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-accent);
  border-radius: 3px;
  transition: width 0.3s ease;
}

.progress-text {
  font-size: var(--font-size-xs);
  color: var(--color-accent);
  font-weight: 700;
  min-width: 35px;
  text-align: left;
}

.slide-enter-active { transition: all 0.3s ease; }
.slide-leave-active { transition: all 0.2s ease; }
.slide-enter-from { transform: translateY(-100%); opacity: 0; }
.slide-leave-to { transform: translateY(-100%); opacity: 0; }
</style>
