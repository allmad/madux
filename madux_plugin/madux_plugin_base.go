package madux_plugin

type BasePlugin struct {
	*Plugin
	delegate ServerDelegater
}
