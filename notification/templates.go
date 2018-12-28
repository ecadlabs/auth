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
{{- end}}

{{define "email_update_request_subject"}}Confirmation required to update email to {{.Email}} on {{.AppName}}{{end}}
{{define "email_update_request_body" -}}
Hello {{.TargetUser.Name}}

You have requested to update your email address on {{.AppName}} to {{.Email}}

To approve this change, you must click this link: {{.UpdateEmailURLPrefix}}{{.Token| urlquery}}

If you did not make this change, you should contact {{.SupportEmail}} immediately.

This was done from the IP address {{.Addr}} on {{.Timestamp.Format "Mon, 02 Jan 2006 15:04:05 MST"}}

Thank you,
{{- end}}

{{define "email_update_subject"}}Your email address has been updated to {{.TargetUser.Email}}{{end}}
{{define "email_update_body" -}}
Hello {{.TargetUser.Name}}

Your email address has been updated to {{.TargetUser.Email}}, and your old email {{.Email}} has been removed. This was done from the IP address {{.Addr}} on {{.Timestamp.Format "Mon, 02 Jan 2006 15:04:05 MST"}}

If you did not make this change, you should contact {{.SupportEmail}} immediately. Otherwise, you can ignore this notice.

Thank you
{{- end}}

{{define "tenant_invite_subject"}}You have been invited to join {{.Tenant.Name}} by {{.CurrentUser.Email}} {{end}}
{{define "tenant_invite_body" -}}
Hello {{.TargetUser.Name}}

You have been invited to {{.Tenant.Name}} by {{.CurrentUser.Email}} 

Click the link to accept the invitation.

{{.TenantInvitePrefix}}{{.Token| urlquery}}

Thank you
{{- end}}`
)

var (
	emailTemplates = template.Must(template.New("email").Parse(templatesSrc))
)
