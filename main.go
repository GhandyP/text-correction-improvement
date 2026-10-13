package main

import (
	"context"
	"log"
	"os"

	"github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper"
)

func main() {
	if err := adkhelper.Run(
		context.Background(),
		os.Args[1:],
		"case_09_text_correction_improvement",
		"A text-only starting point for correct and improve text.",
		"Correct the supplied text while preserving its meaning and voice. List material changes and ask about ambiguities instead of guessing.",
	); err != nil {
		log.Fatal(err)
	}
}
