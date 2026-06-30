export function parseWailsError(err: any, fallback = 'حدث خطأ غير متوقع'): string {
  const raw = err?.message || String(err || fallback)
  try {
    const parsed = JSON.parse(raw)
    return parsed.message || raw
  } catch {
    return raw
  }
}
