package message

type QoS byte

const (
	QoSAtMostOnce QoS = iota
	QoSAtLeastOnce
	QoSExactlyOnce
)

type RetainHandling byte

const (
	RetainHandlingOnSubscribe RetainHandling = iota
	RetainHandlingOnSubscribeIfNew
	RetainDoNotSend
)
