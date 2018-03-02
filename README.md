# Generic Factory

This package is a library for having a factory process data/jobs in parallel without blocking execution of your main program.

Example use cases:

* set number of workers uploading to S3 - limits number of concurrent uploads
* process data in parallel -> use doneChans for easy synchronization (see example)

## Getting Started

```
dep ensure -add github.com/BakerNet/generic-factory
```
or
```
go get github.com/BakerNet/generic-factory
```

Then import library in your code.

To use:
```Go
f := factory.NewFactory(ctx, numWorkers)
// register callbacks - remember to have callbacks use type assertion
f.Register(callback1)
f.Register(callback2)
// send job to next available worker
doneChan := f.Dispatch(jobData)
// close factory - will end all go routines.  Unfinished jobs will send error on their done channels
f.Close()
```

## Documentation (Godoc)

* [Godoc](https://godoc.org/github.com/BakerNet/generic-factory)

## Running the tests

```
go test ./...
```

## Built With

* [Go](https://golang.org) - Best language and standard library around

## Contributing

Have fun

## Authors

* **Hans Baker** - [BakerNet](https://github.com/BakerNet)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

