package account

// Plugin is a component which extends the account's functionality.
type Plugin interface {
	// Starts the given plugin and passes in the account object.
	Start(acc Account) error
	// Shutdown instructs the plugin to terminate.
	Shutdown() error
	// Name returns the name of the plugin.
	Name() string
}
