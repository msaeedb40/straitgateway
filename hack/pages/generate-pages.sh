#!/usr/bin/env bash
# Copyright 2026 straitgateway Authors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

CHARTS_DIR="${ROOT_DIR}/dist/charts"
mkdir -p "${CHARTS_DIR}"

echo "==> Packaging Straitgateway Helm charts..."
helm package "${ROOT_DIR}/straitgateway-helm" -d "${CHARTS_DIR}/"

echo "==> Indexing Helm repository..."
helm repo index "${CHARTS_DIR}/" --url https://msaeedb40.github.io/straitgateway/charts

echo "==> Generating dist/charts/index.html with Tailwind CSS v4..."
cat << 'EOF' > "${CHARTS_DIR}/index.html"
<!DOCTYPE html>
<html lang="en" class="dark scroll-smooth">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Straitgateway Helm Charts Repository</title>
  <meta name="description" content="Official Helm Charts repository for Straitgateway — eBPF-native Kubernetes Gateway and Inter-cluster mesh.">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
  <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
  <style type="text/tailwindcss">
    @layer base {
      body {
        font-family: 'Inter', sans-serif;
        background-color: #06080e;
        color: #e2e8f0;
      }
      code, pre { font-family: 'JetBrains Mono', monospace; }
    }
  </style>
</head>
<body class="min-h-screen flex flex-col antialiased selection:bg-indigo-500 selection:text-white">
  <header class="sticky top-0 z-50 backdrop-blur-xl bg-[#06080e]/80 border-b border-white/5">
    <div class="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
      <a href="../" class="flex items-center gap-3 group">
        <div class="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
          <svg viewBox="0 0 24 24" fill="none" class="w-4 h-4" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <span class="font-bold text-lg text-white">straitgateway <span class="text-indigo-400 text-sm font-normal">/ charts</span></span>
      </a>
      <div class="flex items-center gap-4 text-sm">
        <a href="../" class="text-slate-300 hover:text-white transition-colors">← Main Site</a>
        <a href="index.yaml" class="px-3 py-1.5 rounded-lg bg-indigo-500/20 text-indigo-300 hover:bg-indigo-500/30 transition-colors border border-indigo-500/30">index.yaml</a>
        <a href="https://github.com/msaeedb40/straitgateway" target="_blank" class="text-slate-300 hover:text-white">GitHub</a>
      </div>
    </div>
  </header>

  <main class="flex-1 max-w-6xl mx-auto px-6 py-12 w-full">
    <div class="mb-10">
      <h1 class="text-4xl font-extrabold text-white tracking-tight mb-3">Helm Charts Repository</h1>
      <p class="text-slate-400 text-base">Official packaged releases for Straitgateway eBPF-native Kubernetes networking.</p>
    </div>

    <!-- Installation Card -->
    <div class="bg-slate-900/60 border border-white/10 rounded-2xl p-6 mb-10 backdrop-blur-md">
      <h2 class="text-lg font-semibold text-white mb-3 flex items-center gap-2">
        <span class="w-2.5 h-2.5 rounded-full bg-emerald-400"></span>
        Add this Helm repository
      </h2>
      <pre class="bg-black/60 border border-white/5 rounded-xl p-4 text-sm text-emerald-300 overflow-x-auto"><code>helm repo add straitgateway https://msaeedb40.github.io/straitgateway/charts
helm repo update
helm install straitgateway straitgateway/straitgateway -n straitgateway-system --create-namespace</code></pre>
    </div>

    <!-- Available Charts -->
    <h2 class="text-2xl font-bold text-white mb-6">Available Charts</h2>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="bg-slate-900/40 border border-white/10 hover:border-indigo-500/40 rounded-2xl p-6 transition-all">
        <div class="flex items-start justify-between mb-4">
          <div>
            <h3 class="text-xl font-bold text-white">straitgateway</h3>
            <p class="text-xs text-indigo-400 mt-0.5">Application Chart · Gateway API v1.6.1</p>
          </div>
          <span class="px-2.5 py-1 text-xs font-semibold rounded-full bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">v1.0.0</span>
        </div>
        <p class="text-slate-300 text-sm mb-6 leading-relaxed">
          Full suite: CNI, kube-proxy replacement, WireGuard transit mesh, BGP routing daemon, Gateway API controller, and Angular 22 zoneless dashboard.
        </p>
        <div class="flex items-center justify-between pt-4 border-t border-white/5 text-xs">
          <span class="text-slate-400 font-mono">straitgateway-1.0.0.tgz</span>
          <a href="straitgateway-1.0.0.tgz" class="px-3 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-slate-200 transition-colors border border-white/10">Download .tgz</a>
        </div>
      </div>
    </div>
  </main>

  <footer class="border-t border-white/5 py-8 text-center text-xs text-slate-500">
    <p>© 2026 straitgateway authors. Apache 2.0 License.</p>
  </footer>
</body>
</html>
EOF

echo "==> Done generating charts repository and pages!"
