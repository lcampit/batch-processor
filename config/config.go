package config

import "sync"

// BatchProcessorConfig contains definitions about processor channels, batch size and ticker duration
type BatchProcessorConfig[InputType any, OutputType any, ErrorType any] struct {
	// TODO: add docs for each parameter
	BatchSize            int
	TickerTimeoutSeconds int
	InputChannel         chan InputType
	OutputChannel        chan OutputType
	ErrorChannel         chan ErrorType

	WaitGroup *sync.WaitGroup
}
