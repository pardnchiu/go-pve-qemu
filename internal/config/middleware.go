package config

import (
	"net"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getSubnet() string {
	defaultIP := "192.168.0.*"
	ary, err := net.InterfaceAddrs()
	if err != nil {
		return defaultIP
	}

	for _, e := range ary {
		ipNet, ok := e.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		if ipNet.IP.To4() == nil {
			continue
		}

		ip := ipNet.IP.String()
		part := strings.Split(ip, ".")
		if len(part) != 4 {
			continue
		}

		subnet := part[0] + "." + part[1] + "." + part[2] + ".*"

		if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
			return subnet
		} else if strings.HasPrefix(ip, "172.") {
			part1, err := strconv.Atoi(part[1])
			if err != nil || part1 < 16 || part1 > 31 {
				continue
			}
			return subnet
		}
	}

	return defaultIP
}

func isPrivate(ip string) bool {
	parseIP := net.ParseIP(ip)
	if parseIP == nil {
		return false
	}

	for _, e := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	} {
		_, ipNet, err := net.ParseCIDR(e)
		if err != nil || !ipNet.Contains(parseIP) {
			continue
		}
		return true
	}

	return false
}

func isPassed(origin string) bool {
	if origin == "" {
		return false
	}

	if strings.HasPrefix(origin, "http://") {
		origin = strings.TrimPrefix(origin, "http://")
	} else if strings.HasPrefix(origin, "https://") {
		origin = strings.TrimPrefix(origin, "https://")
	}

	if port := strings.Index(origin, ":"); port != -1 {
		origin = origin[:port]
	}

	// Only allow private IPs
	if !isPrivate(origin) {
		return false
	}

	subnet := getSubnet()
	prefix := strings.TrimSuffix(subnet, "*")

	return strings.HasPrefix(origin, prefix)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if isPassed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if origin == "" {
			c.Header("Access-Control-Allow-Origin", getSubnet())
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
