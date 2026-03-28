package devices

const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

type DeviceToken struct {
	ID       string
	TenantID string
	UserID   string
	Platform string
	Token    string
	Active   bool
}

type CreateParams struct {
	TenantID string
	UserID   string
	Platform string
	Token    string
}
