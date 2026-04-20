export function resolvePublicUrl(url: string, baseURL = window.location.origin): string {
  const raw = String(url || '').trim()
  if (!raw) return ''
  try {
    return new URL(raw, baseURL).toString()
  } catch {
    return raw
  }
}
