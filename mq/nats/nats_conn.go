package natsmq

import (
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"time"
)

func setupDefaultNatsConnOptions(lg *zap.SugaredLogger, opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := 10 * time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, e error) {
		lg.Warnf("Disconnected due: %s. Will attempt reconnects for %.0fm", e, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		lg.Warnf("Successfullly reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		lg.Warnf("NATS connection closed: %v", nc.LastError())
	}))
	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, _ *nats.Subscription, err error) {
		lg.Errorf("Connection error: %s", err)
	}))

	return opts
}

func NewConn(c *Config) (*nats.Conn, error) {
	c.NatsConn = setupDefaultNatsConnOptions(c.Logger.Named("nats"), c.NatsConn)

	nc, err := nats.Connect(c.Broker, c.NatsConn...)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func NewEncodedConn(c *Config) (*nats.EncodedConn, error) {
	nc, err := NewConn(c)
	if err != nil {
		return nil, err
	}

	encoded, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, err
	} else {
		c.Logger.Infow("Successfully created nats.EncodedConn", "broker", c.Broker)
	}

	return encoded, nil
}
