<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { LocalUser } from '../../composables/useAuth'

let AuthService: any = null

const users = ref<LocalUser[]>([])
const showForm = ref(false)
const newName = ref('')
const newPin = ref('')
const newRole = ref('cashier')
const formError = ref('')
const changePinUserId = ref<string | null>(null)
const changePinValue = ref('')

async function initBindings() {
  try {
    AuthService = await import('../../../bindings/coffeeshop-pos/internal/service/authservice')
  } catch {
    console.warn('AuthService bindings not available')
  }
}

async function loadUsers() {
  if (!AuthService) return
  try {
    users.value = (await AuthService.ListUsers()) || []
  } catch (err) {
    console.error('Failed to list users:', err)
  }
}

async function createUser() {
  if (!AuthService) return
  formError.value = ''
  try {
    await AuthService.CreateUser(newName.value, newPin.value, newRole.value)
    newName.value = ''
    newPin.value = ''
    newRole.value = 'cashier'
    showForm.value = false
    await loadUsers()
  } catch (err: any) {
    formError.value = err?.message || 'فشل إنشاء المستخدم'
  }
}

async function deleteUser(id: string) {
  if (!AuthService) return
  try {
    await AuthService.DeleteUser(id)
    await loadUsers()
  } catch (err: any) {
    console.error('Failed to delete user:', err)
  }
}

async function changePin() {
  if (!AuthService || !changePinUserId.value) return
  try {
    await AuthService.ChangePin(changePinUserId.value, changePinValue.value)
    changePinUserId.value = null
    changePinValue.value = ''
  } catch (err: any) {
    console.error('Failed to change PIN:', err)
  }
}

onMounted(async () => {
  await initBindings()
  await loadUsers()
})
</script>

<template>
  <div class="user-management">
    <div class="panel-header">
      <h2 class="panel-title">👥 إدارة المستخدمين</h2>
      <button class="btn btn-sm btn-add" @click="showForm = !showForm">
        {{ showForm ? '✕ إلغاء' : '+ إضافة مستخدم' }}
      </button>
    </div>

    <!-- Add user form -->
    <div v-if="showForm" class="add-form">
      <div class="form-row">
        <input v-model="newName" type="text" placeholder="الاسم" class="form-input" />
        <input v-model="newPin" type="password" placeholder="رمز PIN" class="form-input" style="width: 100px" />
        <select v-model="newRole" class="form-input">
          <option value="cashier">كاشير</option>
          <option value="kitchen">مطبخ</option>
          <option value="admin">مدير</option>
          <option value="dev">مطور</option>
        </select>
        <button class="btn btn-primary btn-sm" @click="createUser">حفظ</button>
      </div>
      <div v-if="formError" class="form-error">{{ formError }}</div>
    </div>

    <!-- Users list -->
    <div class="users-list">
      <div v-for="user in users" :key="user.id" class="user-row">
        <span class="user-name">{{ user.name_ar }}</span>
        <span class="user-role-badge" :class="user.role">
          {{ user.role === 'admin' ? 'مدير' : user.role === 'kitchen' ? 'مطبخ' : user.role === 'dev' ? 'مطور' : 'كاشير' }}
        </span>

        <div class="user-actions">
          <button
            class="btn-icon"
            title="تغيير PIN"
            @click="changePinUserId = changePinUserId === user.id ? null : user.id"
          >🔑</button>
          <button class="btn-icon btn-danger" title="حذف" @click="deleteUser(user.id)">🗑️</button>
        </div>

        <!-- Inline change PIN -->
        <div v-if="changePinUserId === user.id" class="change-pin-row">
          <input v-model="changePinValue" type="password" placeholder="PIN جديد" class="form-input" style="width: 120px" />
          <button class="btn btn-primary btn-sm" @click="changePin">تحديث</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-management {
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
  background: var(--color-surface-2);
  padding: var(--gap-md);
  border-radius: var(--radius-md);
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

.users-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--gap-md);
  padding: var(--gap-sm) var(--gap-md);
  background: var(--color-surface);
  border-radius: var(--radius-sm);
}

.user-row .user-name {
  font-weight: var(--font-weight-bold);
  flex: 1;
}

.user-role-badge {
  padding: 2px 10px;
  border-radius: var(--radius-full);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-bold);
}

.user-role-badge.admin {
  background: rgba(212, 165, 116, 0.15);
  color: var(--color-accent);
}

.user-role-badge.cashier {
  background: rgba(92, 184, 92, 0.12);
  color: var(--color-success);
}

.user-actions {
  display: flex;
  gap: var(--gap-xs);
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

.change-pin-row {
  width: 100%;
  display: flex;
  gap: var(--gap-sm);
  align-items: center;
  margin-top: var(--gap-xs);
}

.btn-sm {
  padding: var(--gap-xs) var(--gap-md);
  font-size: var(--font-size-sm);
}

.btn-add {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-add:hover {
  border-color: var(--color-accent);
}

.form-input {
  padding: var(--gap-xs) var(--gap-sm);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  font-family: var(--font-family);
  font-size: var(--font-size-sm);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}
</style>
