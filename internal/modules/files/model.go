package files

type File struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	ChoirID          string `json:"choir_id,omitempty"`
	VoiceKitID       string `json:"voice_kit_id"`
	OriginalFilename string `json:"original_filename"`
	StoredFilename   string `json:"stored_filename"`
	ContentType      string `json:"content_type"`
	SizeBytes        int64  `json:"size_bytes"`
	StorageKey       string `json:"storage_key"`
	Active           bool   `json:"active"`
}

type CreateInput struct {
	OriginalFilename string `json:"original_filename"`
	StoredFilename   string `json:"stored_filename"`
	ContentType      string `json:"content_type"`
	SizeBytes        int64  `json:"size_bytes"`
	StorageKey       string `json:"storage_key"`
}
