package main

import (
	"crypto/rand"
	"fmt"
	"github.com/efigence/hamon/ipset"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var version string
var log *zap.SugaredLogger
var debug = true

func init() {
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	// naive systemd detection. Drop timestamp if running under it
	if os.Getenv("INVOCATION_ID") != "" || os.Getenv("JOURNAL_STREAM") != "" {
		consoleEncoderConfig.TimeKey = ""
	}
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
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
	app := cli.NewApp()
	app.Name = "hamon-ipset-loader"
	app.Description = "Load hamon top list into specified ipset"
	app.Version = version
	app.HideHelp = true
	log.Errorf("Starting %s version: %s", app.Name, version)
	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "help, h", Usage: "show help"},
		cli.BoolFlag{Name: "daemon", Usage: "daemonize"},
		cli.StringFlag{
			Name:  "address",
			Value: "http://127.0.0.1:3001",
			Usage: "address of hamon, without path",
		},
		cli.Float64Flag{
			Name:     "above",
			Usage:    "only add IPs above the range",
			Required: true,
		},
		cli.StringFlag{
			Name:  "ipset-name",
			Usage: "name of IPset to swap",
			Value: "hamon-blocked",
		},
	}
	app.Action = func(c *cli.Context) error {
		return mainApp(c)
	}
}

func getTmpNameIpset() string {
	r := make([]byte, 8)
	rand.Read(r)

	return fmt.Sprintf("hamon-tmp-%s-%x", time.Now().Format("20060102"), r)
}

func mainApp(c *cli.Context) error {
	ipsetName := c.String("ipset-name")
	loader(ipsetName)
	return nil
}

func loader(ipsetName string) error {
	//we just ignore errors here, no point checking if it exists
	ipset.Create(ipsetName, "hash:ip")
	tmpSet := getTmpNameIpset()
	err := ipset.Create(tmpSet, "hash_ip")
	if err != nil {
		log.Fatalf("error adding temporary chain: %s", err)
	}
	defer ipset.Destroy(tmpSet)
	return err
}
