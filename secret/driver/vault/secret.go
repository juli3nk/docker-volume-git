package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

func (c *Config) GetSecret() (string, error) {
	config := api.DefaultConfig()

	config.Address = c.Keys["vault-addr"]

	// Build the client
	client, err := api.NewClient(config)
	if err != nil {
		return "", fmt.Errorf("Error initializing client: %s", err)
	}

	// Set the token
	client.SetToken(c.Keys["vault-token"])

	// Read the secret
	path := c.Keys["vault-secret-path"]
	secret, err := client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("Error reading %s: %s", path, err)
	}
	if secret == nil {
		return "", fmt.Errorf("No value found at %s", path)
	}

	v, ok := secret.Data[c.Keys["vault-secret-field"]]
	if !ok {
		return "", fmt.Errorf("No value")
	}

	return v.(string), nil
}
