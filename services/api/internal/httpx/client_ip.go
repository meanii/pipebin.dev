package httpx

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")

	if ip != "" {
		parts := strings.Split(ip, ",")
		ip = strings.TrimSpace(parts[0])
	} else {
		ip = r.Header.Get("X-Real-IP")
	}

	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}

	return ""
}
