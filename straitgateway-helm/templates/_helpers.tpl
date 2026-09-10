{{/*
Copyright 2026 straitgateway Authors
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "straitgateway.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "straitgateway.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart label.
*/}}
{{- define "straitgateway.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
*/}}
{{- define "straitgateway.labels" -}}
helm.sh/chart: {{ include "straitgateway.chart" . }}
{{ include "straitgateway.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: straitgateway
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "straitgateway.selectorLabels" -}}
app.kubernetes.io/name: {{ include "straitgateway.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller selector labels.
*/}}
{{- define "straitgateway.controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "straitgateway.name" . }}-controller
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: controller
{{- end }}

{{/*
Agent (DaemonSet) selector labels.
*/}}
{{- define "straitgateway.agent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "straitgateway.name" . }}-agent
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: agent
{{- end }}

{{/*
UI selector labels.
*/}}
{{- define "straitgateway.ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "straitgateway.name" . }}-ui
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: ui
{{- end }}

{{/*
Create the service account name.
*/}}
{{- define "straitgateway.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "straitgateway.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolve image tag — use values tag if set, otherwise chart appVersion.
*/}}
{{- define "straitgateway.imageTag" -}}
{{- default .Chart.AppVersion .tag }}
{{- end }}

{{/*
Controller image.
*/}}
{{- define "straitgateway.controller.image" -}}
{{- printf "%s/%s:%s" .Values.global.registry .Values.images.controller.repository (default .Chart.AppVersion .Values.images.controller.tag) }}
{{- end }}

{{/*
Agent (straitgatewayd) image.
*/}}
{{- define "straitgateway.agent.image" -}}
{{- printf "%s/%s:%s" .Values.global.registry .Values.images.daemon.repository (default .Chart.AppVersion .Values.images.daemon.tag) }}
{{- end }}

{{/*
UI image.
*/}}
{{- define "straitgateway.ui.image" -}}
{{- printf "%s/%s:%s" .Values.global.registry .Values.images.ui.repository (default .Chart.AppVersion .Values.images.ui.tag) }}
{{- end }}
