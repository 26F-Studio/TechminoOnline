package main

import (
	"unsafe"
)

/*
#include "client.h"
#include <string.h>

// luatc_bytespush is the function called from the go side to
// push a slice with its length specified into lua side.
static void luatc_bytespush(lua_State* L, void* b, int len) {
	lua_pushlstring(L, (const char*)b, (size_t)len);
}

// luatc_pop uses the lua_pop macro and expose it to the go side.
static void luatc_pop(lua_State* L, int n) {
	lua_pop(L, n);
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

// luaBytesGet returns the lua lstring back to the go side.
func luaBytesGet(L *C.lua_State, index int) []byte {
	var length C.size_t
	bp := unsafe.Pointer(C.lua_tolstring(L, C.int(index), &length))
	if length == 0 {
		return nil
	} else {
		result := make([]byte, int(length))
		C.memcpy(unsafe.Pointer(&result[0]), bp, length)
		return result
	}
}

// luaStringPush uses the byte push function to push a C-string.
func luaStringPush(L *C.lua_State, s string) {
	luaBytesPush(L, []byte(s))
}

// luaStringGet returns the lua string back to the go side.
func luaStringGet(L *C.lua_State, index int) string {
	return string(luaBytesGet(L, index))
}

// luaNilPush will push a nil onto the lua stack.
func luaNilPush(L *C.lua_State) {
	C.lua_pushnil(L)
}

// luaIntegerPush pushes an integer to lua stack.
func luaIntegerPush(L *C.lua_State, value int) {
	C.lua_pushinteger(L, C.lua_Integer(value))
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

// luaStackPop pops several items from the lua stack.
func luaStackPop(L *C.lua_State, i int) {
	C.luatc_pop(L, C.int(i))
}

// luaType is the type of object returned by the luaTypeOf.
type luaType int

const (
	luaTypeNone = luaType(iota - 1)
	luaTypeNil
	luaTypeBoolean
	luaTypeLightUserdata
	luaTypeNumber
	luaTypeString
	luaTypeTable
	luaTypeFunction
	luaTypeUserdata
	luaTypeThread
)

// luaTypeOf fetches the type of value at given index.
func luaTypeOf(L *C.lua_State, index int) luaType {
	return luaType(int(C.lua_type(L, C.int(index))))
}

// luaTableNew creates a table with given number of initial
// numeric and hash slots, and push it on stack top.
func luaTableNew(L *C.lua_State, narr, nrec int) {
	C.lua_createtable(L, C.int(narr), C.int(nrec))
}

// luaTableRawGet fetches the table item with key at top.
func luaTableRawGet(L *C.lua_State, index int) {
	C.lua_rawget(L, C.int(index))
}

// luaTableRawSet updates the table item with kv pair at top.
func luaTableRawSet(L *C.lua_State, index int) {
	C.lua_rawset(L, C.int(index))
}

// luaTableRawGeti fetches the table item with numerical index.
func luaTableRawGeti(L *C.lua_State, index, n int) {
	C.lua_rawgeti(L, C.int(index), C.int(n))
}

// luaTableRawSeti update the table item with numerical index.
func luaTableRawSeti(L *C.lua_State, index, n int) {
	C.lua_rawseti(L, C.int(index), C.int(n))
}
