package vault

import (
	"fmt"
)

func (c *Config) AddKey(key, value string) error {
	if _, ok := c.Keys[key]; ok {
		return fmt.Errorf("Key '%s' already exists", key)
	}

	c.Keys[key] = value

	return nil
}
