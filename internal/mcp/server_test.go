// Copyright 2026 Eaoum AI
//
// SPDX-License-Identifier: Apache-2.0
//
// This file verifies MCP request handling and message framing.
package mcp

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestHandleInitialize(t *testing.T) {
	resp, ok := NewServer().handle([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	if !ok {
		t.Fatal("handle returned no response")
	}
	if resp.Error != nil {
		t.Fatalf("error = %#v", resp.Error)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result = %#v", resp.Result)
	}
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("protocolVersion = %#v, want %s", result["protocolVersion"], protocolVersion)
	}
}

func TestHandleNotificationDoesNotRespond(t *testing.T) {
	_, ok := NewServer().handle([]byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	if ok {
		t.Fatal("notification produced a response")
	}
}

func TestServeUsesMCPFraming(t *testing.T) {
	body := []byte(`{"jsonrpc":"2.0","id":"tools","method":"tools/list"}`)
	input := strings.NewReader("Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + string(body))
	var output bytes.Buffer

	if err := NewServer().Serve(input, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(output.String(), "Content-Length: ") {
		t.Fatalf("output is not framed: %q", output.String())
	}
	if !strings.Contains(output.String(), `"tools":[]`) {
		t.Fatalf("output does not contain empty tools list: %q", output.String())
	}
}
