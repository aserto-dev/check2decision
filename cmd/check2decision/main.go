package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/check2decision/pkg/cmd"
)

const (
	rcOK  int = 0
	rcErr int = 1
)

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	os.Exit(run())
}

func run() (exitCode int) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cli := cmd.ConvertCmd{}

	kongCtx := kong.Parse(&cli,
		kong.Name("check2decision"),
		kong.Description("converts directory check assertions into authorizer check_decision assertions"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Summary:             false,
			Compact:             true,
			Tree:                false,
			FlagsLast:           true,
			Indenter:            kong.SpaceIndenter,
			NoExpandSubcommands: true,
		}),
		kong.Vars{},
	)

	kongCtx.BindTo(ctx, (*context.Context)(nil))

	if err := kongCtx.Run(); err != nil {
		exitErr(err)
	}

	return rcOK
}

func exitErr(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return rcErr
}
