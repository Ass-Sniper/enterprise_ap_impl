package auth

var plugins []Plugin

func RegisterPlugin(p Plugin) {
	plugins = append(plugins, p)
}

func AllPlugins() []Plugin {
	return plugins
}
