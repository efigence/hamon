package main

import (
	"embed"
	"github.com/efigence/go-mon"
	ingest "github.com/efigence/hamon/ingest"
	"github.com/efigence/hamon/stats"
	"github.com/efigence/hamon/web"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var version string
var log *zap.SugaredLogger
var debug = true

// /* embeds with all files, just dir/ ignores files starting with _ or .
//
//go:embed static templates
var webContent embed.FS

func init() {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	// naive systemd detection. Drop timestamp if running under it
	if os.Getenv("INVOCATION_ID") != "" || os.Getenv("JOURNAL_STREAM") != "" {
		consoleEncoderConfig.TimeKey = ""
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	} else {
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleStderr := zapcore.Lock(os.Stderr)
	_ = consoleStderr
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, os.Stderr, lowPriority),
		zapcore.NewCore(consoleEncoder, os.Stderr, highPriority),
	)
	logger := zap.New(core)
	if debug {
		logger = logger.WithOptions(
			zap.Development(),
			zap.AddCaller(),
			zap.AddStacktrace(highPriority),
		)
	} else {
		logger = logger.WithOptions(
			zap.AddCaller(),
		)
	}
	log = logger.Sugar()

}

func main() {
	defer log.Sync()
	// register internal stats
	mon.RegisterGcStats()
	app := cli.NewApp()
	app.Name = "hamon"
	app.Description = "HAProxy logs monitor"
	app.Version = version
	app.HideHelp = true
	log.Errorf("Starting %s version: %s", app.Name, version)
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "help, h", Usage: "show help"},
		cli.StringFlag{
			Name:   "listen-addr",
			Value:  "127.0.0.1:3001",
			Usage:  "Listen addr",
			EnvVar: "LISTEN_ADDR",
		},
		cli.StringFlag{
			Name:  "debug-addr",
			Usage: "start debug server (pprof) on that [ip]:port",
			Value: "",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("help") {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}
		if len(c.String("debug-addr")) > 0 {
			go func() {
				log.Infof("Running debug/pprof on port %s", c.String("debug-port"))
				log.Errorf("Error when listening on debug port: %s", http.ListenAndServe(c.String("debug-addr"), nil))
				os.Exit(1)
			}()
		}

		ingest, reqCh, err := ingest.New(ingest.Config{
			ListenAddr: "127.0.0.1:50514",
			Logger:     log,
		})
		st := stats.New(reqCh)

		if err != nil {
			log.Panicf("%s", err)
		}
		_ = ingest
		w, err := web.New(web.Config{
			Logger:     log,
			ListenAddr: c.String("listen-addr"),
			Stats:      st,
		}, webContent)
		if err != nil {
			log.Panicf("error starting web listener: %s", err)
		}
		return w.Run()
	}
	app.Commands = []cli.Command{
		{
			Name:    "rem",
			Aliases: []string{"a"},
			Usage:   "example cmd",
			Action: func(c *cli.Context) error {
				log.Warnf("running example cmd")
				return nil
			},
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "example cmd",
			Action: func(c *cli.Context) error {
				log.Warnf("running example cmd")
				return nil
			},
		},
	}
	// to sort do that
	//sort.Sort(cli.FlagsByName(app.Flags))
	//sort.Sort(cli.CommandsByName(app.Commands))
	app.Run(os.Args)
}
