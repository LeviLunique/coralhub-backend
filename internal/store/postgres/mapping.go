package postgres

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	text := value.String
	return &text
}

func textValue(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: *value,
		Valid:  true,
	}
}

func timestamptzValue(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value.UTC(),
		Valid: true,
	}
}

func uuidString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}

	encoded := hex.EncodeToString(value.Bytes[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}

func parseUUID(value string) (pgtype.UUID, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(value), "-", "")
	if len(normalized) != 32 {
		return pgtype.UUID{}, errors.New("invalid uuid length")
	}

	decoded, err := hex.DecodeString(normalized)
	if err != nil {
		return pgtype.UUID{}, err
	}

	var bytes [16]byte
	copy(bytes[:], decoded)

	return pgtype.UUID{
		Bytes: bytes,
		Valid: true,
	}, nil
}
