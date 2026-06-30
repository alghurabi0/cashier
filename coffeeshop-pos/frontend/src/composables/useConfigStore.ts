import { ref } from 'vue'
import { parseWailsError } from '../utils/errors'

let ConfigStoreService: any = null

const isSetup = ref(false)
const isLoading = ref(false)
const error = ref<string | null>(null)

export function useConfigStore() {
  async function initBindings() {
    try {
      ConfigStoreService = await import('../../bindings/coffeeshop-pos/internal/service/configstoreservice')
    } catch {
      console.warn('ConfigStoreService bindings not available')
    }
  }

  async function checkSetup(): Promise<boolean> {
    if (!ConfigStoreService) return false
    try {
      const result = await ConfigStoreService.IsSetup()
      isSetup.value = result
      return result
    } catch {
      return false
    }
  }

  async function setupConnection(apiURL: string, username: string, password: string) {
    if (!ConfigStoreService) return
    error.value = null
    isLoading.value = true
    try {
      await ConfigStoreService.SetupAPIConnection(apiURL, username, password)
      isSetup.value = true
    } catch (err: any) {
      error.value = parseWailsError(err, 'فشل الاتصال بالخادم')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function provisionWithCode(apiURL: string, code: string) {
    if (!ConfigStoreService) return
    error.value = null
    isLoading.value = true
    try {
      await ConfigStoreService.ProvisionWithCode(apiURL, code)
      isSetup.value = true
    } catch (err: any) {
      error.value = parseWailsError(err, 'فشل التفعيل برمز الإعداد')
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function getConnection() {
    if (!ConfigStoreService) return null
    try {
      return await ConfigStoreService.GetAPIConnection()
    } catch {
      return null
    }
  }

  return {
    isSetup,
    isLoading,
    error,
    initBindings,
    checkSetup,
    setupConnection,
    provisionWithCode,
    getConnection,
  }
}
