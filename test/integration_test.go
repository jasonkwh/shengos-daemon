//go:build integration
// +build integration

package integration

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jasonkwh/envconfig"
	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/bus"
	clboard "github.com/jasonkwh/shengos-daemon/internal/clipboard"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

var cfgPath = "../config/config.yaml"

const peerHostnameSuffix = "-integration"

type integrationSuite struct {
	suite.Suite
	cfg      config.Config
	ctx      context.Context
	zl       *zap.Logger
	cancel   context.CancelFunc
	clb      adapter.Clipboard
	pubDBus  adapter.DaemonBus
	dbus     adapter.DaemonBus
	clPool   []io.Closer
	hostname string
	peerHost string
	received chan adapter.DaemonMessage
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(integrationSuite))
}

func (s *integrationSuite) SetupSuite() {
	cfgr := envconfig.New(&cfgPath, &s.cfg)
	cfgr.Load()

	s.ctx, s.cancel = context.WithCancel(s.Suite.T().Context())
	zl, err := zap.NewDevelopment()
	s.Require().NoError(err)
	s.zl = zl

	s.hostname, err = os.Hostname()
	s.Require().NoError(err)
	s.peerHost = s.hostname + peerHostnameSuffix
	s.received = make(chan adapter.DaemonMessage, 1)

	s.Require().NotEmpty(s.cfg.MQTTConn.Host, "MQTTConn.Host required; config file %q", cfgPath)

	s.clb, err = clboard.New(s.cfg.Clipboard, s.hostname,
		clboard.WithLogger(zl),
		clboard.WithSkipInit(true),
		clboard.WithWatchFunc(func(context.Context, ...clipboard.Option) <-chan clipboard.Data {
			changes := make(chan clipboard.Data, 1)
			changes <- clipboard.Data{Format: clipboard.FmtText, Bytes: []byte("test")}

			close(changes)

			return changes
		}))
	s.Require().NoError(err)

	subs := []message.Subscription{
		message.ClipboardSubscription,
	}

	s.pubDBus, err = bus.NewDaemonBus(s.ctx, s.cfg.MQTTConn, s.hostname,
		nil, nil, bus.WithLogger(s.zl))
	s.Require().NoError(err)
	s.clPool = append(s.clPool, s.pubDBus)

	s.dbus, err = bus.NewDaemonBus(s.ctx, s.cfg.MQTTConn, s.peerHost,
		map[string]adapter.DaemonTopicHandler{
			message.ClipboardTopic: suiteClipboardHandler{s: s},
		}, subs, bus.WithLogger(s.zl))
	s.Require().NoError(err)
	s.clPool = append(s.clPool, s.dbus)

	s.T().Log("integration suite setup complete")
}

func (s *integrationSuite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}

	if err := s.gracefulClose(); err != nil {
		s.T().Fatalf("failed to gracefully close services: %v", err)
	}

	s.T().Log("integration suite teardown complete")
}

func (s *integrationSuite) TestIntegration() {
	s.Run("daemon bus", func() {
		s.T().Log("daemon bus test")

		err := s.clb.Watch(s.ctx, s.pubDBus.Publish)
		s.Require().NoError(err)

		select {
		case got := <-s.received:
			clip, ok := got.(*message.ClipboardMessage)
			s.Require().True(ok)
			s.Equal(s.hostname, clip.Origin)
			s.Equal(message.ClipboardFormatText, clip.Format)
			s.Equal([]byte("test"), clip.Data)
		case <-time.After(5 * time.Second):
			s.Fail("timed out waiting for mqtt clipboard message")
		}

		s.T().Log("clipboard message received on bus")
	})
}
