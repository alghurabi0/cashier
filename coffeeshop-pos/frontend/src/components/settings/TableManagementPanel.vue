<script setup lang="ts">
import { ref, onMounted } from 'vue'


interface Table {
  id: string
  number: string
  token: string
  is_active: boolean
}

const tables = ref<Table[]>([])
const newNumber = ref('')
const isLoading = ref(false)
const formError = ref('')
const copiedId = ref<string | null>(null)
const qrData = ref<string | null>(null)
const qrTableNumber = ref('')

const menuBaseURL = ref('http://localhost:5173')

let ManagementService: any = null
let ConfigStoreService: any = null

async function initBindings() {
  try {
    ManagementService = await import('../../../bindings/coffeeshop-pos/internal/service/managementservice')
  } catch {
    console.warn('ManagementService bindings not available')
  }
  try {
    ConfigStoreService = await import('../../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    const savedURL = await ConfigStoreService.Get('menu_base_url')
    if (savedURL) menuBaseURL.value = savedURL
  } catch {
    console.warn('ConfigStoreService bindings not available')
  }
}

async function loadTables() {
  if (!ManagementService) return
  isLoading.value = true
  try {
    tables.value = (await ManagementService.ListTables()) || []
  } catch (err) {
    console.error('Failed to list tables:', err)
  } finally {
    isLoading.value = false
  }
}

async function createTable() {
  if (!ManagementService || !newNumber.value.trim()) return
  formError.value = ''
  try {
    await ManagementService.CreateTable(newNumber.value.trim())
    newNumber.value = ''
    await loadTables()
  } catch (err: any) {
    formError.value = err?.message || 'فشل إنشاء الطاولة'
  }
}

async function deleteTable(id: string) {
  if (!ManagementService) return
  try {
    await ManagementService.DeleteTable(id)
    await loadTables()
  } catch (err: any) {
    console.error('Failed to delete table:', err)
  }
}

function getMenuLink(token: string): string {
  return `${menuBaseURL.value}?token=${token}`
}

function copyLink(table: Table) {
  const link = getMenuLink(table.token)
  navigator.clipboard.writeText(link)
  copiedId.value = table.id
  setTimeout(() => { copiedId.value = null }, 2000)
}

async function showQR(table: Table) {
  if (!ManagementService) return
  try {
    qrTableNumber.value = table.number
    qrData.value = await ManagementService.GetTableQRCode(table.token, menuBaseURL.value)
  } catch (err) {
    console.error('Failed to generate QR code:', err)
  }
}

function downloadQR() {
  if (!qrData.value) return
  const link = document.createElement('a')
  link.href = qrData.value
  link.download = `table-${qrTableNumber.value}-qr.png`
  link.click()
}

onMounted(async () => {
  await initBindings()
  await loadTables()
})
</script>

<template>
  <div class="table-management">
    <div class="panel-header">
      <h2 class="panel-title">🪑 إدارة الطاولات</h2>
    </div>

    <!-- Add table form -->
    <div class="add-form">
      <div class="form-row">
        <input
          v-model="newNumber"
          type="text"
          placeholder="رقم الطاولة (مثال: 1, A1)"
          class="form-input"
          @keydown.enter="createTable"
        />
        <button class="btn btn-primary btn-sm" @click="createTable" :disabled="!newNumber.trim()">
          + إضافة
        </button>
      </div>
      <div v-if="formError" class="form-error">{{ formError }}</div>
    </div>

    <!-- Tables list -->
    <div v-if="isLoading" class="loading">جاري التحميل...</div>
    <div v-else-if="tables.length === 0" class="empty">لا توجد طاولات — أضف أول طاولة</div>
    <div v-else class="tables-list">
      <div v-for="table in tables" :key="table.id" class="table-row">
        <span class="table-number">🪑 {{ table.number }}</span>

        <div class="table-link">
          <code class="token-display">{{ getMenuLink(table.token) }}</code>
        </div>

        <div class="table-actions">
          <button class="btn-icon" title="رمز QR" @click="showQR(table)">📱</button>
          <button class="btn-icon" :title="copiedId === table.id ? 'تم النسخ!' : 'نسخ رابط القائمة'" @click="copyLink(table)">
            {{ copiedId === table.id ? '✅' : '📋' }}
          </button>
          <button class="btn-icon btn-danger" title="حذف" @click="deleteTable(table.id)">🗑️</button>
        </div>
      </div>
    </div>

    <!-- QR Code Modal -->
    <div v-if="qrData" class="qr-modal-overlay" @click.self="qrData = null">
      <div class="qr-modal">
        <button class="qr-close" @click="qrData = null">✕</button>
        <h3 class="qr-title">🪑 طاولة {{ qrTableNumber }}</h3>
        <img :src="qrData" alt="QR Code" class="qr-image" />
        <p class="qr-hint text-muted text-sm">امسح الرمز لفتح القائمة الإلكترونية</p>
        <button class="btn btn-primary" @click="downloadQR">📥 تحميل</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.table-management {
  display: flex;
  flex-direction: column;
  gap: var(--gap-md);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: var(--gap-sm);
}

.form-row {
  display: flex;
  gap: var(--gap-sm);
  align-items: center;
}

.form-error {
  color: var(--color-danger);
  font-size: var(--font-size-sm);
}

.tables-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.table-row {
  display: flex;
  align-items: center;
  gap: var(--gap-md);
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-sm);
}

.table-number {
  font-weight: var(--font-weight-bold);
  font-size: var(--font-size-md);
  min-width: 60px;
}

.table-link {
  flex: 1;
  overflow: hidden;
}

.token-display {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.table-actions {
  display: flex;
  gap: var(--gap-xs);
  flex-shrink: 0;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1rem;
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: background var(--transition-fast);
}

.btn-icon:hover {
  background: var(--color-surface-2);
}

.btn-danger:hover {
  background: rgba(231, 76, 60, 0.12);
}

.hint {
  font-size: var(--font-size-xs);
  color: var(--color-text-dim);
  margin-top: var(--gap-sm);
}

.form-input {
  padding: var(--gap-xs) var(--gap-sm);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
  flex: 1;
  max-width: 300px;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.btn-sm {
  padding: var(--gap-xs) var(--gap-md);
  font-size: var(--font-size-sm);
}

.btn-primary {
  background: var(--color-accent);
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-bg);
  font-family: var(--font-family);
  font-weight: var(--font-weight-bold);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-primary:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.loading, .empty {
  padding: var(--gap-lg);
  text-align: center;
  color: var(--color-text-dim);
  font-size: var(--font-size-sm);
}

/* QR Modal */
.qr-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.qr-modal {
  background: var(--color-surface);
  border-radius: var(--radius-xl, 16px);
  padding: var(--gap-xl);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--gap-md);
  position: relative;
  min-width: 300px;
}

.qr-close {
  position: absolute;
  top: 12px;
  right: 12px;
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  color: var(--color-text-muted);
}

.qr-title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
}

.qr-image {
  width: 256px;
  height: 256px;
  border-radius: var(--radius-md);
}

.qr-hint {
  text-align: center;
}
</style>
