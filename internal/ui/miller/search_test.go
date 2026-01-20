package miller

import (
	"testing"
	"time"

	"github.com/aljazfarkas/lazyhog/internal/client"
)

// mockListItem implements ListItem interface for testing
type mockListItem struct {
	id             string
	searchableText string
}

func (m mockListItem) RenderLine(width int, selected bool) string {
	return m.searchableText
}

func (m mockListItem) GetID() string {
	return m.id
}

func (m mockListItem) GetInspectorData() interface{} {
	return nil
}

func (m mockListItem) GetDistinctID() string {
	return m.id
}

func (m mockListItem) GetSearchableText() string {
	return m.searchableText
}

func TestApplyFilter_EmptyQuery(t *testing.T) {
	items := []ListItem{
		mockListItem{id: "1", searchableText: "event one"},
		mockListItem{id: "2", searchableText: "event two"},
	}

	m := Model{}
	result := m.applyFilter(items, "")

	if len(result) != 2 {
		t.Errorf("applyFilter with empty query should return all items, got %d", len(result))
	}
}

func TestApplyFilter_MatchingQuery(t *testing.T) {
	items := []ListItem{
		mockListItem{id: "1", searchableText: "page_view user123"},
		mockListItem{id: "2", searchableText: "click user456"},
		mockListItem{id: "3", searchableText: "page_view user789"},
	}

	m := Model{}
	result := m.applyFilter(items, "page_view")

	if len(result) != 2 {
		t.Errorf("applyFilter should return 2 matching items, got %d", len(result))
	}
}

func TestApplyFilter_CaseInsensitive(t *testing.T) {
	items := []ListItem{
		mockListItem{id: "1", searchableText: "PAGE_VIEW"},
		mockListItem{id: "2", searchableText: "page_view"},
		mockListItem{id: "3", searchableText: "Page_View"},
	}

	m := Model{}
	result := m.applyFilter(items, "page_view")

	if len(result) != 3 {
		t.Errorf("applyFilter should be case-insensitive, got %d matches", len(result))
	}
}

func TestApplyFilter_NoMatches(t *testing.T) {
	items := []ListItem{
		mockListItem{id: "1", searchableText: "event one"},
		mockListItem{id: "2", searchableText: "event two"},
	}

	m := Model{}
	result := m.applyFilter(items, "nonexistent")

	if len(result) != 0 {
		t.Errorf("applyFilter with no matches should return empty slice, got %d", len(result))
	}
}

func TestApplyFilter_PartialMatch(t *testing.T) {
	items := []ListItem{
		mockListItem{id: "1", searchableText: "user_authenticated"},
		mockListItem{id: "2", searchableText: "user_signup"},
		mockListItem{id: "3", searchableText: "page_view"},
	}

	m := Model{}
	result := m.applyFilter(items, "user")

	if len(result) != 2 {
		t.Errorf("applyFilter should match partial strings, got %d matches", len(result))
	}
}

func TestEventListItem_GetSearchableText(t *testing.T) {
	event := client.Event{
		Event:      "page_view",
		DistinctID: "user123",
	}

	item := EventListItem{Event: event}
	text := item.GetSearchableText()

	if text != "page_view user123" {
		t.Errorf("GetSearchableText() = %q, want %q", text, "page_view user123")
	}
}

func TestPersonListItem_GetSearchableText(t *testing.T) {
	person := client.Person{
		Name:        "John Doe",
		DistinctIDs: []string{"user123", "user456"},
	}

	item := PersonListItem{Person: person}
	text := item.GetSearchableText()

	expected := "John Doe user123 user456"
	if text != expected {
		t.Errorf("GetSearchableText() = %q, want %q", text, expected)
	}
}

func TestFlagListItem_GetSearchableText(t *testing.T) {
	flag := client.FeatureFlag{
		Key:  "new_feature",
		Name: "New Feature Flag",
	}

	item := FlagListItem{Flag: flag}
	text := item.GetSearchableText()

	expected := "new_feature New Feature Flag"
	if text != expected {
		t.Errorf("GetSearchableText() = %q, want %q", text, expected)
	}
}

func TestGetEffectiveListItems_NoFilter(t *testing.T) {
	m := Model{
		listItems:     []ListItem{mockListItem{id: "1"}, mockListItem{id: "2"}},
		filteredItems: nil,
	}

	result := m.getEffectiveListItems()

	if len(result) != 2 {
		t.Errorf("getEffectiveListItems with no filter should return all items, got %d", len(result))
	}
}

func TestGetEffectiveListItems_WithFilter(t *testing.T) {
	m := Model{
		listItems:     []ListItem{mockListItem{id: "1"}, mockListItem{id: "2"}, mockListItem{id: "3"}},
		filteredItems: []ListItem{mockListItem{id: "1"}},
	}

	result := m.getEffectiveListItems()

	if len(result) != 1 {
		t.Errorf("getEffectiveListItems with filter should return filtered items, got %d", len(result))
	}
}

// Test that search works on underlying data, not rendered output with ANSI codes
func TestApplyFilter_SearchesPlainText(t *testing.T) {
	// Create an event that when rendered would have ANSI codes
	event := client.Event{
		Event:      "button_click",
		DistinctID: "user@example.com",
		Timestamp:  time.Now(),
	}

	items := []ListItem{EventListItem{Event: event}}

	m := Model{}

	// Search should work on plain text
	result := m.applyFilter(items, "button_click")
	if len(result) != 1 {
		t.Errorf("applyFilter should find item by event name, got %d matches", len(result))
	}

	// Search should work on distinct_id
	result = m.applyFilter(items, "user@example")
	if len(result) != 1 {
		t.Errorf("applyFilter should find item by distinct_id, got %d matches", len(result))
	}
}
