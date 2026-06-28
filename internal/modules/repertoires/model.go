package repertoires

type Repertoire struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	ChoirID     string  `json:"choir_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Archived    bool    `json:"archived"`
}

type RepertoireSong struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	ChoirID        string  `json:"choir_id"`
	Title          string  `json:"title"`
	Composer       *string `json:"composer,omitempty"`
	Arranger       *string `json:"arranger,omitempty"`
	SongKey        *string `json:"song_key,omitempty"`
	Duration       *string `json:"duration,omitempty"`
	Notes          *string `json:"notes,omitempty"`
	Archived       bool    `json:"archived"`
	ExecutionOrder int     `json:"execution_order"`
	RepertoireNote *string `json:"repertoire_note,omitempty"`
}

type CreateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type AddSongInput struct {
	SongID         string  `json:"song_id"`
	ExecutionOrder int     `json:"execution_order"`
	Notes          *string `json:"notes,omitempty"`
}
