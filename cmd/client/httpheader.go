package main

import (
	"fmt"
	"net/http"
)

/*
#include "client.h"
*/
import "C"

// luaPushHttpHeader pushes HTTP header onto stack top.
func luaPushHttpHeader(L *C.lua_State, hdr http.Header) {
	// Pushes a new table on the stack first.
	luaTableNew(L, 0, 0)
	if hdr == nil {
		return
	}

	// Pushes all keys in the header to the stack.
	for k := range hdr {
		luaStringPush(L, k)
		luaStringPush(L, hdr.Get(k))
		luaTableRawSet(L, -3)
	}
}

// luaReadHttpHeader attempts to read the conten specified
// by stack index into the http request.
func luaReadHttpHeader(L *C.lua_State, idx int) (http.Header, error) {
	// Ensures the table to visit is of valid type.
	result := make(http.Header)
	luaType := luaTypeOf(L, idx)
	switch luaType {
	case luaTypeTable:
	case luaTypeNil:
		return result, nil
	default:
		// XXX: don't use it inside an lua_next loop.
		return nil, fmt.Errorf(
			"invalid header %s", luaStringGet(L, idx))
	}

	// Save the stack index for resuming after returning.
	stackTop := luaStackTopGet(L)
	defer luaStackTopSet(L, stackTop)

	// Reverse the direction of stack indexing since
	// we will be using minus index while reading,
	// causing wrong table to be accessed.
	if idx < 0 {
		idx = stackTop + idx + 1
	}

	// Attempt to visit all table entries in the map.
	luaNilPush(L)
	for luaTableNext(L, idx) {
		key := luaStringGet(L, -1)
		value := luaStringGet(L, -2)
		if luaTypeOf(L, -2) != luaTypeString &&
			luaTypeOf(L, -1) != luaTypeString {
			return nil, fmt.Errorf(
				"invalid header item[%s] = %s", key, value)
		}
		result.Set(key, value)
		luaStackPop(L, 1)
	}
	return result, nil
}
