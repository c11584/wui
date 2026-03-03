package security

import (
	"net"
	"strings"
)

type IPWhitelist struct {
	allowedNets []*net.IPNet
	enabled     bool
}

func NewIPWhitelist() *IPWhitelist {
	return &IPWhitelist{
		allowedNets: make([]*net.IPNet, 0),
		enabled:     false,
	}
}

func (w *IPWhitelist) Enable(enabled bool) {
	w.enabled = enabled
}

func (w *IPWhitelist) IsEnabled() bool {
	return w.enabled
}

func (w *IPWhitelist) SetWhitelist(cidrs []string) error {
	nets := make([]*net.IPNet, 0)

	for _, cidr := range cidrs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr = cidr + "/128"
			} else {
				cidr = cidr + "/32"
			}
		}

		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		nets = append(nets, ipNet)
	}

	w.allowedNets = nets
	return nil
}

func (w *IPWhitelist) GetWhitelist() []string {
	result := make([]string, len(w.allowedNets))
	for i, n := range w.allowedNets {
		result[i] = n.String()
	}
	return result
}

func (w *IPWhitelist) IsAllowed(ipStr string) bool {
	if !w.enabled || len(w.allowedNets) == 0 {
		return true
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, ipNet := range w.allowedNets {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

var GlobalIPWhitelist = NewIPWhitelist()
