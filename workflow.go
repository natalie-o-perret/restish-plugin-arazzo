package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
	az "github.com/pb33f/libopenapi/arazzo"
	"github.com/rest-sh/restish/v2/plugin"
)

func main() {
	plugin.Run(plugin.Manifest{Name: "workflow", Version: "0.1.0", RestishAPIVersion: 2, Hooks: []string{"command"}}, []plugin.CommandDecl{{Name: "workflow", Short: "Run Arazzo workflows"}}, run)
}
func run(_ string, args []string, client *plugin.CommandClient) error {
	if len(args) != 2 || args[0] != "run" {
		return fmt.Errorf("usage: restish workflow run FILE")
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	doc, err := libopenapi.NewArazzoDocument(raw)
	if err != nil {
		return fmt.Errorf("parse Arazzo document: %w", err)
	}
	if len(doc.SourceDescriptions) == 0 || len(doc.Workflows) != 1 {
		return fmt.Errorf("one-workflow Arazzo document with at least one source required")
	}
	workflow := doc.Workflows[0]
	var inputsDoc struct{ Required []string }
	var meta struct {
		Format  string   `yaml:"format"`
		Columns []string `yaml:"columns"`
	}
	if workflow.Inputs != nil {
		if err := workflow.Inputs.Decode(&inputsDoc); err != nil {
			return fmt.Errorf("decode workflow inputs: %w", err)
		}
	}
	inputs := map[string]any{}
	for _, name := range inputsDoc.Required {
		answer, err := client.Prompt(name+": ", false)
		if err != nil {
			return err
		} else if answer.Error != "" {
			return fmt.Errorf("prompt: %s", answer.Error)
		}
		inputs[name] = answer.Value
	}
	if workflow.Extensions != nil {
		if node := workflow.Extensions.GetOrZero("x-restish-workflow"); node != nil {
			if err := node.Decode(&meta); err != nil {
				return fmt.Errorf("decode x-restish-workflow: %w", err)
			}
		}
	}
	exec := &executor{client: client}
	sources, err := loadSources(client, doc)
	if err != nil {
		return err
	}
	if result := az.Validate(doc); result != nil && result.HasErrors() {
		return result
	}
	result, err := az.NewEngine(doc, exec, sources).RunWorkflow(context.Background(), workflow.WorkflowId, inputs)
	if err != nil {
		return fmt.Errorf("run workflow: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("workflow failed: %v", result.Error)
	}
	if meta.Format == "table" && os.Getenv("RSH_OUTPUT_FORMAT") == "" {
		formatted, err := table(exec.output, meta.Columns)
		if err != nil {
			return err
		}
		return client.WriteStdout(formatted)
	}
	return client.Response(200, nil, exec.output)
}
