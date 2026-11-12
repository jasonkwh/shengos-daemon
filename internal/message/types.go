package message

type Subscription struct {
	Topic             string
	QoS               QoS
	NoLocal           bool
	RetainAsPublished bool
	RetainHandling    RetainHandling
}
