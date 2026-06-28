package auth

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Tenant struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name"`
}

type User struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Manager  bool   `json:"manager"`
}

type Session struct {
	Tenant Tenant `json:"tenant"`
	User   User   `json:"user"`
}

type LoginIdentity struct {
	TenantID          string
	TenantSlug        string
	TenantDisplayName string
	UserID            string
	Email             string
	FullName          string
	PasswordHash      *string
	Manager           bool
}
