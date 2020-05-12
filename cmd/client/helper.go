package main

import (
	"unsafe"
)

/*
#include "client.h"

// luatc_bytespush is the function called from the go side to
// push a slice with its length specified into lua side.
static void luatc_bytespush(lua_State* L, void* b, int len) {
	lua_pushlstring(L, (const char*)b, (size_t)len);
}
*/
import "C"

// luaBytesPush pushes the bytes slice as a lua lstring.
func luaBytesPush(L *C.lua_State, b []byte) {
	if len(b) > 0 {
		C.luatc_bytespush(L, unsafe.Pointer(&b[0]), C.int(len(b)))
	} else {
		C.luatc_bytespush(L, unsafe.Pointer(nil), C.int(0))
	}
}

// luaStringPush uses the byte push function to push a C-string.
func luaStringPush(L *C.lua_State, s string) {
	luaBytesPush(L, []byte(s))
}

// luaNilPush will push a nil onto the lua stack.
func luaNilPush(L *C.lua_State) {
	C.lua_pushnil(L)
}

// luaStackTopGet returns the current lua stack top.
func luaStackTopGet(L *C.lua_State) int {
	top := C.lua_gettop(L)
	return int(top)
}

// luaStackTopSet updates the current lua stack top.
func luaStackTopSet(L *C.lua_State, newtop int) {
	C.lua_settop(L, C.int(newtop))
}
