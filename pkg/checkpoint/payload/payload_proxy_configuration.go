package payload

type RequestPayloadProxyConfiguration struct {
	TenantId          string `json:"tenantId"`
	ServePathTemplate string `json:"servePathTemplate"`
	SignSecret        []byte `json:"signSecret"`
}
