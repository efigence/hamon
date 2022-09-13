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
	"sort"
	"strings"
	"time"
)

var version string
var log *zap.SugaredLogger
var debug = false

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
		return lvl < zapcore.ErrorLevel && lvl > zapcore.DebugLevel
	})
	if debug {
		lowPriority = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		})
	}
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
		cli.BoolFlag{Name: "filter-private", Usage: "filter private IP classes"},
		cli.StringFlag{
			Name:  "address",
			Value: "http://127.0.0.1:3001",
			Usage: "address of hamon, without path",
		},
		cli.Float64Flag{
			Name:     "above",
			Usage:    "only add IPs wth req/s above that",
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
	filterPrivate := c.Bool("filter-private")
	url := c.String("address") + "/v1/stats/top_ip/" + c.String("above")
	log.Infof("using url %s", url)
	if c.Bool("daemon") {
		i := 0
		for {
			i++
			time.Sleep(time.Second)
			err := update(url, ipsetName, filterPrivate)
			if err != nil {
				log.Errorf("err: %s", err)
			}

			if i%128 == 1 {
				cleanup()
			}

		}
	} else {
		cleanup()
		return update(url, ipsetName, filterPrivate)
	}
	return nil
}

type TopIP struct {
	IPRate map[string]float64 `json:"ip_rate"`
}

func update(url, ipsetName string, filterPrivate bool) error {
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
		if filterPrivate && i.IsPrivate() {
			continue
		}
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
	current, _ := ipset.List(ipsetName)
	sort.SliceStable(ips, func(i, j int) bool {
		return ips[i].String() > ips[j].String()
	})
	sort.SliceStable(current, func(i, j int) bool {
		return current[i].String() > current[j].String()
	})
	if Equal(ips, current) {
		log.Debugf("current and new list are same, doing nothing")
		return nil
	}
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

func Filter[S any](s []S, f func(S) bool) []S {
	var accum []S
	for _, v := range s {
		if f(v) {
			accum = append(accum, v)
		}
	}
	return accum
}

func Equal(a []net.IP, b []net.IP) bool {
	if len(a) != len(b) {
		return false
	}
	for idx, _ := range a {
		if a[idx].String() != b[idx].String() {
			return false
		}
	}
	return true
}
func cleanup() {
	ipsets, err := ipset.ListIpsets()
	if err != nil {
		log.Errorf("error getting ipsets list: %s", err)
		return
	}
	currentPrefix := fmt.Sprintf("hamon-tmp-%s", time.Now().Format("060102"))
	tmpSets := Filter(ipsets, func(s string) bool {
		return strings.HasPrefix(s, "hamon-tmp")
	})
	toDelete := Filter(tmpSets, func(s string) bool {
		return !strings.HasPrefix(s, currentPrefix)
	})
	for _, v := range toDelete {
		log.Infof("destroying old top ipset %s", v)
		ipset.Destroy(v)
	}

}
