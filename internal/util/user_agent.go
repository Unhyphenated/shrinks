package util

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/mileusna/useragent"
)

func ParseUserAgent(uaString string) (browser, os, deviceType string) {
    ua := useragent.Parse(uaString)

    browser = fmt.Sprintf("%s %s", ua.Name, ua.Version)
    os = ua.OS

    if ua.Mobile {
        deviceType = "Mobile"
    } else if ua.Tablet {
        deviceType = "Tablet"
    } else if ua.Desktop {
        deviceType = "Desktop"
    } else if ua.Bot {
        deviceType = "Bot"
    } else {
        deviceType = "Unknown"
    }
    
    // Clean up strings
    browser = strings.TrimSpace(browser)
    os = strings.TrimSpace(os)
    deviceType = strings.TrimSpace(deviceType)

    return browser, os, deviceType
}

func GetIP(r *http.Request) string {
    // Check if we are behind a proxy (like Nginx, Docker, or Cloudflare)
    ip := r.Header.Get("X-Forwarded-For")
    if ip == "" {
        // No proxy, get it directly
        // net.SplitHostPort removes the port number (e.g., "192.168.1.1:54321" -> "192.168.1.1")
        ip, _, _ = net.SplitHostPort(r.RemoteAddr)
        return ip
    }
    
    // X-Forwarded-For can be a list (client, proxy1, proxy2). We want the first one.
    return strings.Split(ip, ",")[0]
}