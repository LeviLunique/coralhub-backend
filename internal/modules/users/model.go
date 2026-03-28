package users

type User struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Active   bool   `json:"active"`
}

type CreateInput struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
