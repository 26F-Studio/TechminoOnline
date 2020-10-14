package main

/*
#cgo pkg-config: luajit
#include "client.h"

LUALIB_API int luaopen_client(lua_State* L) {
	luaL_Reg regs[] = {
		{ "poll", luatc_poll },
		{ "read", luatc_read },
		{ "write", luatc_write },
		{ "httpraw", luatc_httpraw },
		{ "wsraw", luatc_wsraw },
		{ NULL, NULL },
	};
    lua_createtable(L, 0, 0);
	luaL_register(L, NULL, regs);
	return 1;
}
*/
import "C"

// main is the pseudo main function that will be simply
// ignored in the buildmode c-shared.
func main() {
}
