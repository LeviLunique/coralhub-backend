package memberships

const (
	RoleManager = "manager"
	RoleMember  = "member"
)

type Membership struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	ChoirID  string `json:"choir_id"`
	UserID   string `json:"user_id"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

type CreateInput struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}
