package core

import (
	"context"
	"strings"
	"time"
)

func Run(ctx context.Context, clicker *Clicker, input InputPort) error {
	wasInjected := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !clicker.Enabled() {
			if wasInjected {
				_ = input.SendLeftUp()
				wasInjected = false
			}
			<-time.After(20 * time.Millisecond)
			continue
		}

		cps := clicker.CPS()
		if cps <= 0 {
			<-time.After(50 * time.Millisecond)
			continue
		}

		interval := (time.Second / time.Duration(cps)) - 5*time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		down := input.IsMouseDown()
		title := strings.ToLower(input.ForegroundTitle())
		if down && strings.Contains(title, "minecraft") {
			_ = input.SendLeftUp()
			<-time.After(time.Millisecond * 5)
			_ = input.SendLeftDown()
			wasInjected = true
		} else if wasInjected {
			_ = input.SendLeftUp()
			wasInjected = false
		}
	}
}
