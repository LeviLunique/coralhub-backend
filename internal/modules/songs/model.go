package songs

type Song struct {
	ID       string  `json:"id"`
	TenantID string  `json:"tenant_id"`
	ChoirID  string  `json:"choir_id"`
	Title    string  `json:"title"`
	Composer *string `json:"composer,omitempty"`
	Arranger *string `json:"arranger,omitempty"`
	SongKey  *string `json:"song_key,omitempty"`
	Duration *string `json:"duration,omitempty"`
	Notes    *string `json:"notes,omitempty"`
	Archived bool    `json:"archived"`
}

type CreateInput struct {
	Title    string  `json:"title"`
	Composer *string `json:"composer,omitempty"`
	Arranger *string `json:"arranger,omitempty"`
	SongKey  *string `json:"song_key,omitempty"`
	Duration *string `json:"duration,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

type UpdateInput struct {
	Title    string  `json:"title"`
	Composer *string `json:"composer,omitempty"`
	Arranger *string `json:"arranger,omitempty"`
	SongKey  *string `json:"song_key,omitempty"`
	Duration *string `json:"duration,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}
