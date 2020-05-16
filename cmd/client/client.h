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
 * @brief data, err = client.read(conn)
 *
 * luatc_read is the function that serves the client.read on
 * the lua side. Each call to this function will nonblockingly
 * read some contents from connection and return to the caller.
 *
 * - When the data available for read in the connection is
 *   <data>, this function will returns <data, nil>.
 * - When there's currently no more data available in the
 *   connection, <nil, nil> will be returned.
 * - When the connection has been closed with some unrecoverable
 *   error <err>, then <nil, err> will be returned. The error
 *   must be a string.
 * - This function might also returns <{}, nil> if this function
 *   returns <{data1, data2, ...}, nil> as normal form of result.
 *
 * If the caller unref the connection with data available for
 * read, those data will be discarded automatically.
 */
LUALIB_API int luatc_read(lua_State* L);

/**
 * @brief err = client.write(conn, ...)
 *
 * luatc_write is the function that serves the client.write on
 * the lua side. Each call to this function will nonblockingly
 * write some content to the connection.
 *
 * - When the connection is not closed, calling this function
 *   will pass all content followed by the connection userdata
 *   to the stream function. And if the argument is invalid
 *   then error will be returned to the caller (if any). And
 *   nil will be returned if the write success.
 * - When the connection is closed, the connection close causing
 *   will be returned to the caller.
 * - All errors returned must be a string.
 *
 * If the caller unref the stream with data pending to send,
 * the default behaviour is that the stream closes after all
 * data has sent. If there's any exception, they should explicit
 * point out in the connection initialization function.
 */
LUALIB_API int luatc_write(lua_State* L);

/**
 * reqtask, err = client.httpraw({
 *     "url" = url,       -- http or https url
 *     "method" = method, -- GET, POST, PUSH, etc. (default GET)
 *     "header" = {
 *     },                 -- HTTP request header (nullable)
 *     "body" = body      -- Long string of request content (nullable)
 * })
 *
 * luatc_httpraw creates a lua task where lua side could poll
 * for the completion of this task using the poll interface.
 * The error could still be generated when the request is malformed,
 * such as missing or invalid url, unsupported method, etc.
 *
 * The result received from the task should be:
 *
 * {
 *     "code" = code,     -- status code, like 200, 404
 *     "status" = status, -- status string, like '200 OK', '404 Not Found'.
 *     "header" = {
 *     },                 -- HTTP response header
 *     "body" = body      -- Long string of received content
 * }, err = client.poll(reqtask)
 *
 * XXX: this function is not intended for developing networking
 * in techmino, it is just used for demonstrating how the task
 * mechanism works, and intended for temporary http access.
 */
LUALIB_API int luatc_httpraw(lua_State* L);

/**
 * wsconntask, err = client.wsraw({
 *     "url" = url,       -- http, https, ws or wss url
 *     "origin" = origin, -- origin url (nullable)
 *     "header" = {
 *     },                 -- HTTP request header (nullable)
 * })
 *
 * luatc_wsraw creates a lua task attempting to connect to
 * the websocket server where lua side could poll for completion
 * using the poll interface. The error could still be generated
 * when the request is malformed, such as missing or invalid
 * url, invalid origin, etc.
 *
 * The result retrieved from the task should be:
 *
 * wsconn, err = client.poll(wsconntask)
 *
 * The wsconn obeys the conn interface and send the strings
 * arguments passed in as websocket frames. The received
 * websocket frames would also be strings:
 *
 * { frame1, frame2, ... }, err = client.read(wsconn)
 * err = client.write(wsconn, frame1, frame2, ...)
 *
 * The wsconn closes when there's no reference on lua side.
 *
 * XXX: this function is not intended for developing networking
 * in techmino, it is just used for benchmarking websocket
 * connections, and intended for temporary websocket access.
 */
//LUALIB_API int luatc_wsraw(lua_State* L);

// luaopen_client is the library entry point function that will
// be called in 'require "client"' statement.
LUALIB_API int luaopen_client(lua_State* L);
