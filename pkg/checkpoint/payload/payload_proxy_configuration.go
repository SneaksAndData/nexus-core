package payload

import "strings"

type RequestPayloadProxyConfiguration struct {
	TenantId          string `mapstructure:"tenant-id"`
	ServePathTemplate string `mapstructure:"serve-path-template"`
	SignSecret        string `mapstructure:"sign-secret"`
	ExternalName      string `mapstructure:"external-name"`
	Insecure          bool   `mapstructure:"insecure"`
}

func (rpp *RequestPayloadProxyConfiguration) GetProxyScheme() string {
	if rpp.Insecure {
		return "http"
	}
	return "https"
}

func (rpp *RequestPayloadProxyConfiguration) GetServePathTemplate() string {
	if rpp.ServePathTemplate == "" || strings.HasPrefix(rpp.ServePathTemplate, "/") {
		return rpp.ServePathTemplate
	}
	return "/" + rpp.ServePathTemplate
}
