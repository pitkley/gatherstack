package usagestats

import (
	"context"
	"encoding/json"
	"math/rand"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

// Event represents a request to log telemetry.
type Event struct {
	EventName    string
	UserID       int32
	UserCookieID string
	// FirstSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	FirstSourceURL *string
	// LastSourceURL is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	LastSourceURL    *string
	URL              string
	Source           string
	EvaluatedFlagSet featureflag.EvaluatedFlagSet
	CohortID         *string
	// Referrer is only logged for Cloud events; therefore, this only goes to the BigQuery database
	// and does not go to the Postgres DB.
	Referrer         *string
	OriginalReferrer *string
	SessionReferrer  *string
	SessionFirstURL  *string
	Argument         json.RawMessage
	PublicArgument   json.RawMessage
	UserProperties   json.RawMessage
	DeviceID         *string
	InsertID         *string
	EventID          *int32
	DeviceSessionID  *string
}

// LogBackendEvent is a convenience function for logging backend events.
func LogBackendEvent(db database.DB, userID int32, deviceID, eventName string, argument, publicArgument json.RawMessage, evaluatedFlagSet featureflag.EvaluatedFlagSet, cohortID *string) error {
	insertID, _ := uuid.NewRandom()
	insertIDFinal := insertID.String()
	eventID := int32(rand.Int())
	return LogEvent(context.Background(), db, Event{
		EventName:        eventName,
		UserID:           userID,
		UserCookieID:     "backend", // Use a non-empty string here to avoid the event_logs table's user existence constraint causing issues
		URL:              "",
		Source:           "BACKEND",
		Argument:         argument,
		PublicArgument:   publicArgument,
		UserProperties:   json.RawMessage("{}"),
		EvaluatedFlagSet: evaluatedFlagSet,
		CohortID:         cohortID,
		DeviceID:         &deviceID,
		InsertID:         &insertIDFinal,
		EventID:          &eventID,
	})
}

// LogEvent logs an event.
func LogEvent(ctx context.Context, db database.DB, args Event) error {
	return LogEvents(ctx, db, []Event{args})
}

// LogEvents logs a batch of events.
func LogEvents(ctx context.Context, db database.DB, events []Event) error {
	if !conf.EventLoggingEnabled() {
		return nil
	}

	if err := logLocalEvents(ctx, db, events); err != nil {
		return err
	}

	return nil
}

// logLocalEvents logs a batch of user events.
func logLocalEvents(ctx context.Context, db database.DB, events []Event) error {
	databaseEvents, err := serializeLocalEvents(events)
	if err != nil {
		return err
	}

	return db.EventLogs().BulkInsert(ctx, databaseEvents)
}

func serializeLocalEvents(events []Event) ([]*database.Event, error) {
	databaseEvents := make([]*database.Event, 0, len(events))
	for _, event := range events {
		if event.EventName == "SearchResultsQueried" {
			if err := logSiteSearchOccurred(); err != nil {
				return nil, err
			}
		}
		if event.EventName == "findReferences" {
			if err := logSiteFindRefsOccurred(); err != nil {
				return nil, err
			}
		}

		databaseEvents = append(databaseEvents, &database.Event{
			Name:             event.EventName,
			URL:              event.URL,
			UserID:           uint32(event.UserID),
			AnonymousUserID:  event.UserCookieID,
			Source:           event.Source,
			Argument:         event.Argument,
			Timestamp:        timeNow().UTC(),
			EvaluatedFlagSet: event.EvaluatedFlagSet,
			CohortID:         event.CohortID,
			PublicArgument:   event.PublicArgument,
			FirstSourceURL:   event.FirstSourceURL,
			LastSourceURL:    event.LastSourceURL,
			Referrer:         event.Referrer,
			DeviceID:         event.DeviceID,
			InsertID:         event.InsertID,
		})
	}

	return databaseEvents, nil
}
