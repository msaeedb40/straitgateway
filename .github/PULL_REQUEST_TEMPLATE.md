## Description
<!-- Provide a brief description of the problem solved or feature added -->

## Architecture Alignment Check
Please verify this change respects the StraitGateway core architecture rules:
- [ ] **No CNI assumption**: Does not introduce dependencies on Calico, Cilium, or Flannel.
- [ ] **eBPF/Netkit ownership**: Adheres to Netkit pod interconnect and eBPF datapath ownership.
- [ ] **CNI Transactionality**: CNI `ADD`/`DEL`/`CHECK` operations remain idempotent and crash-recoverable.
- [ ] **Separation of Concerns**: Datapath packet forwarding does not depend on synchronous Kubernetes API (:6443) calls.
- [ ] **Least Privilege**: Capabilities are strictly constrained (no unnecessary `CAP_SYS_ADMIN`).

## Changes Proposed
- [ ] Control plane / SG Controller
- [ ] Node agent (`straitd`) / Netkit / eBPF
- [ ] CNI plugin / IPAM
- [ ] Transit Gateway (`tgwd`)
- [ ] Helm chart (`straitgateway-helm`)
- [ ] Documentation / GitHub Workflows

## Testing & Validation
- [ ] `make test` passes locally
- [ ] `helm lint straitgateway-helm/` passes
- [ ] Tested on Kind cluster (`./scripts/kind/deploy-kind.sh`)
