package events

import "time"

func BuildReminderSchedule(startAt time.Time, userIDs []string, now time.Time) []ScheduledReminder {
	startAt = startAt.UTC()
	now = now.UTC()

	reminders := make([]ScheduledReminder, 0, len(userIDs)*2)
	dayBeforeAt := startAt.Add(-24 * time.Hour)
	hourBeforeAt := startAt.Add(-1 * time.Hour)

	for _, userID := range userIDs {
		if dayBeforeAt.After(now) {
			reminders = append(reminders, ScheduledReminder{
				UserID:       userID,
				ReminderType: ReminderTypeDayBefore,
				ScheduledFor: dayBeforeAt,
				Status:       NotificationStatusPending,
			})
		}

		if hourBeforeAt.After(now) {
			reminders = append(reminders, ScheduledReminder{
				UserID:       userID,
				ReminderType: ReminderTypeHourBefore,
				ScheduledFor: hourBeforeAt,
				Status:       NotificationStatusPending,
			})
		}
	}

	return reminders
}
