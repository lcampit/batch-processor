package processor

import (
	"context"
	"sync"
	"time"

	"github.com/lcampit/batch-processor/config"
)

type batchProcessor[InputType any, OutputType any, ErrorType comparable] struct {
	batchSize            int
	tickerTimeoutSeconds int
	inputChannel         chan InputType
	outputChannel        chan OutputType
	errorChannel         chan ErrorType

	processingFunction func(ctx context.Context, input []InputType) (OutputType, ErrorType)
	waitGroup          *sync.WaitGroup
}

type BatchProcessor[InputType any, OutputType any, ErrorType comparable] interface {
	SetProcessingFunction(function func(ctx context.Context, input []InputType) (OutputType, ErrorType))
	Start(ctx context.Context)
}

// NewBatchProcessor creates a new processor with the given configuration
func NewBatchProcessor[InputType any, OutputType any, ErrorType comparable](config config.BatchProcessorConfig[InputType, OutputType, ErrorType]) BatchProcessor[InputType, OutputType, ErrorType] {
	return &batchProcessor[InputType, OutputType, ErrorType]{
		batchSize:            config.BatchSize,
		tickerTimeoutSeconds: config.TickerTimeoutSeconds,
		inputChannel:         config.InputChannel,
		outputChannel:        config.OutputChannel,
		errorChannel:         config.ErrorChannel,
		waitGroup:            config.WaitGroup,
	}
}

// Start starts the processor. It will panic if the processor has no processing function set, use
// SetProcessingFunction to instruct the processor about what to do with each batch of input data
//
// This is meant to be run in a dedicated go routine and will block until che context provided is canceled
func (p *batchProcessor[InputType, OutputType, ErrorType]) Start(ctx context.Context) {
	if p.processingFunction == nil {
		panic("failed to start batch processor: process function is empty, set it via SetProcessingFunction")
	}

	cache := make([]InputType, p.batchSize)
	tick := time.NewTicker(time.Duration(p.tickerTimeoutSeconds) * time.Second)
	for {
		select {
		case input := <-p.inputChannel:

			cache = append(cache, input)
			if len(cache) < p.batchSize {
				break
			}

			tick.Stop()

			p.processBatch(ctx, cache)
			tick = time.NewTicker(time.Duration(p.tickerTimeoutSeconds) * time.Second)
			cache = cache[:0]

		case <-tick.C:
			p.processBatch(ctx, cache)
			tick = time.NewTicker(time.Duration(p.tickerTimeoutSeconds) * time.Second)
			cache = cache[:0]

		case <-ctx.Done():
			p.processBatch(ctx, cache)
			tick.Stop()
			p.closeOutputChannels()
			if p.waitGroup != nil {
				p.waitGroup.Done()
			}
			return
		}
	}
}

func (p *batchProcessor[InputType, OutputType, ErrorType]) processBatch(ctx context.Context, inputs []InputType) {
	output, err := p.processingFunction(ctx, inputs)
	if !isZeroValue(err) {
		p.errorChannel <- err
	} else {
		p.outputChannel <- output
	}
}

func (p *batchProcessor[InputType, OutputType, ErrorType]) closeOutputChannels() {
	close(p.outputChannel)
	close(p.errorChannel)
}

// SetProcessingFunction sets the processing function that will be applied to each batch of data
// The processing function will receieve a slice of input data and return either an output or an error.
// Processing function results will then be written on the corresponding processor channel
func (p *batchProcessor[InputType, OutputType, ErrorType]) SetProcessingFunction(function func(ctx context.Context, input []InputType) (OutputType, ErrorType)) {
	p.processingFunction = function
}

func isZeroValue[T comparable](error T) bool {
	return error == *new(T)
}
