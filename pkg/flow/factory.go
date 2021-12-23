package flow

type Config struct {
	FlowEndPoint    string `required:"true"`
	MockFlowEnabled bool
}

type InMemSignerConfig struct {
	PrivateKeyHex          string
	PrivateKeyHashAlgoName string
	SignatureAlgorithm     string
}

type KMSConfig struct {
	KMSProjectID  string
	KMSLocationID string
	KMSKeyRingID  string
	KMSKeyID      string
	KMSKeyVersion string
}

func New(c *Config) (Provider, error) {
	if c.MockFlowEnabled {
		return NewMockProvider(), nil
	}
	return newFlowProvider(c)
}

func NewMockProvider() Provider {
	// Set default behaviour to allow traffic
	return &mockImpl{
		client: &MockClient{},
	}
}
