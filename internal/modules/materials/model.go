package materials

import "time"

const (
	TargetTypeVoice      = "voice"
	TargetTypeInstrument = "instrument"

	MaterialTypeAudioGuide = "audio_guide"
	MaterialTypeSheetMusic = "sheet_music"
	MaterialTypePlayback   = "playback"
	MaterialTypeLyrics     = "lyrics"
	MaterialTypeOther      = "other"
)

type Material struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ChoirID      string    `json:"choir_id"`
	SongID       string    `json:"song_id"`
	Name         string    `json:"name"`
	MaterialType string    `json:"material_type"`
	TargetType   string    `json:"target_type"`
	VoiceType    *string   `json:"voice_type,omitempty"`
	Archived     bool      `json:"archived"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateInput struct {
	SongID       string  `json:"song_id"`
	Name         string  `json:"name"`
	MaterialType *string `json:"material_type,omitempty"`
	TargetType   *string `json:"target_type,omitempty"`
	VoiceType    *string `json:"voice_type,omitempty"`
}

type UpdateInput struct {
	Name         string  `json:"name"`
	MaterialType *string `json:"material_type,omitempty"`
	TargetType   *string `json:"target_type,omitempty"`
	VoiceType    *string `json:"voice_type,omitempty"`
}

type LinkInstrumentInput struct {
	InstrumentID string `json:"instrument_id"`
}
