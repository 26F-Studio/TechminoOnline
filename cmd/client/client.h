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

/**
 * reqtask, err = client.httpraw({
 *     "url": url,       -- http or https url
 *     "method": method, -- GET, POST, PUSH, etc. (default GET)
 *     "header": {
 *     },                -- Table map from key to value (nullable)
 *     "body": body      -- Long string of request content (nullable)
 * })
 *
 * luatc_httpraw creates a lua task where lua side could poll
 * for the completion of this task using the poll interface.
 * The error could still be generated when the request is malformed,
 * such as missing or invalid url, unsupported method, etc.
 *
 * The result received from the request task should be:
 *
 * {
 *     "code": code,     -- status code, like 200, 404
 *     "status": status, -- status string, like '200 OK', '404 Not Found'.
 *     "header": {
 *     },                -- Table map from key to value
 *     "body": body      -- Long string of received content
 * }, err = client.poll(reqtask)
 *
 * XXX: this function is not intended for developing networking
 * in techmino, it is just used for demonstrating how the task
 * mechanism works, and intended for temporary http access.
 */
LUALIB_API int luatc_httpraw(lua_State* L);

// luaopen_client is the library entry point function that will
// be called in 'require "client"' statement.
LUALIB_API int luaopen_client(lua_State* L);
