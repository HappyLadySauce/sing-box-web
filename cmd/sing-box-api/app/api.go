package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"sing-box-web/cmd/sing-box-api/app/options"
)

// sing-box-api 是 sing-box-web 的 api 服务，用于提供给 sing-box-web 使用
// 同时，sing-box-api 也是 sing-box-agent 和 sing-box-ctl 的 api 服务，用于提供给 sing-box-agent 和 sing-box-ctl 使用
func NewAPICommand(ctx context.Context) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:  "sing-box-api",
		Long: `The sing-box-api provide api for sing-box-web web ui and sing-box-agent and sing-box-ctl.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			// 验证选项，如果选项不合法，则返回错误
			if errs := opts.Validate(); len(errs) != 0 {
				return errs.ToAggregate()
			}
			// 运行 sing-box-api
			if err := run(ctx, opts); err != nil {
				return err
			}
			return nil
		},
		// 参数验证
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}
	// 直接用 pflag.NewFlagSet
	genericFlagSet := pflag.NewFlagSet("generic", pflag.ExitOnError)
	opts.AddFlags(genericFlagSet)

	logsFlagSet := pflag.NewFlagSet("logs", pflag.ExitOnError)
	klogflag.Add(logsFlagSet)

	cmd.Flags().AddFlagSet(genericFlagSet)
	cmd.Flags().AddFlagSet(logsFlagSet)
	// 返回命令
	return cmd
}

func run(ctx context.Context, opts *options.Options) error {
	return nil
}












