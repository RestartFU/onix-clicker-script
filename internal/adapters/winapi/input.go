package winapi

import (
	"syscall"
	"unsafe"
)

const (
	INPUT_MOUSE          = 0
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procSendInput           = user32.NewProc("SendInput")
)

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type INPUT struct {
	Type uint32
	_    uint32 // padding on 64-bit
	Mi   MOUSEINPUT
}

type Input struct{}

func NewInput() *Input {
	return &Input{}
}

func (i *Input) IsKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return (ret & 0x8000) != 0
}

func (i *Input) ForegroundTitle() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}

	buf := make([]uint16, 256)
	procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf)
}

func (i *Input) SendLeftDown() error {
	return sendMouse(MOUSEEVENTF_LEFTDOWN)
}

func (i *Input) SendLeftUp() error {
	return sendMouse(MOUSEEVENTF_LEFTUP)
}

func sendMouse(flags uint32) error {
	in := INPUT{
		Type: INPUT_MOUSE,
		Mi: MOUSEINPUT{
			DwFlags: flags,
		},
	}
	procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&in)),
		unsafe.Sizeof(in),
	)
	return nil
}
