{{- define "defectdojo-exporter.name" -}}
defectdojo-exporter
{{- end }}

{{- define "defectdojo-exporter.fullname" -}}
{{ include "defectdojo-exporter.name" . }}-{{ .Release.Name }}
{{- end }}

{{- define "defectdojo-exporter.labels" -}}
app.kubernetes.io/name: {{ include "defectdojo-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
