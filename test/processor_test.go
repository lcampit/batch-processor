package test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lcampit/batch-processor/config"
	"github.com/lcampit/batch-processor/processor"
	"github.com/stretchr/testify/suite"
)

type ProcessorTestSuite struct {
	suite.Suite

	BatchSize             int
	TickerDurationSeconds int
}

func (s *ProcessorTestSuite) SetupSuite() {
	s.BatchSize = 100
	s.TickerDurationSeconds = 5
}

func (s *ProcessorTestSuite) TestProcessorPanicsWithoutProcessingFunction() {
	ctx := context.Background()
	config := config.BatchProcessorConfig[int, int, error]{
		BatchSize:            s.BatchSize,
		TickerTimeoutSeconds: s.TickerDurationSeconds,
		InputChannel:         make(chan int),
		OutputChannel:        make(chan int),
		ErrorChannel:         make(chan error),
	}
	processor := processor.NewBatchProcessor(config)
	s.Assert().Panics(func() { processor.Start(ctx) })
}

func (s *ProcessorTestSuite) TestProcessorWithScalarLogic() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.TickerDurationSeconds+2)*time.Second)

	inputChannel := make(chan int, s.BatchSize)
	outputChannel := make(chan int)
	wg := sync.WaitGroup{}
	config := config.BatchProcessorConfig[int, int, error]{
		BatchSize:            s.BatchSize,
		TickerTimeoutSeconds: s.TickerDurationSeconds,
		InputChannel:         inputChannel,
		OutputChannel:        outputChannel,
		ErrorChannel:         make(chan error),
		WaitGroup:            &wg,
	}
	processor := processor.NewBatchProcessor(config)

	processor.SetProcessingFunction(func(ctx context.Context, input []int) (int, error) {
		if len(input) == 0 {
			return 0, nil
		}
		sum := 0
		for _, inp := range input {
			sum = sum + inp
		}
		return sum, nil
	})

	wg.Add(1)
	go processor.Start(ctx)

	expectedSum := 0
	for i := range s.BatchSize {
		inputChannel <- i
		expectedSum = expectedSum + i
	}

	result := 0
	for res := range outputChannel {
		result = result + res
	}
	// wait for the context to expire
	wg.Wait()
	cancel()

	s.Assert().Equal(result, expectedSum)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}
