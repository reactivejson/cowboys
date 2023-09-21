{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "player.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "playerfullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
player.chartref prints a chart name and version.
It does minimal escaping for use in Kubernetes labels.
*/ -}}
{{- define "player.chartref" -}}
  {{- replace "+" "_" .Chart.Version | printf "%s-%s" .Chart.Name -}}
{{- end -}}

{{/*
player.labels.standard prints the standard Helm labels.
The standard labels are frequently used in metadata.
*/}}
{{- define "player.labels.standard" -}}
app: {{template "player.name" .}}
chart: {{template "player.chartref" . }}
app.kubernetes.io/name: {{template "player.name" .}}
helm.sh/chart: {{template "player.chartref" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.Version }}
com.nokia.neo/commitId: ${COMMIT_ID}
{{- end -}}

{{/*
player.template.labels prints the template metadata labels.
*/}}
{{- define "player.template.labels" -}}
app: {{template "player.name" .}}
{{- end -}}

{{- define "player.app" -}}
app: {{template "player.name" .}}
{{- end -}}

{{- define "annotateResources" -}}
# Preserve the workingsetpluginregistrations.ws.nokia.com crd for changes;
kubectl annotate --overwrite crd workingsetpluginregistrations.ws.nokia.com "helm.sh/resource-policy"=keep;
sleep 30;
{{- end -}}
