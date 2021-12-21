package app

import (
	"aiscope/cmd/apiserver/app/options"
	"context"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewAPIServerCommand() *cobra.Command {
	s := options.NewServerRunOptions()

	cmd := &cobra.Command{
		Use: "apiserver",
		Long: `The API server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(s, signals.SetupSignalHandler())
		},
		SilenceUsage: true,
	}

	return cmd
}

func Run(s *options.ServerRunOptions, ctx context.Context) error {
	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(ctx.Done())
	if err != nil {
		return err
	}

	return apiserver.Run(ctx)
}
