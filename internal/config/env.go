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
