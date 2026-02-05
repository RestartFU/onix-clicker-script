package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/restartfu/onix-winapi/internal/adapters/fileipc"
	"github.com/restartfu/onix-winapi/internal/adapters/winapi"
	"github.com/restartfu/onix-winapi/internal/core"
)

const defaultCPS = 12

func main() {
	clicker := core.NewClicker(defaultCPS)
	input := winapi.NewInput()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := core.Run(ctx, clicker, input); err != nil && err != context.Canceled {
			log.Printf("clicker stopped: %v", err)
		}
		stop()
	}()

	statePath := fileipc.DefaultStatePath()
	if value := os.Getenv("CLICKER_STATE_PATH"); value != "" {
		statePath = value
	}
	fileWatcher := fileipc.NewWatcher(statePath, 100*time.Millisecond)
	go fileWatcher.Run(ctx, clicker)
	log.Printf("file ipc watching %s", statePath)

	<-ctx.Done()
}
