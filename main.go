package main

import (
	"context"

	"auction/cmd"
	"auction/internal/shared/modules/config"
	"auction/internal/shared/modules/telemetry"
)

func main() {
	config.Init()

	// Install tracing (stdout exporter; safe with no collector present).
	shutdown := telemetry.InitTracerProvider(context.Background(), "auction")
	defer func() { _ = shutdown(context.Background()) }()

	cmd.Execute()
}
