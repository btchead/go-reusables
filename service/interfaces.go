package service

// ServiceSequence defines the order in which services are started/stopped
type ServiceSequence int

const (
	// SequenceNone starts/stops services concurrently
	SequenceNone ServiceSequence = iota
	// SequenceFIFO starts services in registration order, stops in reverse
	SequenceFIFO
	// SequenceLIFO starts services in reverse registration order, stops in registration order
	SequenceLIFO
)