# Release {{.Version}}

### Changes
{{ range .Changes }}
### {{ .Title }}
* PR {{ .PrUrl }}
{{ range $i, $issue := .Issues }}
* Issue {{ $i }}{{ $issue }}{{ end }}
{{ end }}