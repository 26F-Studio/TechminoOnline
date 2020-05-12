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

/**
 * @brief result, err = client.poll(task)
 *
 * luatc_poll is the function that serves the client.poll on
 * the lua side. Each call to this function will poll for its
 * current status and returns two results:
 *
 * - When the task is not completed, <nil, nil> will be returned.
 * - When the task has been completed with result, <result, nil>
 *   will be returned, subsequent call returns the same result.
 * - When the task encounters error and is interrupted, the
 *   <nil, err> will be returned, subsequent call returns the
 *   same error. The error must be a string.
 */
LUALIB_API int luatc_poll(lua_State *L);

// luaopen_client is the library entry point function that will
// be called in 'require "client"' statement.
LUALIB_API int luaopen_client(lua_State* L);
