package events

type Consumer interface {
	Start() error
	StartBatch() error
}
