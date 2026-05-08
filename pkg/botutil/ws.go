package botutil

import (
	"fmt"
	"strings"

	"github.com/Mininglamp-OSS/octo-lib/config"
)

// DeriveWSURL derives the WuKongIM WebSocket URL from server config.
func DeriveWSURL(cfg *config.Config) string {
	baseURL := strings.TrimSpace(cfg.External.BaseURL)
	if baseURL != "" {
		host := baseURL
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimPrefix(host, "http://")
		if idx := strings.Index(host, "/"); idx >= 0 {
			host = host[:idx]
		}
		// Host contains port → direct mode (e.g. 192.168.x.x:8090), use WuKongIM 5200
		if strings.Contains(host, ":") {
			if idx := strings.LastIndex(host, ":"); idx >= 0 {
				host = host[:idx]
			}
			return fmt.Sprintf("ws://%s:5200", host)
		}
		// Domain mode → Nginx reverse proxy
		if strings.HasPrefix(baseURL, "https://") {
			return fmt.Sprintf("wss://%s/ws", host)
		}
		return fmt.Sprintf("ws://%s/ws", host)
	}
	// Fallback: derive from WuKongIM API URL
	apiURL := cfg.WuKongIM.APIURL
	host := apiURL
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if idx := strings.Index(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	// External.IP overrides derived host (direct-access deployments)
	if strings.TrimSpace(cfg.External.IP) != "" {
		host = cfg.External.IP
	}
	return fmt.Sprintf("ws://%s:5200", host)
}
