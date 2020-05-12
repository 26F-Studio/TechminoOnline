package main

import (
	"container/list"
	"sync"
	"unsafe"
)

/*
#include "client.h"

// luatc_gctname identifies the gc object in techmino.
static const char* luatc_gctname = "techmino.client.gc";

// uintptr_t is the uintptr type exposed from go.
typedef __SIZE_TYPE__ uintptr_t;

// luatc_gcderefhandle will dereference the handle from the
// of the item on the lua stack, if the item cannot be
// dereferenced, an lua error will be generated instead.
static uintptr_t luatc_gcderefhandle(lua_State* L, int index) {
	luaL_checkudata(L, index, luatc_gctname);
	uintptr_t* ud = (uintptr_t*)lua_touserdata(L, index);
	return *ud;
}

// luatc_gcfreegohandle is the exported function to be
// called by the luatc_gcfreehandle.
extern void luatc_gcfreegohandle(uintptr_t);

// luatc_gcfreehandle will deallocate the gc object once
// the lua reference handle been gc-ed.
static int luatc_gcfreehandle(lua_State* L) {
	luatc_gcfreegohandle(luatc_gcderefhandle(L, -1));
	return 0;
}

// luatc_gcpushhandle will push and initialize the userdata
// after allocation of the gc handle.
static void luatc_gcpushhandle(lua_State* L, uintptr_t handle) {
	// Initialize the content inside the userdata.
	uintptr_t* ud = (uintptr_t*)
		lua_newuserdata(L, sizeof(uintptr_t));
	*ud = handle;

	// Initialize the metatable of the userdata.
	if(luaL_newmetatable(L, luatc_gctname)) {
		lua_pushcfunction(L, luatc_gcfreehandle);
		lua_setfield(L, -2, "__gc");
	}
	lua_setmetatable(L, -2);
}
*/
import "C"

// luaGcRootMutex is the mutex for modifying the gc root.
var luaGcRootMutex sync.Mutex

// luaGcRoot is the gc root for generating the gc binding.
//
// Since there'll not be too many items being created by
// the client connector, it is assumed not to cause too
// serious performance problems.
var luaGcRoot list.List

// luaGcItem is the item that is registered in the gc root.
type luaGcItem struct {
	// i is the interface object referenced by the root.
	i interface{}

	// f is the function called for garbage collection.
	f func()
}

// luaGcAlloc an identifier for the specified object,
// the returned identifier is safe to be stored in lua.
//
// This function is assumed to be invoked from the go side,
// and the corresponding lua state is always required.
func luaGcAlloc(L *C.lua_State, i interface{}, f func()) {
	luaGcRootMutex.Lock()
	defer luaGcRootMutex.Unlock()
	element := luaGcRoot.PushBack(luaGcItem{i, f})
	C.luatc_gcpushhandle(L,
		C.uintptr_t(uintptr(unsafe.Pointer(element))))
}

// luaGcLookup dereferences an interface from the pointer.
//
// This function is assumed to be invoked from the go side,
// and it will always call luaL_typerror if the specified
// object is not an gc reference object.
func luaGcLookup(L *C.lua_State, index int) interface{} {
	p := C.luatc_gcderefhandle(L, C.int(index))
	pp := unsafe.Pointer(uintptr(p))
	return ((*list.Element)(pp)).Value.(luaGcItem).i
}

//export luatc_gcfreegohandle
func luatc_gcfreegohandle(p C.uintptr_t) {
	luaGcRootMutex.Lock()
	defer luaGcRootMutex.Unlock()
	pp := unsafe.Pointer(uintptr(p))
	item := luaGcRoot.Remove((*list.Element)(pp)).(luaGcItem)
	item.f()
}
