package fcm

import (
	"context"
	"strings"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/LeviLunique/coralhub-backend/internal/modules/devices"
	"github.com/LeviLunique/coralhub-backend/internal/modules/notifications"
	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	"google.golang.org/api/option"
)

type messagingClient interface {
	SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

type classifyFunc func(error) string

type Sender struct {
	client      messagingClient
	deviceStore devices.Repository
	classify    classifyFunc
}

func New(ctx context.Context, cfg platformconfig.FirebaseConfig, deviceStore devices.Repository) (*Sender, error) {
	app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return newSender(client, deviceStore, classifyMessagingError), nil
}

func newSender(client messagingClient, deviceStore devices.Repository, classify classifyFunc) *Sender {
	return &Sender{
		client:      client,
		deviceStore: deviceStore,
		classify:    classify,
	}
}

func (s *Sender) Deliver(ctx context.Context, notification notifications.Notification) notifications.DeliveryResult {
	tokens, err := s.deviceStore.ListActiveByUserID(ctx, notification.TenantID, notification.UserID)
	if err != nil {
		return notifications.DeliveryResult{
			Kind:         notifications.DeliveryTransientFailure,
			ErrorMessage: err.Error(),
		}
	}

	if len(tokens) == 0 {
		return notifications.DeliveryResult{
			Kind:         notifications.DeliveryInvalidToken,
			ErrorMessage: "no active device tokens",
		}
	}

	message := &messaging.MulticastMessage{
		Tokens: collectTokens(tokens),
		Data: map[string]string{
			"tenant_id":     notification.TenantID,
			"event_id":      notification.EventID,
			"reminder_type": notification.ReminderType,
		},
		Notification: &messaging.Notification{
			Title: "CoralHub reminder",
			Body:  buildBody(notification.ReminderType),
		},
	}

	response, err := s.client.SendEachForMulticast(ctx, message)
	if err != nil {
		return notifications.DeliveryResult{
			Kind:         notifications.DeliveryTransientFailure,
			ErrorMessage: err.Error(),
		}
	}

	successCount := 0
	remainingActive := len(tokens)
	lastTransientMessage := ""

	for index, sendResponse := range response.Responses {
		if sendResponse.Success {
			successCount++
			continue
		}

		classification := s.classify(sendResponse.Error)
		switch classification {
		case notifications.DeliveryInvalidToken:
			if err := s.deviceStore.DeactivateByToken(ctx, notification.TenantID, tokens[index].Token); err == nil {
				remainingActive--
			}
		default:
			if message := strings.TrimSpace(sendResponse.Error.Error()); message != "" {
				lastTransientMessage = message
			}
		}
	}

	if successCount > 0 {
		return notifications.DeliveryResult{Kind: notifications.DeliverySent}
	}

	if remainingActive == 0 {
		return notifications.DeliveryResult{
			Kind:         notifications.DeliveryInvalidToken,
			ErrorMessage: "all device tokens are inactive or invalid",
		}
	}

	if lastTransientMessage == "" {
		lastTransientMessage = "temporary fcm delivery failure"
	}

	return notifications.DeliveryResult{
		Kind:         notifications.DeliveryTransientFailure,
		ErrorMessage: lastTransientMessage,
	}
}

func collectTokens(items []devices.DeviceToken) []string {
	tokens := make([]string, 0, len(items))
	for _, item := range items {
		tokens = append(tokens, item.Token)
	}
	return tokens
}

func buildBody(reminderType string) string {
	switch reminderType {
	case "day_before":
		return "You have a choir event tomorrow."
	case "hour_before":
		return "Your choir event starts in one hour."
	default:
		return "You have an upcoming choir event."
	}
}

func classifyMessagingError(err error) string {
	if err == nil {
		return notifications.DeliverySent
	}

	if messaging.IsRegistrationTokenNotRegistered(err) || messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
		return notifications.DeliveryInvalidToken
	}

	if messaging.IsUnavailable(err) || messaging.IsInternal(err) || messaging.IsQuotaExceeded(err) || messaging.IsThirdPartyAuthError(err) || messaging.IsSenderIDMismatch(err) {
		return notifications.DeliveryTransientFailure
	}

	return notifications.DeliveryTransientFailure
}

var _ notifications.Sender = (*Sender)(nil)
