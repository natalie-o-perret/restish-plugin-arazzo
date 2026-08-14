package main

import (
	"context"
	"fmt"
	"strings"

	az "github.com/pb33f/libopenapi/arazzo"
	"github.com/rest-sh/restish/v2/plugin"
)

type executor struct {
	client *plugin.CommandClient
	output any
}

func (e *executor) Execute(_ context.Context, request *az.ExecutionRequest) (*az.ExecutionResponse, error) {
	parts := strings.Split(request.OperationPath, "/")
	if !strings.Contains(request.OperationPath, "#/paths/") || len(parts) < 3 {
		return nil, fmt.Errorf("invalid operationPath %q", request.OperationPath)
	}
	path, method := strings.ReplaceAll(strings.ReplaceAll(parts[len(parts)-2], "~1", "/"), "~0", "~"), strings.ToUpper(parts[len(parts)-1])
	body := request.RequestBody
	if node, ok := body.(interface{ Decode(any) error }); ok {
		_ = node.Decode(&body)
	}
	response, err := e.client.Do(&plugin.HTTPRequestMsg{Method: method, URI: request.Source.Name + path, Body: body, ContentType: request.ContentType})
	if err != nil {
		return nil, err
	} else if response.Error != "" {
		return nil, fmt.Errorf("request: %s", response.Error)
	}
	e.output = response.Body
	return &az.ExecutionResponse{StatusCode: response.Status, Headers: response.Headers, Body: response.Body, URL: response.URL, Method: method}, nil
}
