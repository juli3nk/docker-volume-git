package driver

type Secreter interface {
	AddKey(key, value string) error
	ValidateKeys() error
	GetSecret() (string, error)
}
