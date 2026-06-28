import { ref } from 'vue'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const token = ref<string>(localStorage.getItem('admin_token') || '')

export function useApi() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('admin_token', t)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('admin_token')
  }

  function isAuthenticated(): boolean {
    return token.value !== ''
  }

  async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
    loading.value = true
    error.value = null

    try {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      }
      if (token.value) {
        headers['Authorization'] = `Bearer ${token.value}`
      }

      const resp = await fetch(`${API_BASE}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      })

      if (resp.status === 401) {
        clearToken()
        window.location.href = '/login'
        throw new Error('غير مصرح')
      }

      if (resp.status === 403) {
        throw new Error('غير مصرح — يتطلب صلاحية مدير المنصة')
      }

      const data = await resp.json()

      if (!resp.ok) {
        throw new Error(data.error || `خطأ ${resp.status}`)
      }

      return data.data as T
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const get = <T>(path: string) => request<T>('GET', path)
  const post = <T>(path: string, body: unknown) => request<T>('POST', path, body)
  const put = <T>(path: string, body: unknown) => request<T>('PUT', path, body)

  return { loading, error, get, post, put, setToken, clearToken, isAuthenticated, token }
}
