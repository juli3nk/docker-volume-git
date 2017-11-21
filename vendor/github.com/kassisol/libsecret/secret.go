package libsecret

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kassisol/libsecret/driver"
)

type Initialize func() (driver.Secreter, error)

var initializers = make(map[string]Initialize)

func supportedDrivers() string {
	drivers := make([]string, 0, len(initializers))

	for v := range initializers {
		drivers = append(drivers, string(v))
	}

	sort.Strings(drivers)

	return strings.Join(drivers, ",")
}

func NewDriver(driver string) (driver.Secreter, error) {
	if init, exists := initializers[driver]; exists {
		return init()
	}

	return nil, fmt.Errorf("The Secret Driver: %s is not supported. Supported drivers are %s", driver, supportedDrivers())
}

func RegisterDriver(driver string, init Initialize) {
	initializers[driver] = init
}
