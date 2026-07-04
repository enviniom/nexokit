package config

// IsLocal returns true if the environment is local or development.
func (c *Config) IsLocal() bool {
	return c.App.Env == "local" || c.App.Env == "development"
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsTest returns true if the environment is test.
func (c *Config) IsTest() bool {
	return c.App.Env == "test"
}

// ExposeDebugErrors returns true when the application environment allows
// internal error details to be exposed in API responses. It is the single
// source of truth for the debug field emitted by response.HandleError.
func (c *Config) ExposeDebugErrors() bool {
	return c.IsLocal() || c.IsTest()
}
