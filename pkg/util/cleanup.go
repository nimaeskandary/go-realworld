package util

import (
	"context"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
)

// CleanupManager is a utility to keep a list of cleanup functions to be executed
// in reverse order
type CleanupManager interface {
	// RegisterCleanupFunc registers a cleanup function to be called on Cleanup
	RegisterCleanupFunc(fn func())
	// Cleanup calls all registered cleanup functions in reverse order. Cleanup is
	// safe to run multiple times. It is guaranteed to run cleanup fns only once
	Cleanup()
}

type cleanupManagerImpl struct {
	once              sync.Once
	cleanupFns        []func()
	stopSignalHandler context.CancelFunc
}

// NewCleanupManager - If listenForExitSignal is true,
// the CleanupManager will listen for OS exit signals and call Cleanup automatically when such signal is received.
// The usecase here is to allow for graceful shutdown of the application when it recieves an exit signal.
// It is safe to also defer a call to CleanupManager.Cleanup.
func NewCleanupManager(ctx context.Context, listenForExitSignal bool) CleanupManager {
	cm := &cleanupManagerImpl{
		cleanupFns: make([]func(), 0),
	}

	if listenForExitSignal {
		signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
		cm.stopSignalHandler = stop
		go func() {
			<-signalCtx.Done()
			cm.Cleanup()
		}()
	}

	return cm
}

func (cm *cleanupManagerImpl) RegisterCleanupFunc(fn func()) {
	cm.cleanupFns = append(cm.cleanupFns, fn)
}

func (cm *cleanupManagerImpl) Cleanup() {
	cm.once.Do(func() {
		for _, fn := range slices.Backward(cm.cleanupFns) {
			fn()
		}
		if cm.stopSignalHandler != nil {
			cm.stopSignalHandler()
		}
	})
}
