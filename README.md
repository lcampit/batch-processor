# Batch Processor

Batch Processor is a library that aims at extrapolating batching logic in Go,
letting you focus on what matters the most: Your logic.

Feed the processor with data and reap the results and errors from the
channels provided. Is that simple!

## Usage

Here's a usage example:

```go
ctx := context.Background()

inputChannel := make(chan int, 10)
outputChannel := make(chan int, 10)
errorChannel := make(chan error, 10)

processor := batchProcessor.NewBatchProcessor(
  &BatchProcessor.Config[int, int, error]{
    BatchSize: 10,
    TickerTimeoutSeconds: 5,
    InputChannel: inputChannel,
    OutputChannel: outputChannel,
    ErrorChannel: errorChannel,
  }
)

processor.SetProcessingFunction(func(batch []int) (int, error) {
  // your amazing logic here
})

// Let's go!
go processor.Start(ctx)


// Hook up any data producing and consuming modules to the three channels above
```

In the example above, any data sent on the `inputChannel` will be accumulated
until the ticker expires, or the data size reaches the batch size defined.
Then, the current batch of data will be processed using the function provided.
Any results or errors will be sent to the relevant channel, ready for consumption.

Thanks to Go _generics_, BatchProcessor can handle just any type of data, making sure that the channels and processing function types are those expected.

This allows developers to focus on the actual processing logic rather than lose
precious time devising batching mechanisms. Moreover, the channels structure allows
for easy integration testing as the batch processor itself has no added dependency.

## Contributing

Contributions are welcome! Feel free to clone the project and create pull
requests or add issues.
