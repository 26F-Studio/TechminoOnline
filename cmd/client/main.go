package main

/*
#cgo pkg-config: luajit
#include "client.h"

LUALIB_API int luaopen_client(lua_State* L) {
	luaL_Reg regs[] = {
		{ "poll", luatc_poll },
		{ NULL, NULL },
	};
	luaL_newlib(L, regs);
	return 1;
}
*/
import "C"

// main is the pseudo main function that will be simply
// ignored in the buildmode c-shared.
func main() {
}
