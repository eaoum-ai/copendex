package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	protocolVersion = "2024-11-05"
	serverName      = "copendex"
	serverVersion   = "0.0.0"
)

type Server struct{}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewServer() Server {
	return Server{}
}

func (s Server) Serve(in io.Reader, out io.Writer) error {
	reader := bufio.NewReader(in)
	for {
		msg, err := readMessage(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		resp, ok := s.handle(msg)
		if !ok {
			continue
		}
		if err := writeMessage(out, resp); err != nil {
			return err
		}
	}
}

func (s Server) handle(msg []byte) (response, bool) {
	var req request
	if err := json.Unmarshal(msg, &req); err != nil {
		return response{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: -32700, Message: "parse error"},
		}, true
	}
	if len(req.ID) == 0 {
		return response{}, false
	}
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]any{
				"name":    serverName,
				"version": serverVersion,
			},
		}
	case "tools/list":
		resp.Result = map[string]any{
			"tools": []any{},
		}
	case "shutdown":
		resp.Result = nil
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found"}
	}
	return resp, true
}

func readMessage(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid MCP header: %s", line)
		}
		if strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, err
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, errors.New("missing MCP Content-Length header")
	}
	msg := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func writeMessage(out io.Writer, resp response) error {
	body, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	var framed bytes.Buffer
	fmt.Fprintf(&framed, "Content-Length: %d\r\n\r\n", len(body))
	framed.Write(body)
	_, err = out.Write(framed.Bytes())
	return err
}
