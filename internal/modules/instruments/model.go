package instruments

type Instrument struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	ChoirID     string  `json:"choir_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Archived    bool    `json:"archived"`
}

type InstrumentUser struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Active   bool   `json:"active"`
}

type CreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

type UpdateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
}

type LinkUserInput struct {
	UserID string `json:"user_id"`
}
