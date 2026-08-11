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

export const MEMORY_PARAM_KEYS: Record<MemoryParam, string> = {
  hugepages: 'memoryParams.hugepages',
  shmem: 'memoryParams.shmem',
  compaction_proactiveness: 'memoryParams.compaction_proactiveness',
  defrag: 'memoryParams.defrag',
  page_lock_unfairness: 'memoryParams.page_lock_unfairness',
}
