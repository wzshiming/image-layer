package put

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	image_layer "github.com/wzshiming/image-layer/pkg/registry"
)

type flagpole struct {
	username string
	password string
	insecure bool
}

func NewCommand(ctx context.Context) *cobra.Command {
	var flags flagpole

	cmd := &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   "put image:tag [path-to-output or - for stdout]",
		Short: "put is a tool to put the image layer",
		RunE: func(cmd *cobra.Command, args []string) error {
			var opts []image_layer.Option
			if flags.username != "" {
				opts = append(opts, image_layer.WithUserPass(flags.username, flags.password))
			}
			if flags.insecure {
				opts = append(opts, image_layer.WithInsecure(flags.insecure))
			}
			cli, err := image_layer.NewClient(args[0], opts...)
			if err != nil {
				return err
			}

			var r io.ReadCloser
			if args[1] != "-" {
				f, err := os.Open(args[1])
				if err != nil {
					return err
				}
				defer f.Close()
				r = f
			} else {
				r = os.Stdin
			}

			err = cli.Put(cmd.Context(), r)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.username, "username", "u", "", "username")
	cmd.Flags().StringVarP(&flags.password, "password", "p", "", "password")
	cmd.Flags().BoolVar(&flags.insecure, "insecure", false, "insecure")

	return cmd
}
