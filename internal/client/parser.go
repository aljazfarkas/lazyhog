package client

import (
	"time"
)

// parseEventFromRow parses a HogQL query result row into an Event struct.
// Expected column order: uuid, event, timestamp, distinct_id, properties, person_id
func parseEventFromRow(row []interface{}) (Event, bool) {
	if len(row) < 6 {
		return Event{}, false
	}

	event := Event{}

	// UUID (column 0)
	if uuid, ok := row[0].(string); ok {
		event.UUID = uuid
		event.ID = uuid // Use UUID as ID
	}

	// Event name (column 1)
	if eventName, ok := row[1].(string); ok {
		event.Event = eventName
	}

	// Timestamp (column 2)
	if ts, ok := row[2].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
			event.Timestamp = parsed
		}
	}

	// Distinct ID (column 3)
	if distinctID, ok := row[3].(string); ok {
		event.DistinctID = distinctID
	}

	// Properties (column 4)
	if props, ok := row[4].(map[string]interface{}); ok {
		event.Properties = props
	}

	// Person ID (column 5)
	if personID, ok := row[5].(string); ok {
		event.PersonID = personID
	}

	return event, true
}
