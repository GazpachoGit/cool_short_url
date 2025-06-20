package domain

// some domain event format consumable by anther service(db event record -> domain event)
type Event struct {
	ID        int
	EventType string
	Payload   string
}
