package notification

import (
	"context"
	"time"

	"github.com/ecadlabs/auth/storage"
)

type NotificationData struct {
	Addr        string
	Email       string
	Tenant      *storage.TenantModel
	CurrentUser *storage.User
	TargetUser  *storage.User
	To          []string
	Token       string
	TokenMaxAge time.Duration
}

const (
	NotificationInvite             = "invite"
	NotificationTenantInvite       = "tenant_invite"
	NotificationReset              = "reset"
	NotificationEmailUpdateRequest = "email_update_request"
	NotificationEmailUpdate        = "email_update"
)

type Notifier interface {
	Notify(context.Context, string, *NotificationData) error
}
