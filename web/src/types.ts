export interface StatusData {
  SwapFile: string
  SwapSizeGB: string
  Swappiness: string
  VRAM: string
  HugePages: string
  ShMem: string
  CompactionProactiveness: string
  Defrag: string
  PageLockUnfairness: string
}

export interface ApiResponse<T = unknown> {
  success: boolean
  error?: string
  data?: T
}

export interface ProgressEvent {
  message: string
}

export type MemoryParam = 'hugepages' | 'shmem' | 'compaction_proactiveness' | 'defrag' | 'page_lock_unfairness'

export const MEMORY_PARAM_LABELS: Record<MemoryParam, string> = {
  hugepages: 'Huge Pages',
  shmem: 'SHMem',
  compaction_proactiveness: 'Compaction Proactiveness',
  defrag: 'Defrag',
  page_lock_unfairness: 'Page Lock Unfairness',
}
