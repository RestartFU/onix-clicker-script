package winapi

import (
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	INPUT_MOUSE          = 0
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004

	VK_LBUTTON = 0x01

	WH_MOUSE_LL = 14

	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202

	LLMHF_INJECTED          = 0x00000001
	LLMHF_LOWER_IL_INJECTED = 0x00000002
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procSendInput           = user32.NewProc("SendInput")
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	_                       = kernel32.NewProc("GetCurrentThreadId")
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

type POINT struct {
	X int32
	Y int32
}

type MSLLHOOKSTRUCT struct {
	Pt          POINT
	MouseData   uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	HWnd     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       POINT
	LPrivate uint32
}

type Input struct {
	mouseDown atomic.Bool
}

var (
	hookOnce     sync.Once
	hookReady    atomic.Bool
	hookInput    *Input
	hookHandle   uintptr
	hookCallback = syscall.NewCallback(mouseHookProc)
)

func NewInput() *Input {
	i := &Input{}
	hookInput = i
	hookOnce.Do(func() {
		go runMouseHook()
	})
	return i
}

func (i *Input) IsMouseDown() bool {
	if hookReady.Load() {
		return i.mouseDown.Load()
	}
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(VK_LBUTTON))
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

func runMouseHook() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hookHandle, _, _ = procSetWindowsHookExW.Call(
		uintptr(WH_MOUSE_LL),
		hookCallback,
		0,
		0,
	)
	if hookHandle == 0 {
		return
	}
	hookReady.Store(true)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)
		if int32(ret) == 0 || int32(ret) == -1 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	hookReady.Store(false)
	procUnhookWindowsHookEx.Call(hookHandle)
}

func mouseHookProc(code int, wParam uintptr, lParam uintptr) uintptr {
	if code >= 0 && hookInput != nil {
		if wParam == WM_LBUTTONDOWN || wParam == WM_LBUTTONUP {
			info := (*MSLLHOOKSTRUCT)(unsafe.Pointer(lParam))
			if info.Flags&(LLMHF_INJECTED|LLMHF_LOWER_IL_INJECTED) == 0 {
				if wParam == WM_LBUTTONDOWN {
					hookInput.mouseDown.Store(true)
				} else {
					hookInput.mouseDown.Store(false)
				}
			}
		}
	}
	ret, _, _ := procCallNextHookEx.Call(hookHandle, uintptr(code), wParam, lParam)
	return ret
}
