package voicekits

type VoiceKit struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	ChoirID     string  `json:"choir_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Active      bool    `json:"active"`
}

type CreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}
