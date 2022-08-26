package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/efigence/hamon/ipset"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

var version string
var log *zap.SugaredLogger
var debug = true

var httpclient = http.Client{
	Timeout: time.Second * 10,
}

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
	log.Infof("Starting %s version: %s", app.Name, version)
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
	app.Run(os.Args)
}

// max 31 chars
func getTmpNameIpset() string {
	r := make([]byte, 5)
	rand.Read(r)

	return fmt.Sprintf("hamon-tmp-%s-%x", time.Now().Format("060102"), r)
}

func mainApp(c *cli.Context) error {
	ipsetName := c.String("ipset-name")

	err := update("http://127.0.0.1:3001/stats/top_ip/0.1", ipsetName)
	if err != nil {
		log.Errorf("err: %s", err)
	}
	return nil
}

type TopIP struct {
	IPRate map[string]float64 `json:"ip_rate"`
}

func update(url, ipsetName string) error {
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	ips := TopIP{}
	err = json.Unmarshal(body, &ips)
	if err != nil {
		return err
	}
	ipList := []net.IP{}
	for ip, _ := range ips.IPRate {
		i := net.ParseIP(ip)
		if i == nil {
			log.Errorf("could not decode IP: %s", ip)
			continue
		}
		ipList = append(ipList, i)
	}
	return loader(ipsetName, ipList)
}

func loader(ipsetName string, ips []net.IP) error {
	////we just ignore errors here, no point checking if it exists
	ipset.Create(ipsetName, "hash:ip")
	tmpSet := getTmpNameIpset()
	err := ipset.Create(tmpSet, "hash:ip")
	// first before err check because better safe than littering
	defer func() {
		err := ipset.Destroy(tmpSet)
		if err != nil {
			log.Fatalf("error when removing temporary set [%s]: %s", tmpSet, err)
		}
	}()
	if err != nil {
		log.Fatalf("error adding temporary chain: %s", err)
	}
	list, err := ipset.List("hamon-swap")
	fmt.Printf("current: %+v\n", list)
	for _, ip := range ips {
		log.Infof("adding %s to %s", ipsetName, ip)
		err := ipset.Add(tmpSet, ip)
		if err != nil {
			log.Errorf("err: %s", err)
		}
	}
	err = ipset.Swap(tmpSet, ipsetName)
	return err
}
