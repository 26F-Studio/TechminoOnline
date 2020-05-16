package main

import (
	"runtime"
)

/*
#cgo pkg-config: luajit
#include "lua.h"
*/
import "C"

// luaReadResult is the data returned by the client.read.
// Please notice it should always occupies at most one lua
// stack slot.
type luaReadResult interface {
	// marshal the result into the lua stack, occupying
	// exactly one slot.
	marshal(*C.lua_State)
}

// luaConn is the connection that could be manipulated
// using the client.read and client.write function.
//
// XXX: When a instance of lua connection is created, it
// must always be wrapped into a luaConnHandle as soon
// as possible, to avoid potential memory leak.
type luaConn interface {
	// read nonblockingly reads some from the connection,
	// it is exposed by the client.read functino.
	read() (luaReadResult, error)

	// write nonblockingly writes to the connection, it is
	// exposed by the client.write function.
	write(*C.lua_State) error

	// close attempt to close the connection, it is only
	// called when there's no reference from lua side.
	close()
}

// luaConnHandle is the controllable connection bind to
// the lua side. The lua side could execute read and
// write for communication, or unref the connection to
// close the connection.
type luaConnHandle struct {
	// conn is the connection backed by this handle.
	conn luaConn
}

// finalizeLuaConnHandle is the go finalizer for the
// luaConnHandle set through runtime.SetFinalizer.
func finalizeLuaConnHandle(c *luaConnHandle) {
	if c.conn != nil {
		c.conn.close()
		c.conn = nil
	}
}

// newLuaConnHandle creates the lua connection handle
// which could be returned as luaTaskResult.
func newLuaConnHandle(c luaConn) *luaConnHandle {
	handle := &luaConnHandle{
		conn: c,
	}
	runtime.SetFinalizer(handle, finalizeLuaConnHandle)
	return handle
}

// marshal the luaConnHandle as userdata when it
// should be returned as task result.
func (c *luaConnHandle) marshal(L *C.lua_State) {
	// XXX: since the connection object has a relatively
	// complex state, it is managed on the go side using
	// go finalizer by accounting reachability from the
	// luaGcRoot (referenced as userdata or in luaTask).
	luaGcAlloc(L, c, nil)
}

//export luatc_read
func luatc_read(L *C.lua_State) C.int {
	// First, attempt to cast the interface into a conn.
	connHandle, ok := luaGcLookup(L, 1).(*luaConnHandle)
	if !ok {
		// return nil, "not main.luaConnHandle"
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaStringPush(L, "not main.luaConnHandle")
		return C.int(2)
	}

	// Second, attempt to invoke the read method of
	// the connection.
	result, err := connHandle.conn.read()
	if err != nil {
		// return nil, err
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaStringPush(L, err.Error())
		return C.int(2)
	}

	// Third, push back the result normally.
	// return result, nil.
	luaStackTopSet(L, 0)
	result.marshal(L)
	luaNilPush(L)
	return C.int(2)
}

//export luatc_write
func luatc_write(L *C.lua_State) C.int {
	// First, attempt to cast the interface into a conn.
	connHandle, ok := luaGcLookup(L, 1).(*luaConnHandle)
	if !ok {
		// return nil, "not main.luaConnHandle"
		luaStackTopSet(L, 0)
		luaNilPush(L)
		luaStringPush(L, "not main.luaConnHandle")
		return C.int(2)
	}

	// Second, attempt to invoke the write method
	// of the connection.
	err := connHandle.conn.write(L)

	// Third, push back the result normally.
	luaStackTopSet(L, 0)
	if err != nil {
		// return err
		luaStringPush(L, err.Error())
	} else {
		// return nil
		luaNilPush(L)
	}
	return C.int(2)
}
