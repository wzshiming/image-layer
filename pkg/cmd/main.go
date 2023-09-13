package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/wzshiming/image-layer/pkg/cmd/get"
	"github.com/wzshiming/image-layer/pkg/cmd/put"
)

// NewCommand returns a new cobra.Command for root
func NewCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Args:          cobra.NoArgs,
		Use:           "image-layer [command]",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		put.NewCommand(ctx),
		get.NewCommand(ctx),
	)
	return cmd
}
