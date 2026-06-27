package main

import (
	"os"

	"github.com/addozhang/jira-cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
