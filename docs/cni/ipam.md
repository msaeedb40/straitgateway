# Lockless Bitmap IPAM Allocator

StraitGateway includes a high-performance IP Address Management (IPAM) engine designed to handle thousands of container churn operations per second without database lock contention.

---

## 1. Algorithmic Foundation

The allocator resides in `internal/cni/ipam/` and uses atomic 64-bit integer bitmaps:

* **Subnet Representation**: Each node PodCIDR (e.g. `/24` yielding 256 IPs) is represented as a contiguous array of `uint64` words (for `/24`, exactly four 64-bit words).
* **Allocation Operation**:
  1. Finds the first word with non-zero inverse bits (`^word != 0`).
  2. Uses hardware bitwise instruction (`bits.TrailingZeros64(^word)`) to locate the first free IP index in $O(1)$ constant time.
  3. Uses atomic Compare-And-Swap (`sync/atomic` CAS) or spin-guarded bitmap update to reserve the slot without global lock blocking.
* **Deallocation Operation**:
  * Clears the bit and updates the node allocator statistics atomically.

---

## 2. Dual-Stack Support (IPv4 & IPv6)

StraitGateway supports single-stack IPv4, single-stack IPv6, and dual-stack IPv4/IPv6 deployments:

```json
{
  "cniVersion": "0.4.0",
  "name": "straitgateway",
  "type": "sg-cni",
  "ipam": {
    "type": "sg-ipam",
    "ranges": [
      [{"subnet": "10.244.1.0/24"}],
      [{"subnet": "fd00:10:244:1::/64"}]
    ]
  }
}
```

* **IPv4 Allocations**: Assigned with `/32` netmask directly to the container `eth0` device.
* **IPv6 Allocations**: Assigned with `/128` netmask.

---

## 3. Crash Recovery & Garbage Collection

To prevent leaked IP addresses when containers terminate abruptly:
1. **Persistent State Cache**: The node agent writes allocated IPs to a lightweight journal in `/var/run/straitgateway/ipam.state`.
2. **Periodic Reconciliation**: Every 30 seconds, `straitd` cross-references active container namespaces against the IPAM state map and frees any orphaned allocations.
