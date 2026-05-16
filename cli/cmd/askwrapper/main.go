package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"module/lib/internal/askwrapper"
	"module/lib/internal/askwrapperui"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "askwrapper",
		Usage: "Go-native wrapper UI for ask workflows",
		Commands: []*cli.Command{
			{
				Name:  "ui",
				Usage: "Interactive UI commands",
				Commands: []*cli.Command{
					{
						Name:  "ask",
						Usage: "Ask a new question with history preview",
						Flags: []cli.Flag{
							&cli.IntFlag{
								Name:  "timeout",
								Usage: "Timeout in seconds for ask command",
								Value: int(askwrapper.DefaultTimeout.Seconds()),
							},
						},
						Action: func(ctx context.Context, command *cli.Command) error {
							seconds := command.Int("timeout")
							if seconds <= 0 {
								return fmt.Errorf("timeout must be greater than zero")
							}
							timeout := time.Duration(seconds) * time.Second
							return askwrapperui.RunAskMode(ctx, timeout)
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
