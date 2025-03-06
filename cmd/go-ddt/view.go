package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/xhd2015/data-driven-testing/decision_tree"
	"github.com/xhd2015/data-driven-testing/decision_tree/svg"
)

func handleView(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: go-ddt view <file>")
	}

	file := args[0]
	var tree *decision_tree.Node
	var err error
	if strings.HasSuffix(file, ".json") {
		var jsonData []byte
		jsonData, err = os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
		err = json.Unmarshal(jsonData, &tree)
	}
	if err != nil {
		return fmt.Errorf("failed to load tree: %v", err)
	}

	server := svg.NewServer(svg.NewRenderer(decision_tree.DefaultConfig()))
	return server.ServeFile(file)
}
