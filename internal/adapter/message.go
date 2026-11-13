package adapter

type DaemonMessage interface {
	Payload() ([]byte, error)
	Unmarshal(data []byte) error
	Topic() string
	QoS() byte
	Retain() bool
}
