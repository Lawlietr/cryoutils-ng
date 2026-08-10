# Steam Deck 手動優化命令

以下是針對 Steam Deck 效能優化的手動命令：

## 1. 透明巨頁 (Transparent Hugepages)
```bash
echo always | sudo tee /sys/kernel/mm/transparent_hugepage/enabled
echo advise | sudo tee /sys/kernel/mm/transparent_hugepage/shmem_enabled
```

## 2. 禁用 Compaction Proactiveness
```bash
echo 0 | sudo tee /proc/sys/vm/compaction_proactiveness
```

## 3. 禁用 Hugepage Defragmentation
```bash
echo 0 | sudo tee /sys/kernel/mm/transparent_hugepage/khugepaged/defrag
```

## 4. 啟用 Page Lock Unfairness
```bash
echo 1 | sudo tee /proc/sys/vm/page_lock_unfairness
```
