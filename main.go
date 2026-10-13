// Command text-correction-improvement corrects a supplied text and reports what
// changed, why it changed, and which protected tokens survived.
//
// Usage and exit codes are documented in the help text and in
// odd/tasks/text-correction-v1.md.
package main

import (
	"context"
	"os"

	"github.com/GhandyP/text-correction-improvement/internal/agent"
	"github.com/GhandyP/text-correction-improvement/internal/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), build, os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func build(ctx context.Context, model string) (cli.Corrector, error) {
	return agent.New(ctx, agent.Config{Model: model})
}
