package stdin

import (
	"fmt"
	"strings"

	"github.com/juliengk/go-utils"
)

func (c *Config) ValidateKeys() error {
	missing := []string{}
	configKeys := []string{}

	keys := []string{
		"auth-password",
	}

	for k := range c.Keys {
		configKeys = append(configKeys, k)
	}

	for _, key := range keys {
		if !utils.StringInSlice(key, configKeys, false) {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		if len(missing) == 1 {
			return fmt.Errorf("The option key is missing: %s", missing[0])
		}
		return fmt.Errorf("The option keys are missing: %s", strings.Join(missing, ", "))
	}

	return nil
}
