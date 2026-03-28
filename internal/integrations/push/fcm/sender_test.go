package fcm

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/LeviLunique/coralhub-backend/internal/modules/devices"
	"github.com/LeviLunique/coralhub-backend/internal/modules/notifications"
)

type fakeClient struct {
	response *messaging.BatchResponse
	err      error
	message  *messaging.MulticastMessage
}

func (f *fakeClient) SendEachForMulticast(_ context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	f.message = message
	if f.err != nil {
		return nil, f.err
	}
	return f.response, nil
}

type fakeDeviceStore struct {
	tokens      []devices.DeviceToken
	listErr     error
	deactivated []string
}

func (f *fakeDeviceStore) Create(_ context.Context, _ devices.CreateParams) (devices.DeviceToken, error) {
	return devices.DeviceToken{}, nil
}

func (f *fakeDeviceStore) ListActiveByUserID(_ context.Context, _, _ string) ([]devices.DeviceToken, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.tokens, nil
}

func (f *fakeDeviceStore) DeactivateByToken(_ context.Context, _ string, token string) error {
	f.deactivated = append(f.deactivated, token)
	return nil
}

func TestSenderDeliverReturnsInvalidTokenWhenNoActiveTokens(t *testing.T) {
	sender := newSender(&fakeClient{}, &fakeDeviceStore{}, func(error) string {
		return notifications.DeliveryTransientFailure
	})

	result := sender.Deliver(context.Background(), notifications.Notification{
		TenantID: "tenant-1",
		UserID:   "user-1",
	})

	if result.Kind != notifications.DeliveryInvalidToken {
		t.Fatalf("result.Kind = %q, want %q", result.Kind, notifications.DeliveryInvalidToken)
	}
}

func TestSenderDeliverReturnsSentWhenAnyTokenSucceeds(t *testing.T) {
	client := &fakeClient{
		response: &messaging.BatchResponse{
			Responses: []*messaging.SendResponse{
				{Success: true},
				{Success: false, Error: errors.New("invalid-token")},
			},
		},
	}
	store := &fakeDeviceStore{
		tokens: []devices.DeviceToken{
			{Token: "token-1"},
			{Token: "token-2"},
		},
	}
	sender := newSender(client, store, func(err error) string {
		if err != nil && err.Error() == "invalid-token" {
			return notifications.DeliveryInvalidToken
		}
		return notifications.DeliveryTransientFailure
	})

	result := sender.Deliver(context.Background(), notifications.Notification{
		TenantID:     "tenant-1",
		UserID:       "user-1",
		EventID:      "event-1",
		ReminderType: "day_before",
	})

	if result.Kind != notifications.DeliverySent {
		t.Fatalf("result.Kind = %q, want %q", result.Kind, notifications.DeliverySent)
	}
	if len(store.deactivated) != 1 || store.deactivated[0] != "token-2" {
		t.Fatalf("deactivated = %v, want [token-2]", store.deactivated)
	}
	if len(client.message.Tokens) != 2 {
		t.Fatalf("len(client.message.Tokens) = %d, want 2", len(client.message.Tokens))
	}
}

func TestSenderDeliverReturnsTransientFailureWhenRetryableTokensRemain(t *testing.T) {
	client := &fakeClient{
		response: &messaging.BatchResponse{
			Responses: []*messaging.SendResponse{
				{Success: false, Error: errors.New("temporary")},
				{Success: false, Error: errors.New("invalid")},
			},
		},
	}
	store := &fakeDeviceStore{
		tokens: []devices.DeviceToken{
			{Token: "token-1"},
			{Token: "token-2"},
		},
	}
	sender := newSender(client, store, func(err error) string {
		switch err.Error() {
		case "invalid":
			return notifications.DeliveryInvalidToken
		default:
			return notifications.DeliveryTransientFailure
		}
	})

	result := sender.Deliver(context.Background(), notifications.Notification{
		TenantID: "tenant-1",
		UserID:   "user-1",
	})

	if result.Kind != notifications.DeliveryTransientFailure {
		t.Fatalf("result.Kind = %q, want %q", result.Kind, notifications.DeliveryTransientFailure)
	}
}

func TestSenderDeliverReturnsInvalidTokenWhenAllTokensAreInvalid(t *testing.T) {
	client := &fakeClient{
		response: &messaging.BatchResponse{
			Responses: []*messaging.SendResponse{
				{Success: false, Error: errors.New("invalid-1")},
				{Success: false, Error: errors.New("invalid-2")},
			},
		},
	}
	store := &fakeDeviceStore{
		tokens: []devices.DeviceToken{
			{Token: "token-1"},
			{Token: "token-2"},
		},
	}
	sender := newSender(client, store, func(error) string {
		return notifications.DeliveryInvalidToken
	})

	result := sender.Deliver(context.Background(), notifications.Notification{
		TenantID: "tenant-1",
		UserID:   "user-1",
	})

	if result.Kind != notifications.DeliveryInvalidToken {
		t.Fatalf("result.Kind = %q, want %q", result.Kind, notifications.DeliveryInvalidToken)
	}
	if len(store.deactivated) != 2 {
		t.Fatalf("len(store.deactivated) = %d, want 2", len(store.deactivated))
	}
}
