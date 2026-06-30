import { ref } from 'vue'
import { parseWailsError } from '../utils/errors'

let AuthService: any = null

export interface LocalUser {
  id: string
  name_ar: string
  role: 'admin' | 'cashier' | 'kitchen' | 'dev'
}

const currentUser = ref<LocalUser | null>(null)
const isLoading = ref(false)
const error = ref<string | null>(null)

export function useAuth() {
  async function initBindings() {
    try {
      AuthService = await import('../../bindings/coffeeshop-pos/internal/service/authservice')
    } catch {
      console.warn('AuthService bindings not available')
    }
  }

  async function checkExistingSession() {
    if (!AuthService) return
    try {
      const user = await AuthService.GetCurrentUser()
      if (user) {
        currentUser.value = user
      }
    } catch {
      // No session
    }
  }

  async function listUsers(): Promise<LocalUser[]> {
    if (!AuthService) return []
    try {
      return await AuthService.ListUsers()
    } catch {
      return []
    }
  }

  async function loginUser(userID: string, pin: string) {
    if (!AuthService) return
    error.value = null
    isLoading.value = true
    try {
      const user = await AuthService.LoginUser(userID, pin)
      currentUser.value = user
    } catch (err: any) {
      error.value = parseWailsError(err, 'فشل تسجيل الدخول')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function logout() {
    if (!AuthService) return
    try {
      await AuthService.Logout()
    } catch {
      // ignore
    }
    currentUser.value = null
  }

  function isAdmin(): boolean {
    return currentUser.value?.role === 'admin'
  }

  return {
    currentUser,
    isLoading,
    error,
    initBindings,
    checkExistingSession,
    listUsers,
    loginUser,
    logout,
    isAdmin,
  }
}
