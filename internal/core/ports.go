package core

type InputPort interface {
	IsMouseDown() bool
	ForegroundTitle() string
	SendLeftDown() error
	SendLeftUp() error
}
