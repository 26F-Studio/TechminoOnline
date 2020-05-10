#pragma once

/**
 * @file client.h
 * @brief Techmino Online Client Lua Binding
 *
 * This file defines the common function that should be included
 * by different lua files in this package, since it defines the
 * entrypoint and some common binding functions aggregating all
 * the defined files.
 */
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>

// luaopen_client is the library entry point function that will
// be called in 'require "client"' statement.
LUALIB_API int luaopen_client(lua_State* L);
