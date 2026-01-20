package client

import "context"

// PostHogClient defines the interface for PostHog API operations.
// This interface enables testability by allowing mock implementations.
type PostHogClient interface {
	// Events
	ListRecentEvents(ctx context.Context, limit int) ([]Event, error)
	GetEvent(ctx context.Context, eventID string) (*Event, error)

	// Persons
	GetPerson(ctx context.Context, distinctID string) (*Person, error)
	GetPersonEvents(ctx context.Context, distinctID string, limit int) ([]Event, error)
	ListPersons(ctx context.Context, limit int) ([]Person, error)

	// Feature Flags
	ListFlags(ctx context.Context) ([]FeatureFlag, error)
	GetFlag(ctx context.Context, flagID int) (*FeatureFlag, error)
	ToggleFlag(ctx context.Context, flagID int, active bool) error

	// Projects
	FetchProjects(ctx context.Context) ([]Project, error)
	GetProjectID() int
	SetProjectID(projectID int)
	GetProjects() []Project

	// Connection
	TestConnection(ctx context.Context) error
	InitializeProject(ctx context.Context) error

	// Query
	ExecuteQuery(ctx context.Context, query string) (*QueryResult, error)

	// Lifecycle
	Close() error
}

// Ensure Client implements PostHogClient
var _ PostHogClient = (*Client)(nil)
