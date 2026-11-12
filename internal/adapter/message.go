package adapter

type DaemonMessage interface {
	Payload() ([]byte, error)
	Topic() string
	QoS() byte
}
