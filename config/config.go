// Package config contains everything related to the
// batch processor configuration, including BatchSize, TickerTimeour
// and the input or output channels
package config

import "sync"

// BatchProcessorConfig is used to create a BatchProcessor with the given characteristics
type BatchProcessorConfig[InputType any, OutputType any, ErrorType any] struct {
	BatchSize            int // batch size before sending the batch to process
	TickerTimeoutSeconds int // ticker duration before sending the batch to process, in seconds

	InputChannel  chan InputType
	OutputChannel chan OutputType
	ErrorChannel  chan ErrorType

	WaitGroup *sync.WaitGroup // optional, used for synchronization if needed
}
