package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	pipelineStream := in

	for _, stage := range stages {
		pipelineStream = stage(stopWhenDone(done, pipelineStream))
	}

	return stopWhenDone(done, pipelineStream)
}

func stopWhenDone(doneChannel In, inputChannel In) Out {
	outputChannel := make(Bi)

	go func() {
		defer func() {
			close(outputChannel)
			drainChannel(inputChannel)
		}()

		for {
			select {
			case <-doneChannel:
				return
			default:
			}

			select {
			case <-doneChannel:
				return
			case value, isOpen := <-inputChannel:
				if !isOpen {
					return
				}

				select {
				case <-doneChannel:
					return
				case outputChannel <- value:
				}
			}
		}
	}()

	return outputChannel
}

func drainChannel(channel In) {
	for value := range channel {
		_ = value
	}
}
