# Tweak Explanation

## Swap Size

### What It Is

As per the [Arch Wiki](https://wiki.archlinux.org/title/swap):

```
Linux divides its physical RAM (random access memory) into chunks of memory called pages. Swapping is the process 
whereby a page of memory is copied to the preconfigured space on the hard disk, called swap space, to free up that 
page of memory. The combined sizes of the physical memory and the swap space is the amount of virtual memory available.
```

### Why It Was Changed

By increasing the swap size, the following benefits are achieved:

* Significantly reduce memory pressure
    * This allows more to be cached while simultaneously allowing for VRAM to inflate a bit more
* Have a stash of "emergency memory" in case physical memory runs low
    * This prevents bulk evictions and distributes memory management across a longer period of time, preventing latency
      spikes

### How It's Done

```bash
sudo swapoff -a
sudo dd if=/dev/zero of=/home/swapfile bs=1G count=SIZE_IN_GB status=none
sudo chmod 0600 /home/swapfile
sudo mkswap /home/swapfile  
sudo swapon /home/swapfile
```

## Swappiness

### What It Does

Also from the [Arch Wiki](https://wiki.archlinux.org/title/swap#Swappiness):

> The swappiness sysctl parameter represents the kernel's preference (or avoidance) of swap space. Swappiness can have
> a value between 0 and 200 (max 100 if Linux < 5.8), the default value is 60. A low value causes the kernel to avoid
> swapping, a high value causes the kernel to try to use swap space, and a value of 100 means IO cost is assumed to be
> equal. Using a low value on sufficient memory is known to improve responsiveness on many systems.

### Why It Was Changed

The default swappiness on the Deck is 100, which can cause data to be swapped even when significant physical memory remains.

This can can be bad for two reasons:

* Excessive writes can shorten the life of your drive
* Swap is much slower than memory, and using it slows things down

By reducing swappiness to a lower value (recommended: 1):

1. Ensure that swap is only used at the very last second, when it's really needed
2. Preserve drive health

### How It's Done

```bash
echo VALUE | sudo tee /proc/sys/vm/swappiness
```

## Transparent Hugepages

### What It Does

From an [excellent writeup by Emin here](https://xeome.github.io/notes/Transparent-Huge-Pages/):

> When the CPU assigns memory to processes that require it, it typically does so in 4 KB page chunks. Because the CPU’s
> MMU unit actively needs to translate virtual memory to physical memory upon incoming I/O requests, going through all 4
> KB pages is naturally an expensive operation. Fortunately, it has its own TLB cache (translation lookaside buffer),
> which reduces the potential amount of time required to access a specific memory address by caching the most recently
> used memory.

### Why It Was Changed

Pages are expensive to allocate. Hugepages are significantly easier to allocate and look up, and they reduce stutter when dealing with large amounts of memory.

### How It's Done

```bash
echo always | sudo tee /sys/kernel/mm/transparent_hugepage/enabled
```

## Shared Memory in Transparent Hugepages

### What It Does

As per [the kernel docs](https://www.kernel.org/doc/html/next/admin-guide/mm/transhuge.html#hugepages-in-tmpfs-shmem):

> The mount is used for SysV SHM, memfds, shared anonymous mmaps (of /dev/zero or MAP_ANONYMOUS), GPU drivers’ DRM
> objects, Ashmem.

Essentially, it allows those things to end up in hugepages.

### Why It Was Changed

For the same reasons as enabling hugepages — this can reduce latency in memory management.

### How It's Done

```bash
echo advise | sudo tee /sys/kernel/mm/transparent_hugepage/shmem_enabled
```

## Compaction Proactiveness

### What It Does

This feature proactively defragments memory when Linux detects "downtime".

### Why It Was Changed

The [kernel documentation](https://docs.kernel.org/admin-guide/sysctl/vm.html#compaction-proactiveness) confirms
that this feature has a system-wide impact on performance:

> Note that compaction has a non-trivial system-wide impact as pages belonging to different processes are moved around,
> which could also lead to latency spikes in unsuspecting applications.

Essentially, even though Linux tried to detect the proper time to do compaction, there's _never_ a good time during
gaming, so it's best to disable it.

### How It's Done

```bash
echo 0 | sudo tee /proc/sys/vm/compaction_proactiveness
```

## Hugepage Defragmentation

### What It Does

It's the same thing as proactive compaction, but for hugepages.

### Why It Was Changed

See the reasons for disabling proactive compaction above.

### How It's Done

```bash
echo 0 | sudo tee /sys/kernel/mm/transparent_hugepage/khugepaged/defrag
```

## Page Lock Unfairness

### What It Does

PLU configures how many times a process can try to get a lock on a page before "fair" behavior kicks in, and guarantees
that process access to a page.
See [the commit](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=5ef64cc8987a9211d3f3667331ba3411a94ddc79)
for details.

### Why It Was Changed

[it can have negative side effects](https://www.phoronix.com/review/linux-59-fairness), especially in gaming.
Processes waiting repeatedly can cause stutter and unexpected sleeping in games.

### How It's Done

```bash
echo 1 | sudo tee /proc/sys/vm/page_lock_unfairness
```

## Sources

Content was reviewed and refined with contributions from [Emin](https://github.com/xeome), who has a strong interest
in low-level Linux optimizations.

Additional sources include [the Arch Wiki](https://wiki.archlinux.org), [Phoronix](https://www.phoronix.com),
and [kernel documentation](https://docs.kernel.org).
