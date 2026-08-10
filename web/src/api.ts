import type { ApiResponse, ProgressEvent, StatusData } from './types'

// The token is passed via URL query string; retrieved from location.search on init.
function getToken(): string {
  const params = new URLSearchParams(window.location.search)
  return params.get('token') ?? ''
}

function buildUrl(path: string): string {
  const token = getToken()
  return `${path}?token=${encodeURIComponent(token)}`
}

export async function fetchStatus(): Promise<StatusData> {
  const res = await fetch(buildUrl('/api/status'))
  const json: ApiResponse<StatusData> = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to fetch status')
  return json.data as StatusData
}

export async function auth(password: string): Promise<void> {
  const res = await fetch(buildUrl('/api/auth'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Authentication failed')
}

export async function resizeSwap(size: number): Promise<void> {
  const res = await fetch(buildUrl('/api/swap/resize'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ size }),
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to resize swap')
}

export async function setSwappiness(value: string): Promise<void> {
  const res = await fetch(buildUrl('/api/swap/swappiness'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ value }),
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to set swappiness')
}

export async function toggleMemory(param: string): Promise<void> {
  const res = await fetch(buildUrl(`/api/memory/${param}`), {
    method: 'POST',
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? `Failed to toggle ${param}`)
}

export async function applyRecommended(): Promise<void> {
  const res = await fetch(buildUrl('/api/recommended'), { method: 'POST' })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to apply recommended settings')
}

export async function applyStock(): Promise<void> {
  const res = await fetch(buildUrl('/api/stock'), { method: 'POST' })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to apply stock settings')
}

export async function syncGameData(left: string, right: string): Promise<void> {
  const res = await fetch(buildUrl('/api/gamedata/sync'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ left, right }),
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to sync game data')
}

export async function cleanupGameData(left: string, right: string): Promise<void> {
  const res = await fetch(buildUrl('/api/gamedata/cleanup'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ left, right }),
  })
  const json: ApiResponse = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to cleanup game data')
}

export async function fetchLibraries(): Promise<string[]> {
  const res = await fetch(buildUrl('/api/libraries'))
  const json: ApiResponse<string[]> = await res.json()
  if (!json.success) throw new Error(json.error ?? 'Failed to fetch libraries')
  return json.data ?? []
}

// SSE subscription – returns an unsubscribe function.
export function subscribeProgress(onEvent: (event: ProgressEvent) => void): () => void {
  const token = getToken()
  const es = new EventSource(`/api/progress?token=${encodeURIComponent(token)}`)

  es.addEventListener('progress', (e) => {
    try {
      onEvent(JSON.parse(e.data as string))
    } catch {
      // ignore malformed events
    }
  })

  es.onerror = () => {
    // EventSource will auto-reconnect; do nothing.
  }

  return () => {
    es.close()
  }
}
