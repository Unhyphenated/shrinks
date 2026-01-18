package util

import (
	"fmt"
	"net"
	"strings"

	"github.com/mileusna/useragent"
)

type UserAgentInfo struct {
    Browser    string
    OS         string
    DeviceType string
}

func ParseUserAgent(uaString string) UserAgentInfo {
    ua := useragent.Parse(uaString)
    
    return UserAgentInfo{
        Browser:    fmt.Sprintf("%s %s", ua.Name, ua.Version),
        OS:         ua.OS,
        DeviceType: determineDevice(ua), // break logic into small private funcs
    }
}

func determineDevice(ua useragent.UserAgent) string {
    if ua.Mobile {
        return "Mobile"
    } else if ua.Tablet {
        return "Tablet"
    } else if ua.Desktop {
        return "Desktop"
    } else if ua.Bot {
        return "Bot"
    }
    return "Unknown"
}

func GetIP(ip string, remoteAddr string) string {
    // Check if we are behind a proxy (like Nginx, Docker, or Cloudflare)
    if ip == "" {
        // No proxy, get it directly
        // net.SplitHostPort removes the port number (e.g., "192.168.1.1:54321" -> "192.168.1.1")
        ip, _, _ = net.SplitHostPort(remoteAddr)
        return ip
    }
    
    // X-Forwarded-For can be a list (client, proxy1, proxy2). We want the first one.
    return strings.Split(ip, ",")[0]
}

func AnonymizeIP(ip string) string {
    parsedIP := net.ParseIP(ip)
    if parsedIP == nil {
        return "0.0.0.0"
    }

    if ipv4 := parsedIP.To4(); ipv4 != nil {
        ipv4[3] = 0
        return ipv4.String()
    }

    return ip
}