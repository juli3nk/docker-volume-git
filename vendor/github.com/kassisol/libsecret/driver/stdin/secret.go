package stdin

func (c *Config) GetSecret() (string, error) {
	return c.Keys["auth-password"], nil
}
