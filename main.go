package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/fatih/color"
	"github.com/go-ping/ping"
	"github.com/jessevdk/go-flags"
)

var pingu = []string{
	` ...        .     ...   ..    ..     .........           `,
	` ...     ....          ..  ..      ... .....  .. ..      `,
	` ...    .......      ...         ... . ..... BBBBBBB     `,
	`.....  ........ .BBBBBBBBBBBBBBB.....  ... BBBBBBBBBB.  .`,
	` .... ........BBBBBBBBBBBBBBBBBBBBB.  ... BBBBBBBBBBB    `,
	`      ....... BBWWWWBBBBBBBBBBBBBBBB.... BBBBBBBBBBBB    `,
	`.    .  .... BBWWBBWWBBBBBBBBBBWWWWBB... BBBBBBBBBBB     `,
	`   ..   ....BBBBWWWWBBRRRRRRBBWWBBWWB.. .BBBBBBBBBBB     `,
	`    .       BBBBBBBBRRRRRRRRRRBWWWWBB.   .BBBBBBBBBB     `,
	`   ....     .BBBBBBBBRRRRRRRRBBBBBBBB.      BBBBBBBB     `,
	`  .....      .  BBBBBBBBBBBBBBBBBBBB.        BBBBBBB.    `,
	`......     .. . BBBBBBBBBBBBBBBBBB . .      .BBBBBBB     `,
	`......       BBBBBBBBBBBBBBBBBBBBB  .      .BBBBBBB      `,
	`......   .BBBBBBBBBBBBBBBBBBYYWWBBBBB  ..  BBBBBBB       `,
	`...    . BBBBBBBBBBBBBBBBYWWWWWWWWWBBBBBBBBBBBBBB.       `,
	`       BBBBBBBBBBBBBBBBYWWWWWWWWWWWWWBBBBBBBBB .         `,
	`      BBBBBBBBBBBBBBBYWWWWWWWWWWWWWWWWBB    .            `,
	`     BBBBBBBBBBBBBBBYWWWWWWWWWWWWWWWWWWW  ........       `,
	`  .BBBBBBBBBBBBBBBBYWWWWWWWWWWWWWWWWWWWW    .........    `,
	` .BBBBBBBBBBBBBBBBYWWWWWWWWWWWWWWWWWWWWWW       .... . . `,
}

// nolint:gochecknoglobals
var (
	appName     = "pingu"
	appUsage    = "[OPTIONS] HOST"
	appVersion  = "???"
	appRevision = "???"
)

type exitCode int

const (
	exitCodeOK exitCode = iota
	exitCodeErrArgs
	exitCodeErrPing
)

type options struct {
	Version bool `short:"V" long:"version" description:"Show version"`
}

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", err)
	}

	os.Exit(int(code))
}

func run(cliArgs []string) (exitCode, error) {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = appName
	parser.Usage = appUsage

	args, err := parser.ParseArgs(cliArgs)
	if err != nil {
		if flags.WroteHelp(err) {
			return exitCodeOK, nil
		}

		return exitCodeErrArgs, fmt.Errorf("parse error: %w", err)
	}

	if opts.Version {
		// nolint:forbidigo
		fmt.Printf("%s: v%s-rev%s\n", appName, appVersion, appRevision)

		return exitCodeOK, nil
	}

	if 1 < len(args) {
		// nolint:goerr113
		return exitCodeErrArgs, errors.New("too many arguments")
	}

	pinger, err := initPinger(args[0])
	if err != nil {
		return exitCodeOK, fmt.Errorf("failed to init pinger: %w", err)
	}

	if err := pinger.Run(); err != nil {
		return exitCodeErrPing, fmt.Errorf("an error occurred when running ping: %w", err)
	}

	return exitCodeOK, nil
}

// nolint:forbidigo
func initPinger(host string) (*ping.Pinger, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("failed to init pinger %w", err)
	}

	// Listen for Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			pinger.Stop()
		}
	}()

	fmt.Printf(
		"PING %s (%s):\n",
		pinger.Addr(),
		pinger.IPAddr(),
	)

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%s seq=%s %sbytes from %s: ttl=%s time=%s\n",
			renderASCIIArt(pkt.Seq),
			color.New(color.FgHiYellow, color.Bold).Sprintf("%d", pkt.Seq),
			color.New(color.FgHiBlue, color.Bold).Sprintf("%d", pkt.Nbytes),
			color.New(color.FgHiBlue, color.Bold).Sprintf("%s", pkt.IPAddr),
			color.New(color.FgHiCyan, color.Bold).Sprintf("%d", pkt.Ttl),
			color.New(color.FgHiMagenta, color.Bold).Sprintf("%v", pkt.Rtt),
		)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf(
			"\n───── %s ping statistics ─────\n",
			stats.Addr,
		)
		fmt.Printf(
			"%s %v packets transmitted => %v packets received, (%v packet loss)\n",
			color.New(color.FgHiWhite, color.Bold).Sprintf("PACKET STATICS"),
			color.New(color.FgHiCyan, color.Bold).Sprintf("%d", stats.PacketsSent),
			color.New(color.FgHiBlue, color.Bold).Sprintf("%d", stats.PacketsRecv),
			color.New(color.FgHiRed, color.Bold).Sprintf("%v%%", stats.PacketLoss),
		)
		fmt.Printf(
			"%s: min=%v avg=%v max=%v stddev=%v\n",
			color.New(color.FgHiWhite, color.Bold).Sprintf("ROUND TRIP"),
			color.New(color.FgHiBlue, color.Bold).Sprintf("%v", stats.MinRtt),
			color.New(color.FgHiCyan, color.Bold).Sprintf("%v", stats.AvgRtt),
			color.New(color.FgHiGreen, color.Bold).Sprintf("%v", stats.MaxRtt),
			color.New(color.FgMagenta, color.Bold).Sprintf("%v", stats.StdDevRtt),
		)
	}

	return pinger, nil
}

func renderASCIIArt(idx int) string {
	if len(pingu) <= idx {
		return strings.Repeat(" ", len(pingu[0]))
	}

	line := pingu[idx]

	line = colorize(line, 'R', color.New(color.FgHiRed, color.Bold))
	line = colorize(line, 'Y', color.New(color.FgHiYellow, color.Bold))
	line = colorize(line, 'B', color.New(color.FgHiBlack, color.Bold))
	line = colorize(line, 'W', color.New(color.FgHiWhite, color.Bold))

	return line
}

func colorize(text string, target rune, color *color.Color) string {
	return strings.ReplaceAll(
		text,
		string(target),
		color.Sprint("#"),
	)
}
