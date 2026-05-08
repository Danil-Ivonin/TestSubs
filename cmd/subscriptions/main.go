package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/Danil-Ivonin/TestSubs/docs"
	"github.com/Danil-Ivonin/TestSubs/internal/app"
)

// @title Subscriptions Aggregation API
// @version 1.0
// @description REST service for managing user online subscriptions and calculating subscription totals.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
func main() {
	if err := app.Run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "application error: %v\n", err)
		os.Exit(1)
	}
}
