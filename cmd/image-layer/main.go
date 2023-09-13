package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wzshiming/image-layer/pkg/cmd"
)

func main() {
	ctx := context.Background()
	c := cmd.NewCommand(ctx)
	_, err := c.ExecuteContextC(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
