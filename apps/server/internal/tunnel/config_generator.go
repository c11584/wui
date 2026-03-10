package tunnel

import (
	"encoding/json"
	"fmt"

	"github.com/your-org/wui/internal/models"
)

// ConfigGenerator generates Xray configuration from tunnel model
type ConfigGenerator struct{}

func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// XrayConfig represents Xray configuration structure
type XrayConfig struct {
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
	Routing   Routing    `json:"routing"`
	Policy    Policy     `json:"policy"`
}

type Inbound struct {
	Tag      string                 `json:"tag"`
	Protocol string                 `json:"protocol"`
	Listen   string                 `json:"listen"`
	Port     int                    `json:"port"`
	Settings map[string]interface{} `json:"settings"`
	Sniffing *Sniffing              `json:"sniffing,omitempty"`
}

type Sniffing struct {
	Enabled      bool     `json:"enabled"`
	DestOverride []string `json:"destOverride"`
}

type Outbound struct {
	Tag            string                 `json:"tag"`
	Protocol       string                 `json:"protocol"`
	Settings       map[string]interface{} `json:"settings"`
	StreamSettings map[string]interface{} `json:"streamSettings,omitempty"`
}

type Routing struct {
	DomainStrategy string        `json:"domainStrategy"`
	Rules          []RoutingRule `json:"rules"`
}

type RoutingRule struct {
	Type        string   `json:"type"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
}

type Policy struct {
	Levels map[string]PolicyLevel `json:"levels"`
	System PolicySystem           `json:"system"`
}

type PolicyLevel struct {
	StatsUserUplink   bool `json:"statsUserUplink"`
	StatsUserDownlink bool `json:"statsUserDownlink"`
}

type PolicySystem struct {
	StatsInboundUplink    bool `json:"statsInboundUplink"`
	StatsInboundDownlink  bool `json:"statsInboundDownlink"`
	StatsOutboundUplink   bool `json:"statsOutboundUplink"`
	StatsOutboundDownlink bool `json:"statsOutboundDownlink"`
}

// Generate generates Xray configuration from tunnel and outbounds
func (g *ConfigGenerator) Generate(tunnel *models.Tunnel) (*XrayConfig, error) {
	config := &XrayConfig{
		Inbounds:  []Inbound{},
		Outbounds: []Outbound{},
		Routing: Routing{
			DomainStrategy: "IPIfNonMatch",
			Rules:          []RoutingRule{},
		},
		Policy: Policy{
			Levels: map[string]PolicyLevel{
				"0": {
					StatsUserUplink:   true,
					StatsUserDownlink: true,
				},
			},
			System: PolicySystem{
				StatsInboundUplink:    true,
				StatsInboundDownlink:  true,
				StatsOutboundUplink:   true,
				StatsOutboundDownlink: true,
			},
		},
	}

	// Generate inbound
	inbound, err := g.generateInbound(tunnel)
	if err != nil {
		return nil, err
	}
	config.Inbounds = append(config.Inbounds, *inbound)

	// Generate outbounds
	outboundTags := []string{}
	for _, out := range tunnel.Outbounds {
		outbound, err := g.generateOutbound(&out)
		if err != nil {
			return nil, err
		}
		config.Outbounds = append(config.Outbounds, *outbound)
		outboundTags = append(outboundTags, outbound.Tag)
	}

	defaultOutbound := "direct"
	if len(outboundTags) > 0 {
		defaultOutbound = outboundTags[0]
	}

	config.Outbounds = append(config.Outbounds, Outbound{
		Tag:      "block",
		Protocol: "blackhole",
		Settings: map[string]interface{}{},
	})

	config.Outbounds = append(config.Outbounds, Outbound{
		Tag:      "direct",
		Protocol: "freedom",
		Settings: map[string]interface{}{},
	})

	if tunnel.ACLEnabled {
		aclRules := g.generateACLRules(tunnel, inbound.Tag, defaultOutbound)
		config.Routing.Rules = append(config.Routing.Rules, aclRules...)
	}

	config.Routing.Rules = append(config.Routing.Rules, RoutingRule{
		Type:        "field",
		InboundTag:  []string{inbound.Tag},
		OutboundTag: defaultOutbound,
	})

	return config, nil
}

func (g *ConfigGenerator) generateInbound(tunnel *models.Tunnel) (*Inbound, error) {
	inbound := &Inbound{
		Tag:      fmt.Sprintf("inbound_%d", tunnel.ID),
		Protocol: tunnel.InboundProtocol,
		Listen:   tunnel.InboundListen,
		Port:     tunnel.InboundPort,
		Settings: map[string]interface{}{},
		Sniffing: &Sniffing{
			Enabled:      true,
			DestOverride: []string{"http", "tls"},
		},
	}

	// Protocol-specific settings
	switch tunnel.InboundProtocol {
	case "socks5", "socks":
		inbound.Protocol = "socks"
		inbound.Settings["udp"] = tunnel.UDPEnabled
		inbound.Settings["auth"] = "noauth"
		if tunnel.InboundAuth {
			inbound.Settings["auth"] = "password"
		}

	case "http":
		inbound.Settings["allowTransparent"] = false

	default:
		return nil, fmt.Errorf("unsupported inbound protocol: %s", tunnel.InboundProtocol)
	}

	return inbound, nil
}

// generateOutbound generates outbound configuration
func (g *ConfigGenerator) generateOutbound(outbound *models.Outbound) (*Outbound, error) {
	out := &Outbound{
		Tag:      fmt.Sprintf("outbound_%d", outbound.ID),
		Protocol: outbound.Protocol,
		Settings: map[string]interface{}{},
	}

	// Parse protocol-specific config
	var protocolConfig map[string]interface{}
	if outbound.Config != "" {
		if err := json.Unmarshal([]byte(outbound.Config), &protocolConfig); err != nil {
			return nil, fmt.Errorf("failed to parse outbound config: %v", err)
		}
	}

	// Protocol-specific settings
	switch outbound.Protocol {
	case "vless":
		out.Settings["vnext"] = []map[string]interface{}{
			{
				"address": outbound.Address,
				"port":    outbound.Port,
				"users": []map[string]interface{}{
					{
						"id":         protocolConfig["uuid"],
						"encryption": protocolConfig["encryption"],
						"flow":       protocolConfig["flow"],
					},
				},
			},
		}
		// Add stream settings if present
		if streamSettings, ok := protocolConfig["streamSettings"].(map[string]interface{}); ok {
			out.StreamSettings = streamSettings
		}

	case "vmess":
		out.Settings["vnext"] = []map[string]interface{}{
			{
				"address": outbound.Address,
				"port":    outbound.Port,
				"users": []map[string]interface{}{
					{
						"id":       protocolConfig["uuid"],
						"alterId":  protocolConfig["alterId"],
						"security": protocolConfig["security"],
					},
				},
			},
		}

	case "trojan":
		out.Settings["servers"] = []map[string]interface{}{
			{
				"address":  outbound.Address,
				"port":     outbound.Port,
				"password": protocolConfig["password"],
			},
		}

	case "shadowsocks":
		out.Settings["servers"] = []map[string]interface{}{
			{
				"address":  outbound.Address,
				"port":     outbound.Port,
				"method":   protocolConfig["method"],
				"password": protocolConfig["password"],
			},
		}

	case "socks5", "socks":
		out.Protocol = "socks"
		out.Settings["servers"] = []map[string]interface{}{
			{
				"address": outbound.Address,
				"port":    outbound.Port,
			},
		}

	case "freedom", "direct":
		out.Protocol = "freedom"
		out.Settings = map[string]interface{}{}

	default:
		return nil, fmt.Errorf("unsupported outbound protocol: %s", outbound.Protocol)
	}

	return out, nil
}

// ToJSON converts config to JSON string
func (g *ConfigGenerator) ToJSON(config *XrayConfig) (string, error) {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (g *ConfigGenerator) generateACLRules(tunnel *models.Tunnel, inboundTag, defaultOutbound string) []RoutingRule {
	var rules []RoutingRule

	denyDomains := parseJSONArray(tunnel.DenyDomains)
	denyIPs := parseJSONArray(tunnel.DenyIPs)
	allowDomains := parseJSONArray(tunnel.AllowDomains)
	allowIPs := parseJSONArray(tunnel.AllowIPs)

	if tunnel.ACLMode == "blacklist" {
		if len(denyDomains) > 0 {
			rules = append(rules, RoutingRule{
				Type:        "field",
				InboundTag:  []string{inboundTag},
				Domain:      denyDomains,
				OutboundTag: "block",
			})
		}
		if len(denyIPs) > 0 {
			rules = append(rules, RoutingRule{
				Type:        "field",
				InboundTag:  []string{inboundTag},
				IP:          denyIPs,
				OutboundTag: "block",
			})
		}
	} else {
		if len(allowDomains) > 0 {
			rules = append(rules, RoutingRule{
				Type:        "field",
				InboundTag:  []string{inboundTag},
				Domain:      allowDomains,
				OutboundTag: defaultOutbound,
			})
		}
		if len(allowIPs) > 0 {
			rules = append(rules, RoutingRule{
				Type:        "field",
				InboundTag:  []string{inboundTag},
				IP:          allowIPs,
				OutboundTag: defaultOutbound,
			})
		}
		rules = append(rules, RoutingRule{
			Type:        "field",
			InboundTag:  []string{inboundTag},
			OutboundTag: "block",
		})
	}

	return rules
}

func parseJSONArray(s string) []string {
	if s == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil
	}
	return arr
}
