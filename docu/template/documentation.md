# {{ .Service.Name }} service

{{ if .VersionName }}**Version:** {{ .VersionName }}  {{ end }}
**Version time:** {{ .VersionTime }}

{{ .Service.Description }}

## Configuration

Path prefix: `{{ if .Service.HttpPrefix }}{{ .Service.HttpPrefix }}{{ else }}/{{ end }}`  
Compression enabled: {{ if .Service.CompressionEnabled }}yes{{ else }}no{{ end }}  
{{ if .Service.WSPath }} WebSocket supported, path: {{ .Service.WSPath }}{{ end }}

## Methods
{{ range $m := .Service.Methods }}
### {{ $m.Name }}
{{ range $b := $m.HttpBindings }}
**`{{ $b.Method | ToUpper }}`** `{{ $b.PathRaw }}`  
{{ end }}
<details>

{{ $m.Method.Comments | NormalizeComments }}
{{ if (index $.Messages $m.Request).Fields }}
#### Parameters
> | name | required | type | description | additional information |
> |---|---|---|---|---|{{ range $p := (index $.Messages $m.Request).Fields }}
> | {{ $p.Name }} | {{ if $p.Optional }}no{{ else }}yes{{ end }} | {{ $p.Type | Escape }} | {{ $p.Description }} | {{ range $o := $p.Options }}`{{ $o | Escape }}`<br/>{{ end }} | {{ end }}
{{ end }}
{{ if (index $.Messages $m.Response).Fields }}
#### Response
> | name | required | type | description | additional information |
> |---|---|---|---|---|{{ range $p := (index $.Messages $m.Response).Fields }}
> | {{ $p.Name }} | {{ if $p.Optional }}no{{ else }}yes{{ end }} | {{ $p.Type | Escape }} | {{ $p.Description }} | {{ range $o := $p.Options }}`{{ $o | Escape }}`<br/>{{ end }} | {{ end }}
> {{ end }}
</details>
{{ end }}

## Objects
{{ range $name := .Referenced }}{{ $msg := (index $.Messages $name) }}{{ if $msg.Name }}
### {{ $msg.Name }}
{{ $msg.Description }}

> | name | required | type | description | additional information |
> |---|---|---|---|---|{{ range $p := $msg.Fields }}
> | {{ $p.Name }} | {{ if $p.Optional }}no{{ else }}yes{{ end }} | {{ $p.Type | Escape }} | {{ $p.Description }} | {{ range $o := $p.Options }}`{{ $o | Escape }}`<br/>{{ end }} | {{ end }}
{{ end }}{{ end }}