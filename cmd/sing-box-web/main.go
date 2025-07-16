package main

import (
	"context"
	"os"

	"k8s.io/component-base/cli"

	"github.com/karmada-io/sing-box-web/cmd/sing-box-web/app"
)

func main() {
	ctx := context.TODO()
	cmd := app.NewAPICommand(ctx)
	code := cli.Run(cmd)
	os.Exit(code)
}