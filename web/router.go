package web

import (
	"fmt"
	"github.com/efigence/go-mon"
	"github.com/efigence/hamon/stats"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"html/template"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"
)

type WebBackend struct {
	l     *zap.SugaredLogger
	r     *gin.Engine
	cfg   *Config
	stats *stats.Stats
}

type Config struct {
	Logger     *zap.SugaredLogger `yaml:"-"`
	ListenAddr string             `yaml:"listen_addr"`
	Stats      *stats.Stats
}

func New(cfg Config, webFS fs.FS) (backend *WebBackend, err error) {
	if cfg.Logger == nil {
		panic("missing logger")
	}
	if len(cfg.ListenAddr) == 0 {
		panic("missing listen addr")
	}
	w := WebBackend{
		l:     cfg.Logger,
		cfg:   &cfg,
		stats: cfg.Stats,
	}
	r := gin.New()
	w.r = r
	gin.SetMode(gin.ReleaseMode)
	t, err := template.ParseFS(webFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("error loading templates: %s", err)
	}
	r.SetHTMLTemplate(t)
	// for zap logging
	r.Use(ginzap.GinzapWithConfig(w.l.Desugar(), &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        false,
		SkipPaths: []string{
			"/_status/health",
			"/_status/metrics",
		},
	}))
	//r.Use(ginzap.RecoveryWithZap(w.l.Desugar(), true))
	// basic logging to stdout
	//r.Use(gin.LoggerWithWriter(os.Stdout))
	r.Use(gin.Recovery())

	// monitoring endpoints
	r.GET("/_status/health", gin.WrapF(mon.HandleHealthcheck))
	r.HEAD("/_status/health", gin.WrapF(mon.HandleHealthcheck))
	r.GET("/_status/metrics", gin.WrapF(mon.HandlePrometheus))
	// healthcheckHandler, haproxyStatus := mon.HandleHealthchecksHaproxy()
	// r.GET("/_status/metrics", gin.WrapF(healthcheckHandler))

	httpFS := http.FileServer(http.FS(webFS))
	r.GET("/s/*filepath", func(c *gin.Context) {
		// content is embedded under static/ dir
		p := strings.Replace(c.Request.URL.Path, "/s/", "/static/", -1)
		c.Request.URL.Path = p
		//c.Header("Cache-Control", "public, max-age=3600, immutable")
		httpFS.ServeHTTP(c.Writer, c.Request)
	})
	r.GET("/", func(c *gin.Context) {
		keys := make([]string, 0)
		for k := range w.stats.Frontends.TotalDurationMs {
			if k == "" {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":         c.Request.RemoteAddr,
			"frontend_list": keys,
		})
	})
	r.GET("/v1/gcstat", w.V1GCStats)
	r.GET("/v1/stats/frontend", w.V1FrontendStats)
	r.GET("/v1/stats/frontend/rate/:frontend", w.V1FrontendBackendStats)
	r.GET("/v1/stats/frontend/top/:name", w.V1FrontendTop)
	r.GET("/v1/stats/frontend/slow/:name", w.V1FrontendSlow)
	r.GET("/v1/stats/top_ip/:rate", w.V1TopRate)

	r.GET("/frontend/:name", w.Frontend)
	r.GET("/slow/:name", w.Slow)
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.tmpl", gin.H{
			"notfound": c.Request.URL.Path,
		})
	})
	go func() {
		for {
			updateGC()
			time.Sleep(time.Second)
		}
	}()

	return &w, nil
}

func (b *WebBackend) Run() error {
	b.l.Infof("listening on %s", b.cfg.ListenAddr)
	return b.r.Run(b.cfg.ListenAddr)
}
