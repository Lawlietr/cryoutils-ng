export interface StatusData {
  SwapFile: string
  SwapSizeGB: string
  ZramSizeGB: string
  ZramActive: string
  TotalSwapGB: string
  Swappiness: string
  VRAM: string
  HugePages: string
  ShMem: string
  CompactionProactiveness: string
  Defrag: string
  PageLockUnfairness: string
  // The answer for each of the above, as the kernel actually holds it ("always",
  // "60", ... or "unknown"). The boolean alone could not tell recommended,
  // stock default and unreadable apart.
  HugePagesValue: string
  ShMemValue: string
  CompactionProactivenessValue: string
  DefragValue: string
  PageLockUnfairnessValue: string
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

// Where each parameter's raw value lives in the status payload.
export const MEMORY_VALUE_KEYS: Record<MemoryParam, keyof StatusData> = {
  hugepages: 'HugePagesValue',
  shmem: 'ShMemValue',
  compaction_proactiveness: 'CompactionProactivenessValue',
  defrag: 'DefragValue',
  page_lock_unfairness: 'PageLockUnfairnessValue',
}

export const MEMORY_PARAM_KEYS: Record<MemoryParam, string> = {
  hugepages: 'memoryParams.hugepages',
  shmem: 'memoryParams.shmem',
  compaction_proactiveness: 'memoryParams.compaction_proactiveness',
  defrag: 'memoryParams.defrag',
  page_lock_unfairness: 'memoryParams.page_lock_unfairness',
}
