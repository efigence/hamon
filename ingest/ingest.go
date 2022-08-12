package ingest

import (
	"github.com/efigence/go-haproxy"
	"go.uber.org/zap"
	"net"
	"strings"
)

type Config struct {
	ListenAddr string
	Logger     *zap.SugaredLogger
}

type Ingest struct {
	l *zap.SugaredLogger
}

func New(cfg Config) (*Ingest, chan haproxy.HTTPRequest, error) {
	i := Ingest{
		l: cfg.Logger,
	}
	ch := make(chan haproxy.HTTPRequest, 4096)

	la, err := net.ResolveUDPAddr("udp", cfg.ListenAddr)
	if err != nil {
		return nil, ch, err
	}
	c, err := net.ListenUDP("udp", la)
	if err != nil {
		return nil, ch, err
	}
	go i.ingestor(c, ch)
	go i.ingestor(c, ch)
	go i.ingestor(c, ch)
	go i.ingestor(c, ch)
	return &i, ch, nil
}

func (i *Ingest) ingestor(conn *net.UDPConn, ch chan haproxy.HTTPRequest) {
	buf := make([]byte, 65535)
	go func() {
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			_ = addr
			log_str := string(buf[0:n])
			if err != nil {
				i.l.Errorf("Error: %s", err)
			}
			if strings.Contains(log_str, " SSL handshake") {
				continue
			}
			req, err := haproxy.DecodeHTTPLog(log_str)
			ch <- req
		}
	}()

}
