package payload

type RequestPayloadProxyConfiguration struct {
	TenantId          string `mapstructure:"tenant-id"`
	ServePathTemplate string `mapstructure:"serve-path-template"`
	SignSecret        string `mapstructure:"sign-secret"`
}
