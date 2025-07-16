package main

import (
	"context"
	"os"

	"sing-box-web/cmd/sing-box-web/app"
)

func main() {
	ctx := context.TODO()
	rootCmd := app.NewAPICommand(ctx)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}