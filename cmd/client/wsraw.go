package main

import (
	"context"
	"errors"
	"net"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

/*
#cgo pkg-config: luajit
#include "lua.h"
*/
import "C"

// luaWebSocketConn is a connection of websocket with
// which data could be transferred and received between
// websocket.
type luaWebSocketConn struct {
	// conn is the connection established for websocket
	// communication.
	conn *websocket.Conn

	// sendMtx is the mutex for blocking the sending.
	sendMtx sync.Mutex

	// sendQueue for the websocket stream.
	sendQueue [][]byte

	// sendWaitCh is the channel for waiting for send
	// queue payloads.
	sendWaitCh chan struct{}

	// sendErr is the error sending back to the caller.
	sendErr error

	// receiveMtx is the mutex for blocking the receive.
	receiveMtx sync.Mutex

	// receiveQueue for the websocket stream.
	receiveQueue [][]byte

	// receiveErr is the error while running reader.
	receiveErr error

	// closeCh is the channel that is unblocked when
	// the websocket should close.
	closeCh chan struct{}
}

// runWebSocketWriter executes the websocket writer for
// a websocket connection.
func (wsconn *luaWebSocketConn) runWebSocketWriter() error {
	defer func() { _ = wsconn.conn.Close() }()
	for {
		// Wait for the socket closing or new content.
		select {
		case <-wsconn.closeCh:
			return errors.New("connection closed")
		case <-wsconn.sendWaitCh:
		}

		// Swap out the send queue content and write
		// out to the websocket writer.
		swappedSendQueue := func() (result [][]byte) {
			wsconn.sendMtx.Lock()
			defer wsconn.sendMtx.Unlock()
			result, wsconn.sendQueue = wsconn.sendQueue, nil
			wsconn.sendWaitCh = make(chan struct{})
			return
		}()

		// Attempt to write out to the writer.
		for _, item := range swappedSendQueue {
			err := websocket.Message.Send(wsconn.conn, item)
			if err != nil {
				return err
			}
		}
	}
}

// runWebSocketReader executes the websocket reader for
// a websocket connection.
func (wsconn *luaWebSocketConn) runWebSocketReader() error {
	for {
		select {
		case <-wsconn.closeCh:
			return errors.New("connection closed")
		default:
		}

		// Setup non-blocking reading deadline.
		ddl := time.Now().Add(5*time.Second)
		if err := wsconn.conn.SetReadDeadline(ddl); err != nil {
			return err
		}

		// Attempt to read frame from the frame.
		var data []byte
		if err := websocket.Message.Receive(wsconn.conn, &data); err != nil {
			// If the error is not of type net.Error, we will
			// return that error immediately.
			netErr, ok := err.(net.Error)
			if !ok {
				return err
			}

			// If the error is temporary error or timeout error,
			// we will treat it as if the nonblocking read
			// has reached its end.
			if netErr.Timeout() || netErr.Temporary() {
				continue
			}

			// For other cases, also return the error to caller.
			return netErr
		}

		// Append the item into the receive queue.
		func() {
			wsconn.receiveMtx.Lock()
			defer wsconn.receiveMtx.Unlock()
			wsconn.receiveQueue = append(wsconn.receiveQueue, data)
		}()
	}
}

// luaWebSocketReadResult is multiple frames read using
// the read interface of websocket connection.
type luaWebSocketReadResult struct {
	// frames are the received frames from the websocket.
	frames [][]byte
}

// marshal the websocket read result to lua stack.
func (r *luaWebSocketReadResult) marshal(L *C.lua_State) {
	luaTableNew(L, len(r.frames), 0)
	for i := 0; i < len(r.frames); i ++ {
		luaBytesPush(L, r.frames[i])
		luaTableRawSeti(L, -2, i + 1)
	}
}

// read implements the luaConn.read for luaWebSocketConn.
func (wsconn *luaWebSocketConn) read() (luaReadResult, error) {
	wsconn.receiveMtx.Lock()
	defer wsconn.receiveMtx.Unlock()
	readResult := &luaWebSocketReadResult{}
	readResult.frames, wsconn.receiveQueue = wsconn.receiveQueue, nil
	return readResult, wsconn.receiveErr
}

// write implements the luaConn.write for luaWebSocketConn.
func (wsconn *luaWebSocketConn) write(L *C.lua_State) error {
	// Attempt to read the pending frames on the lua stack.
	top := luaStackTopGet(L)
	var pendingFrames [][]byte
	for i := 2; i <= top; i ++ {
		if luaTypeOf(L, i) != luaTypeString {
			return errors.New("invalid argument type")
		}

		pendingFrames = append(pendingFrames, luaBytesGet(L, i))
	}

	// Emplace the read content to the writer goroutine.
	wsconn.sendMtx.Lock()
	defer wsconn.sendMtx.Unlock()
	if len(wsconn.sendQueue) == 0 {
		close(wsconn.sendWaitCh)
		wsconn.sendWaitCh = nil
	}
	wsconn.sendQueue = append(wsconn.sendQueue, pendingFrames...)
	return wsconn.sendErr
}

// close implements the luaConn.close for luaWebSocketConn.
func (wsconn *luaWebSocketConn) close() {
	close(wsconn.closeCh)
}

//export luatc_wsraw
func luatc_wsraw(L *C.lua_State) C.int {
	var config websocket.Config
	config.Version = websocket.ProtocolVersionHybi13

	// Make sure that the fields are valid for returning first.
	if luaTypeOf(L, 1) != luaTypeTable {
		luaNilPush(L)
		luaStringPush(L, "missing table argument")
		return C.int(2)
	}

	// Attempt to fetch the url field from the table.
	luaStringPush(L, "url")
	luaTableRawGet(L, 1)
	if luaTypeOf(L, -1) != luaTypeString {
		luaNilPush(L)
		luaStringPush(L, "missing url argument")
		return C.int(2)
	}
	argumentURL := luaStringGet(L, -1)
	luaStackPop(L, 1)

	// Attempt to parse the URL given at index.
	parsedURL, urlErr := url.Parse(argumentURL)
	if urlErr != nil {
		luaNilPush(L)
		luaStringPush(L, urlErr.Error())
		return C.int(2)
	}
	config.Location = parsedURL

	// Attempt to fetch the origin field from the table.
	luaStringPush(L, "origin")
	luaTableRawGet(L, 1)
	if typeOf := luaTypeOf(L, -1); typeOf != luaTypeString &&
		typeOf != luaTypeNone && typeOf != luaTypeNil {
		luaNilPush(L)
		luaStringPush(L, "invalid origin argument")
		return C.int(2)
	}

	// The origin argument will be dispensible and only
	// parsed when it is present.
	if luaTypeOf(L, -1) == luaTypeString {
		argumentOrigin := luaStringGet(L, -1)
		luaStackPop(L, 1)

		// Attempt to parse the URL given at index.
		parsedOrigin, urlErr := url.Parse(argumentOrigin)
		if urlErr != nil {
			luaNilPush(L)
			luaStringPush(L, urlErr.Error())
			return C.int(2)
		}
		config.Origin = parsedOrigin
	}

	// TODO: parse more fields in the table in the future.

	// Create the websocket connect task and return.
	luaTaskPush(L, func(ctx context.Context) (luaTaskResult, error) {
		var err error

		// Attempt to connect to the remote server with
		// provided configuration.
		conn, err := websocket.DialConfig(&config)
		if err != nil {
			return nil, err
		}

		// Create the connection instance and return.
		result := &luaWebSocketConn{
			conn:       conn,
			sendWaitCh: make(chan struct{}),
			closeCh:    make(chan struct{}),
		}
		go func() {
			err := result.runWebSocketWriter()
			result.sendMtx.Lock()
			defer result.sendMtx.Unlock()
			result.sendErr = err
		}()
		go func() {
			err := result.runWebSocketReader()
			result.receiveMtx.Lock()
			defer result.receiveMtx.Unlock()
			result.receiveErr = err
		}()
		return newLuaConnHandle(result), nil
	})
	luaNilPush(L)
	return C.int(2)
}
