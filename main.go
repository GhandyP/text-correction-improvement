// Command text-correction-improvement corrects a supplied text and reports what
// changed, why it changed, and which protected tokens survived.
//
// The CLI is wired incrementally by the tasks described in
// odd/tasks/text-correction-v1.md. Until T5 lands, this command reports that it
// is not wired yet instead of pretending to work.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "text-correction-improvement: not wired yet; see odd/tasks/text-correction-v1.md")
	os.Exit(2)
}
