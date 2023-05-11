package tools

import (
	"os"
	"sync"
)

type exitFunc struct {
	id int64
	f  func()
}

var exitFuncs = []*exitFunc{}
var exitFuncsMutex sync.Mutex
var nextExitID int64

// AtExit registers the given function to be run when Exit() is called. It returns a cancel
// function that allows to remove the exit function.
func AtExit(f func()) (cancel func()) {
	if f == nil {
		return func() {}
	}

	exitFuncsMutex.Lock()
	defer exitFuncsMutex.Unlock()

	e := &exitFunc{id: nextExitID, f: f}
	nextExitID = nextExitID + 1

	exitFuncs = append(exitFuncs, e)

	return func() {
		exitFuncsMutex.Lock()
		defer exitFuncsMutex.Unlock()

		for i := range exitFuncs {
			if exitFuncs[i].id == e.id {
				exitFuncs = append(exitFuncs[:i], exitFuncs[i+1:]...)
				break
			}
		}
	}
}

// Exit runs all registered exit functions in reverse order of their registration and then uses os.Exit to exit with
// the given code.
func Exit(code int) {
	exitFuncsMutex.Lock()
	defer exitFuncsMutex.Unlock()

	for i := len(exitFuncs) - 1; i >= 0; i-- {
		exitFuncs[i].f()
	}
	os.Exit(code)
}
