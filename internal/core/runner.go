package core

import (
	"context"
	"strings"
	"time"
)

const (
	VK_LBUTTON = 0x01
	VK_F9      = 0x78
)

func Run(ctx context.Context, clicker *Clicker, input InputPort) error {
	wasDown := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if input.IsKeyDown(VK_F9) {
			return nil
		}

		if !clicker.Enabled() {
			if wasDown {
				_ = input.SendLeftUp()
				wasDown = false
			}
			time.Sleep(20 * time.Millisecond)
			continue
		}

		cps := clicker.CPS()
		if cps <= 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		interval := (time.Second / time.Duration(cps)) - 5*time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		down := input.IsKeyDown(VK_LBUTTON)
		title := strings.ToLower(input.ForegroundTitle())
		if down && strings.Contains(title, "minecraft") {
			_ = input.SendLeftUp()
			<-time.After(time.Millisecond * 5)
			_ = input.SendLeftDown()
			wasDown = true
		} else if wasDown {
			_ = input.SendLeftUp()
			wasDown = false
		}
	}
}
