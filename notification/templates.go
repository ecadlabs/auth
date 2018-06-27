package notification

import (
	"text/template"
)

const (
	templatesSrc = `{{define "reset_subject"}}Reset your password{{end}}
{{define "reset_body" -}}
Reset your password

You told us you forgot your password. If you really did, click here to choose a new one:

{{.ResetURLPrefix}}{{.Token | urlquery}}

If you didn’t mean to reset your password, then you can just ignore this email; your password will not change.
{{- end}}

{{define "invite_subject"}}{{or .CurrentUser.Name .CurrentUser.Email}} has invited you to join {{.AppName}}{{end}}
{{define "invite_body" -}}
Join the {{.AppName}}

{{if .CurrentUser.Name}}{{.CurrentUser.Name}} ({{.CurrentUser.Email}}){{else}}{{.CurrentUser.Email}}{{end}} invited you to join the {{.AppName}}.

Click the link to choose a password and activate your account.

{{.ResetURLPrefix}}{{.Token| urlquery}}
{{- end}}`
)

var (
	emailTemplates = template.Must(template.New("email").Parse(templatesSrc))
)
