package main

import (
	"context"
	"fmt"
)

/*
#include "client.h"
*/
import "C"

// luaTaskResult is the result returned by a luaTaskFunc.
type luaTaskResult interface {
	// marshal the result back to the lua side, returning
	// exactly one result and exit.
	marshal(*C.lua_State)
}

// luaTaskHandle is a tracable and controllable task bind to
// the lua side. The lua side could execute poll for testing
// its status, or unref the task to terminate it.
type luaTaskHandle struct {
	// ctx controlling the execution of the task.
	ctx context.Context

	// cancel function to cancel the execution of task.
	cancel context.CancelFunc

	// result portion returned by the lua task function.
	result luaTaskResult

	// err portion returned by the lua task or recovered
	// from the last panic of task goroutine. The error must
	// be flatten so that the lua side can recognize it.
	err *string

	// completionCh channel used for completion notification.
	completionCh chan struct{}
}

// luaTask is the function that will be executed after the
// task control block pushsed onto the lua stack.
type luaTask func(context.Context) (luaTaskResult, error)

// luaTaskPush will initialize a task control block and
// push it onto the lua stack.
func luaTaskPush(L *C.lua_State, task luaTask) {
	// Create the task control block on go side.
	ctx, cancel := context.WithCancel(context.Background())
	taskHandle := &luaTaskHandle{
		ctx:          ctx,
		cancel:       cancel,
		completionCh: make(chan struct{}),
	}

	// Bind the task control block to the lua side.
	luaGcAlloc(L, taskHandle, taskHandle.cancel)

	// Startup the task execution goroutine.
	go func() {
		defer close(taskHandle.completionCh)
		defer func() {
			if err := recover(); err != nil {
				taskHandle.result = nil
				taskHandle.err = new(string)
				*taskHandle.err = fmt.Sprintf("%s", err)
			}
		}()
		var err error
		if taskHandle.result, err = task(ctx); err != nil {
			taskHandle.err = new(string)
			*taskHandle.err = err.Error()
		}
	}()
}

//export luatc_poll
func luatc_poll(L *C.lua_State) C.int {
	// First, attempt to cast the interface into a task.
	taskHandle, ok := luaGcLookup(L, -1).(*luaTaskHandle)
	if !ok {
		// return nil, "not main.luaTaskHandle"
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaStringPush(L, "not main.luaTaskHandle")
		return C.int(2)
	}

	// Second, attemp test the current task state.
	select {
	case <-taskHandle.ctx.Done():
		// return nil, "context canceled"
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaStringPush(L, "context canceled")
		return C.int(2)

	case <-taskHandle.completionCh:
		// return result, err
		luaStackTopSet(L, 0)
		if taskHandle.result != nil {
			taskHandle.result.marshal(L)
		} else {
			luaNilPush(L)
		}
		luaStackTopSet(L, 1)
		if taskHandle.err != nil {
			luaStringPush(L, *taskHandle.err)
		} else {
			luaNilPush(L)
		}
		return C.int(2)

	default:
		// return nil, nil
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaNilPush(L)
		return C.int(2)
	}
}
