# StraitGateway Policy Architecture & CEL Engine

## Overview
StraitGateway enforces fine-grained, label-, identity-, and expression-based network policies directly in kernel eBPF.

## Policy Compilation Pipeline
```
Kubernetes API Server (NetworkPolicy / StraitGatewayPolicy)
        │
        ▼
SG Controller (Policy Reconciler)
        │
        ▼
Policy Compiler (Common Expression Language Evaluator)
        │
        ▼
Policy Intermediate Representation (IR)
        │
        ▼
StraitD Node Agent (Policy Manager)
        │
        ▼
eBPF Policy Maps (Kernel Hash / LPM Trie)
        │
        ▼
Kernel Datapath Enforcement (TCX Ingress / Egress)
  ┌─────┴─────┐
  ▼           ▼
ALLOW       DROP
```

## Why CEL without Kernel Overhead
CEL (Common Expression Language) allows rich conditional logic (e.g. matching on custom labels, environment attributes, and protocols). Rather than attempting to run a CEL interpreter inside the eBPF virtual machine, the **Policy Compiler** evaluates CEL rules ahead-of-time against workload context and generates efficient binary lookup keys for the eBPF policy maps.
