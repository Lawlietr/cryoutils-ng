import { useState, useEffect, useCallback, useRef } from 'react'
import {
  fetchStatus,
  auth,
  resizeSwap,
  setSwappiness,
  toggleMemory,
  applyRecommended,
  applyStock,
  fetchLibraries,
  syncGameData,
  cleanupGameData,
  subscribeProgress,
} from './api'
import type { StatusData, MemoryParam } from './types'
import './App.css'

// ── Sub-components ──────────────────────────────────────────────────────────

function Header({
  sudoLocked,
  password,
  onPasswordChange,
  onAuth,
}: {
  sudoLocked: boolean
  password: string
  onPasswordChange: (v: string) => void
  onAuth: () => Promise<void>
}) {
  return (
    <header className="header">
      <h1>CryoUtils NG</h1>
      <span className="version">v2.2.2</span>
      <div className="sudo-group">
        <span className={sudoLocked ? 'sudo-locked' : 'sudo-unlocked'}>
          {sudoLocked ? '🔒 Locked' : '🔓 Unlocked'}
        </span>
        <input
          type="password"
          placeholder="sudo password"
          value={password}
          onChange={(e) => onPasswordChange(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') onAuth() }}
        />
        <button className="btn-sudo" onClick={onAuth} disabled={!password}>
          Unlock
        </button>
      </div>
    </header>
  )
}

function StatusSection({ status }: { status: StatusData }) {
  const boolFields: Array<keyof StatusData> = [
    'HugePages', 'ShMem', 'CompactionProactiveness', 'Defrag', 'PageLockUnfairness',
  ]

  return (
    <section className="section">
      <h2>System Status</h2>
      <div className="status-grid">
        <div className="status-item">
          <span className="status-label">Swap File</span>
          <span className="status-value">{status.SwapFile || 'None'}</span>
        </div>
        <div className="status-item">
          <span className="status-label">Swap Size</span>
          <span className="status-value">{status.SwapSizeGB || '?'} GB</span>
        </div>
        <div className="status-item">
          <span className="status-label">Swappiness</span>
          <span className="status-value">{status.Swappiness}</span>
        </div>
        <div className="status-item">
          <span className="status-label">VRAM</span>
          <span className="status-value readonly">{status.VRAM} (BIOS)</span>
        </div>
        {boolFields.map((key) => (
          <div className="status-item" key={key}>
            <span className="status-label">{key.replace(/_/g, ' ')}</span>
            <span className={`status-value ${status[key] === 'true' ? 'ok' : 'not-ok'}`}>
              {status[key] === 'true' ? '✓ Recommended' : '✗ Default'}
            </span>
          </div>
        ))}
      </div>
    </section>
  )
}

function SwapSection({
  status,
  sudoLocked,
  onProgress,
}: {
  status: StatusData
  sudoLocked: boolean
  onProgress: (msg: string) => void
}) {
  const [swapSizes, setSwapSizes] = useState<string[]>([])
  const [selectedSize, setSelectedSize] = useState('')
  const [swappiness, setSwappiness] = useState('60')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    // Swap sizes are derived from available space; we expose the current size
    // and let the user pick from the predefined list.
    const sizes = ['2', '4', '6', '8', '12', '16', '20', '24', '32']
    setSwapSizes(sizes)
    setSelectedSize(status.SwapSizeGB || '1')
    setSwappiness(status.Swappiness || '60')
  }, [status.SwapSizeGB, status.Swappiness])

  const handleResize = async () => {
    const size = parseInt(selectedSize, 10)
    if (isNaN(size) || size < 1) return
    setBusy(true)
    try {
      await resizeSwap(size)
      onProgress(`Swap resized to ${size} GB`)
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(false)
    }
  }

  const handleSwappiness = async () => {
    setBusy(true)
    try {
      await setSwappiness(swappiness)
      onProgress(`Swappiness set to ${swappiness}`)
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="section">
      <h2>Swap Settings</h2>
      <div className="swap-row">
        <label>Swap Size (GB)</label>
        <select
          value={selectedSize}
          onChange={(e) => setSelectedSize(e.target.value)}
          disabled={sudoLocked || busy}
        >
          {swapSizes.map((s) => (
            <option key={s} value={s}>
              {s}{s === status.SwapSizeGB ? ' (current)' : ''}
            </option>
          ))}
        </select>
        <button className="btn" onClick={handleResize} disabled={sudoLocked || busy}>
          {busy ? 'Resizing…' : 'Resize Swap'}
        </button>
      </div>
      <div className="swap-row">
        <label>Swappiness</label>
        <input
          type="number"
          min={0}
          max={200}
          value={swappiness}
          onChange={(e) => setSwappiness(e.target.value)}
          disabled={sudoLocked || busy}
          style={{ width: '80px' }}
        />
        <button className="btn" onClick={handleSwappiness} disabled={sudoLocked || busy}>
          {busy ? 'Setting…' : 'Apply'}
        </button>
      </div>
    </section>
  )
}

function MemorySection({
  status,
  sudoLocked,
  onProgress,
}: {
  status: StatusData
  sudoLocked: boolean
  onProgress: (msg: string) => void
}) {
  const [busy, setBusy] = useState<string | null>(null)

  const params: MemoryParam[] = [
    'hugepages', 'shmem', 'compaction_proactiveness', 'defrag', 'page_lock_unfairness',
  ]

  const handleToggle = async (param: MemoryParam) => {
    setBusy(param)
    try {
      await toggleMemory(param)
      onProgress(`${param} toggled`)
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(null)
    }
  }

  // Map status keys to param names
  const statusMap: Record<MemoryParam, keyof StatusData> = {
    hugepages: 'HugePages',
    shmem: 'ShMem',
    compaction_proactiveness: 'CompactionProactiveness',
    defrag: 'Defrag',
    page_lock_unfairness: 'PageLockUnfairness',
  }

  return (
    <section className="section">
      <h2>Memory Settings</h2>
      <div className="memory-list">
        {params.map((param) => {
          const key = statusMap[param]
          const isActive = status[key] === 'true'
          return (
            <div className="memory-item" key={param}>
              <span className="memory-name">{param.replace(/_/g, ' ')}</span>
              <button
                className={`btn-toggle ${isActive ? 'active' : ''}`}
                onClick={() => handleToggle(param)}
                disabled={sudoLocked || busy !== null}
              >
                {isActive ? 'ON' : 'OFF'}
              </button>
            </div>
          )
        })}
      </div>
    </section>
  )
}

function VramSection() {
  return (
    <section className="section">
      <h2>VRAM</h2>
      <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
        VRAM is read-only from this application. To change VRAM allocation,
        reboot into BIOS and adjust the setting there.
      </p>
    </section>
  )
}

function PresetSection({
  sudoLocked,
  onProgress,
}: {
  sudoLocked: boolean
  onProgress: (msg: string) => void
}) {
  const [busy, setBusy] = useState(false)

  const handleRecommended = async () => {
    setBusy(true)
    try {
      await applyRecommended()
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(false)
    }
  }

  const handleStock = async () => {
    setBusy(true)
    try {
      await applyStock()
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="section">
      <h2>Presets</h2>
      <div className="preset-row">
        <button
          className="btn-preset btn-recommended"
          onClick={handleRecommended}
          disabled={sudoLocked || busy}
        >
          {busy ? 'Applying…' : 'Recommended'}
        </button>
        <button
          className="btn-preset btn-stock"
          onClick={handleStock}
          disabled={sudoLocked || busy}
        >
          {busy ? 'Applying…' : 'Stock'}
        </button>
      </div>
    </section>
  )
}

function GameDataSection({
  sudoLocked,
  onProgress,
}: {
  sudoLocked: boolean
  onProgress: (msg: string) => void
}) {
  const [libraries, setLibraries] = useState<string[]>([])
  const [left, setLeft] = useState('')
  const [right, setRight] = useState('')
  const [busy, setBusy] = useState<string | null>(null)

  useEffect(() => {
    fetchLibraries()
      .then(setLibraries)
      .catch(() => {})
  }, [])

  const handleSync = async () => {
    if (!left || !right) return
    setBusy('sync')
    try {
      await syncGameData(left, right)
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(null)
    }
  }

  const handleCleanup = async () => {
    if (!left || !right) return
    setBusy('cleanup')
    try {
      await cleanupGameData(left, right)
    } catch (e) {
      onProgress(`Error: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(null)
    }
  }

  return (
    <section className="section">
      <h2>Game Data</h2>
      <div className="gamedata-row">
        <div className="gamedata-select-row">
          <div>
            <label>SSD Library (Source)</label>
            <select value={left} onChange={(e) => setLeft(e.target.value)} disabled={sudoLocked}>
              <option value="">— Select —</option>
              {libraries.map((l) => (
                <option key={l} value={l}>{l}</option>
              ))}
            </select>
          </div>
          <div>
            <label>External Library (Target)</label>
            <select value={right} onChange={(e) => setRight(e.target.value)} disabled={sudoLocked}>
              <option value="">— Select —</option>
              {libraries.filter((l) => !l.startsWith('/home')).map((l) => (
                <option key={l} value={l}>{l}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="gamedata-btn-row">
          <button
            className="btn-gamedata btn-sync"
            onClick={handleSync}
            disabled={sudoLocked || !left || !right || busy !== null}
          >
            {busy === 'sync' ? 'Syncing…' : 'Sync Game Data'}
          </button>
          <button
            className="btn-gamedata btn-cleanup"
            onClick={handleCleanup}
            disabled={sudoLocked || !left || !right || busy !== null}
          >
            {busy === 'cleanup' ? 'Cleaning…' : 'Cleanup Orphaned Data'}
          </button>
        </div>
      </div>
    </section>
  )
}

// ── Main App ────────────────────────────────────────────────────────────────

export default function App() {
  const [status, setStatus] = useState<StatusData>({
    SwapFile: 'Loading…',
    SwapSizeGB: '?',
    Swappiness: '?',
    VRAM: '?',
    HugePages: 'false',
    ShMem: 'false',
    CompactionProactiveness: 'false',
    Defrag: 'false',
    PageLockUnfairness: 'false',
  })
  const [sudoLocked, setSudoLocked] = useState(true)
  const [password, setPassword] = useState('')
  const [progressMsg, setProgressMsg] = useState('')
  const [busy, setBusy] = useState(false)
  const unsubscribeRef = useRef<(() => void) | null>(null)

  const refreshStatus = useCallback(async () => {
    try {
      const data = await fetchStatus()
      setStatus(data)
    } catch {
      // ignore – will retry
    }
  }, [])

  // Initial load
  useEffect(() => {
    refreshStatus()
  }, [refreshStatus])

  // SSE progress subscription
  useEffect(() => {
    unsubscribeRef.current = subscribeProgress((event) => {
      setProgressMsg(event.message)
      if (event.message.includes('complete') || event.message.includes('error') || event.message.toLowerCase().includes('error')) {
        // Refresh status after operation finishes
        setTimeout(refreshStatus, 500)
      }
    })
    return () => {
      unsubscribeRef.current?.()
      unsubscribeRef.current = null
    }
  }, [refreshStatus])

  const handleAuth = async () => {
    setBusy(true)
    try {
      await auth(password)
      setSudoLocked(false)
    } catch (e) {
      setProgressMsg(`Auth failed: ${e instanceof Error ? e.message : String(e)}`)
    } finally {
      setBusy(false)
      setPassword('')
    }
  }

  return (
    <>
      {progressMsg && <div className="progress-message">{progressMsg}</div>}
      <Header
        sudoLocked={sudoLocked}
        password={password}
        onPasswordChange={setPassword}
        onAuth={handleAuth}
      />
      <StatusSection status={status} />
      <SwapSection status={status} sudoLocked={sudoLocked} onProgress={(msg) => setProgressMsg(msg)} />
      <MemorySection status={status} sudoLocked={sudoLocked} onProgress={(msg) => setProgressMsg(msg)} />
      <VramSection />
      <PresetSection sudoLocked={sudoLocked} onProgress={(msg) => setProgressMsg(msg)} />
      <GameDataSection sudoLocked={sudoLocked} onProgress={(msg) => setProgressMsg(msg)} />
    </>
  )
}
