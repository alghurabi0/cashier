import { ref } from 'vue'

let ReportService: any = null

export interface ProfitReport {
  total_sales: number
  total_recipe_cost: number
  gross_profit: number
  profit_margin: number
  order_count: number
  voided_count: number
  top_items: TopSellingItem[]
  sales_by_source: Record<string, number>
  daily_breakdown: DailyEntry[]
}

export interface TopSellingItem {
  name_ar: string
  total_qty: number
  total_revenue: number
}

export interface DailyEntry {
  date: string
  order_count: number
  total_sales: number
}

function localDateStr(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const report = ref<ProfitReport | null>(null)
const isLoading = ref(false)
const dateFrom = ref(localDateStr())
const dateTo = ref(localDateStr())

export function useReports() {
  async function initBindings() {
    try {
      ReportService = await import('../../bindings/coffeeshop-pos/internal/service/reportservice')
    } catch {
      console.warn('ReportService bindings not available')
    }
  }

  async function loadReport() {
    if (!ReportService) return
    isLoading.value = true
    try {
      report.value = await ReportService.GetProfitReport(dateFrom.value, dateTo.value)
    } catch (err) {
      console.error('Failed to load report:', err)
      report.value = null
    } finally {
      isLoading.value = false
    }
  }

  return {
    report,
    isLoading,
    dateFrom,
    dateTo,
    initBindings,
    loadReport,
  }
}
