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

var (
	runAskMode      = askwrapperui.RunAskMode
	runFollowUpMode = askwrapperui.RunFollowUpMode
)

func main() {
	cmd := buildCommand()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func buildCommand() *cli.Command {
	return &cli.Command{
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
							return runAskMode(ctx, timeout)
						},
					},
					{
						Name:  "followup",
						Usage: "Ask a follow-up question with selected history context",
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
							return runFollowUpMode(ctx, timeout)
						},
					},
				},
			},
		},
	}
}
