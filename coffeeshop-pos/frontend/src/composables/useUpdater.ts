import { ref } from 'vue'

let UpdateService: any = null

const updateAvailable = ref(false)
const updateVersion = ref('')
const releaseNotes = ref('')
const downloadURL = ref('')
const isDownloading = ref(false)
const downloadProgress = ref(0)
const updateError = ref('')

let checkInterval: ReturnType<typeof setInterval> | null = null
let progressPoll: ReturnType<typeof setInterval> | null = null

export function useUpdater() {
  async function initBindings() {
    try {
      UpdateService = await import('../../bindings/coffeeshop-pos/internal/service/updateservice')
    } catch { /* dev mode */ }
  }

  async function checkForUpdate() {
    if (!UpdateService) return
    try {
      const info = await UpdateService.CheckForUpdate()
      if (info?.available) {
        updateAvailable.value = true
        updateVersion.value = info.version
        releaseNotes.value = info.release_notes
        downloadURL.value = info.download_url
      }
    } catch { /* ignore network errors */ }
  }

  function startPeriodicCheck() {
    setTimeout(checkForUpdate, 10_000)
    checkInterval = setInterval(checkForUpdate, 4 * 60 * 60 * 1000)
  }

  async function downloadAndInstall() {
    if (!UpdateService || !downloadURL.value) return
    isDownloading.value = true
    downloadProgress.value = 0
    updateError.value = ''

    progressPoll = setInterval(async () => {
      try {
        const p = await UpdateService.GetDownloadProgress()
        downloadProgress.value = Math.round(p * 100)
      } catch {}
    }, 500)

    try {
      const path = await UpdateService.DownloadUpdate(downloadURL.value)
      clearInterval(progressPoll!)
      downloadProgress.value = 100
      await UpdateService.ApplyUpdate(path)
    } catch (err: any) {
      clearInterval(progressPoll!)
      isDownloading.value = false
      updateError.value = err?.message || 'فشل التحديث'
    }
  }

  function dismiss() {
    updateAvailable.value = false
  }

  function cleanup() {
    if (checkInterval) clearInterval(checkInterval)
    if (progressPoll) clearInterval(progressPoll)
  }

  return {
    updateAvailable,
    updateVersion,
    releaseNotes,
    downloadURL,
    isDownloading,
    downloadProgress,
    updateError,
    initBindings,
    checkForUpdate,
    startPeriodicCheck,
    downloadAndInstall,
    dismiss,
    cleanup,
  }
}
