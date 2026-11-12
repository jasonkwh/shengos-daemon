package config

type Config struct {
	PProfEnabled bool

	MQTTConn MQTTConnConfig
}

type MQTTConnConfig struct {
	Host      string
	Port      string
	KeepAlive uint16
	Username  *string
	Password  *string
}
