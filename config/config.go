package config

type Config struct {
	LogLevel     int8
	PProfEnabled bool

	Clipboard ClipboardConfig
	MQTTConn  MQTTConnConfig
}

type ClipboardConfig struct {
	SSHUser string
}

type MQTTConnConfig struct {
	Host      string
	Port      string
	KeepAlive uint16
	Username  *string
	Password  *string
}
