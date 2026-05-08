package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Danil-Ivonin/TestSubs/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}
