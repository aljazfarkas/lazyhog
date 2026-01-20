package client

import (
	"testing"
	"time"
)

func TestParseEventFromRow(t *testing.T) {
	tests := []struct {
		name     string
		row      []interface{}
		wantOk   bool
		wantUUID string
		wantEvent string
	}{
		{
			name:   "empty row",
			row:    []interface{}{},
			wantOk: false,
		},
		{
			name:   "row too short",
			row:    []interface{}{"uuid1", "event1", "2024-01-01T00:00:00Z", "distinct1", map[string]interface{}{}},
			wantOk: false,
		},
		{
			name: "valid row",
			row: []interface{}{
				"test-uuid-123",
				"page_view",
				"2024-01-15T10:30:00Z",
				"user-456",
				map[string]interface{}{"page": "/home"},
				"person-789",
			},
			wantOk:    true,
			wantUUID:  "test-uuid-123",
			wantEvent: "page_view",
		},
		{
			name: "valid row with nil properties",
			row: []interface{}{
				"test-uuid-456",
				"click",
				"2024-01-15T11:00:00Z",
				"user-123",
				nil,
				"person-456",
			},
			wantOk:    true,
			wantUUID:  "test-uuid-456",
			wantEvent: "click",
		},
		{
			name: "row with wrong types",
			row: []interface{}{
				123, // wrong type for UUID
				"event",
				"2024-01-15T10:30:00Z",
				"distinct",
				map[string]interface{}{},
				"person",
			},
			wantOk:    true, // should still parse, but UUID will be empty
			wantUUID:  "",
			wantEvent: "event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, ok := parseEventFromRow(tt.row)
			if ok != tt.wantOk {
				t.Errorf("parseEventFromRow() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && tt.wantOk {
				if event.UUID != tt.wantUUID {
					t.Errorf("parseEventFromRow() UUID = %v, want %v", event.UUID, tt.wantUUID)
				}
				if event.Event != tt.wantEvent {
					t.Errorf("parseEventFromRow() Event = %v, want %v", event.Event, tt.wantEvent)
				}
			}
		})
	}
}

func TestParseEventFromRow_Timestamp(t *testing.T) {
	row := []interface{}{
		"uuid",
		"event",
		"2024-01-15T10:30:00Z",
		"distinct",
		map[string]interface{}{},
		"person",
	}

	event, ok := parseEventFromRow(row)
	if !ok {
		t.Fatal("parseEventFromRow() returned false")
	}

	expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	if !event.Timestamp.Equal(expected) {
		t.Errorf("parseEventFromRow() Timestamp = %v, want %v", event.Timestamp, expected)
	}
}

func TestParseEventFromRow_Properties(t *testing.T) {
	props := map[string]interface{}{
		"page":      "/home",
		"referrer":  "google.com",
		"timestamp": 1234567890,
	}

	row := []interface{}{
		"uuid",
		"event",
		"2024-01-15T10:30:00Z",
		"distinct",
		props,
		"person",
	}

	event, ok := parseEventFromRow(row)
	if !ok {
		t.Fatal("parseEventFromRow() returned false")
	}

	if len(event.Properties) != 3 {
		t.Errorf("parseEventFromRow() Properties length = %d, want 3", len(event.Properties))
	}

	if event.Properties["page"] != "/home" {
		t.Errorf("parseEventFromRow() Properties[page] = %v, want /home", event.Properties["page"])
	}
}

func TestParseEventFromRow_IDEqualsUUID(t *testing.T) {
	row := []interface{}{
		"my-unique-uuid",
		"event",
		"2024-01-15T10:30:00Z",
		"distinct",
		map[string]interface{}{},
		"person",
	}

	event, ok := parseEventFromRow(row)
	if !ok {
		t.Fatal("parseEventFromRow() returned false")
	}

	if event.ID != event.UUID {
		t.Errorf("parseEventFromRow() ID (%v) should equal UUID (%v)", event.ID, event.UUID)
	}
}
