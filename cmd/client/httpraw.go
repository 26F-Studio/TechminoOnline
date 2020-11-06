package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
)

/*
#include "client.h"
*/
import "C"

// rawClient is the tcp client which is shared among raw requests.
var rawClient http.Client

// httpRawResponse is the task result which should be written back
// to the caller side for reading.
//
// TODO: the header fields should also be copied here, but for
// rapid demo we've ignored them.
type httpRawResponse struct {
	// statusCode of the response.
	statusCode int

	// status string of the response.
	status string

	// header of the response that is received.
	header http.Header

	// body of the response that is received.
	body []byte
}

// marshal the http response back to the lua side.
func (r httpRawResponse) marshal(L *C.lua_State) {
	luaTableNew(L, 0, 3)

	// Set the result.code field.
	luaStringPush(L, "code")
	luaIntegerPush(L, r.statusCode)
	luaTableRawSet(L, -3)

	// Set the result.status field.
	luaStringPush(L, "status")
	luaStringPush(L, r.status)
	luaTableRawSet(L, -3)

	// Set the result.header field.
	luaStringPush(L, "header")
	luaPushHttpHeader(L, r.header)
	luaTableRawSet(L, -3)

	// Set the result.body field.
	luaStringPush(L, "body")
	luaBytesPush(L, r.body)
	luaTableRawSet(L, -3)
}

//export luatc_httpraw
func luatc_httpraw(L *C.lua_State) C.int {
	// Make sure that the fields are valid for returning first.
	if luaTypeOf(L, 1) != luaTypeTable {
		luaNilPush(L)
		luaStringPush(L, "missing table argument")
		return C.int(2)
	}

	// Attempt to fetch the request method.
	parsedMethod := "GET"
	luaStringPush(L, "method")
	luaTableRawGet(L, 1)
	if luaTypeOf(L, -1) == luaTypeString {
		parsedMethod = luaStringGet(L, -1)
	}
	luaStackPop(L, 1)

	// Attempt to fetch the url field from the table.
	luaStringPush(L, "url")
	luaTableRawGet(L, 1)
	if luaTypeOf(L, -1) != luaTypeString {
		luaNilPush(L)
		luaStringPush(L, "missing url argument")
		return C.int(2)
	}
	argumentURL := luaStringGet(L, -1)
	luaStackPop(L, 1)

	// Attempt to parse the URL given at index.
	parsedURL, urlErr := url.Parse(argumentURL)
	if urlErr != nil {
		luaNilPush(L)
		luaStringPush(L, urlErr.Error())
		return C.int(2)
	}

	// Attempt to parse the http header at index.
	luaStringPush(L, "header")
	luaTableRawGet(L, 1)
	parsedHeader, headerErr := luaReadHttpHeader(L, -1)
	luaStackPop(L, 1)
	if headerErr != nil {
		luaNilPush(L)
		luaStringPush(L, headerErr.Error())
	}

	// Attempt to fetch the request body
	luaStringPush(L, "body")
	luaTableRawGet(L, 1)
	if luaTypeOf(L, -1) == luaTypeString {
		parsedBody := luaStringGet(L, -1)
	}
	luaStackPop(L, 1)

	// TODO: parse more arguments from the request, now we omit
	// them since we want a quick demo.

	// Create the request handle and return.
	luaTaskPush(L, func(ctx context.Context) (luaTaskResult, error) {
		var err error

		// Initialize the request with parsed arguments.
		var request http.Request
		request.Method = parsedMethod
		request.URL = parsedURL
		request.Header = parsedHeader
		if parsedBody != nil {
			request.Body = parsedBody
		}
		// Perform the task request with the raw client.
		response, err := rawClient.Do(request.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		// Receive the response body and status from caller.
		var receiver bytes.Buffer
		_, err = io.Copy(&receiver, response.Body)
		if err != nil {
			return nil, err
		}

		// Return the collected result to the caller.
		return httpRawResponse{
			status:     response.Status,
			statusCode: response.StatusCode,
			header:     response.Header,
			body:       receiver.Bytes(),
		}, nil
	})
	luaNilPush(L)
	return C.int(2)
}
