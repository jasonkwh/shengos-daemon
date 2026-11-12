package message

type QoS byte

const (
	QoSAtMostOnce QoS = iota
	QoSAtLeastOnce
	QoSExactlyOnce
)
