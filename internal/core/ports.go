package core

type InputPort interface {
	IsKeyDown(vk int) bool
	ForegroundTitle() string
	SendLeftDown() error
	SendLeftUp() error
}
