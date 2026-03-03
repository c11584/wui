package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-org/wui/internal/models"
)

func (s *Server) handleGetSubscription(c *gin.Context) {
	userID := c.GetUint("userId")
	format := c.Query("format")
	if format == "" {
		format = "clash"
	}

	var tunnels []models.Tunnel
	s.db.Preload("Outbounds").Where("user_id = ? AND enabled = ?", userID, true).Find(&tunnels)

	switch format {
	case "clash":
		s.generateClashConfig(c, tunnels)
	case "v2ray", "vmess":
		s.generateV2rayLinks(c, tunnels)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse("Unsupported format"))
	}
}

func (s *Server) generateClashConfig(c *gin.Context, tunnels []models.Tunnel) {
	config := map[string]interface{}{
		"port":                7890,
		"socks-port":          7891,
		"allow-lan":           false,
		"mode":                "rule",
		"log-level":           "info",
		"external-controller": "127.0.0.1:9090",
	}

	var proxies []map[string]interface{}
	var proxyNames []string

	for _, tunnel := range tunnels {
		for _, outbound := range tunnel.Outbounds {
			proxy := s.outboundToClashProxy(tunnel, outbound)
			if proxy != nil {
				proxies = append(proxies, proxy)
				proxyNames = append(proxyNames, proxy["name"].(string))
			}
		}
	}

	config["proxies"] = proxies
	config["proxy-groups"] = []map[string]interface{}{
		{
			"name":    "PROXY",
			"type":    "select",
			"proxies": proxyNames,
		},
	}
	config["rules"] = []string{"MATCH,PROXY"}

	yaml := s.toYAML(config)
	c.Header("Content-Disposition", "attachment; filename=wui.yaml")
	c.Data(http.StatusOK, "text/yaml", []byte(yaml))
}

func (s *Server) outboundToClashProxy(tunnel models.Tunnel, outbound models.Outbound) map[string]interface{} {
	proxy := map[string]interface{}{
		"name":   fmt.Sprintf("%s-%s", tunnel.Name, outbound.Name),
		"server": outbound.Address,
		"port":   outbound.Port,
	}

	var cfg map[string]interface{}
	if outbound.Config != "" {
		json.Unmarshal([]byte(outbound.Config), &cfg)
	}

	switch outbound.Protocol {
	case "vmess":
		proxy["type"] = "vmess"
		proxy["uuid"] = cfg["uuid"]
		proxy["alterId"] = cfg["alterId"]
		proxy["cipher"] = cfg["security"]
		if ws, ok := cfg["streamSettings"].(map[string]interface{}); ok {
			if ws["network"] == "ws" {
				proxy["network"] = "ws"
				if wsOpts, ok := ws["wsSettings"].(map[string]interface{}); ok {
					proxy["ws-path"] = wsOpts["path"]
				}
			}
			if tls, ok := ws["security"].(string); ok && tls == "tls" {
				proxy["tls"] = true
			}
		}
	case "vless":
		proxy["type"] = "vless"
		proxy["uuid"] = cfg["uuid"]
		proxy["flow"] = cfg["flow"]
	case "trojan":
		proxy["type"] = "trojan"
		proxy["password"] = cfg["password"]
	case "shadowsocks":
		proxy["type"] = "ss"
		proxy["cipher"] = cfg["method"]
		proxy["password"] = cfg["password"]
	case "socks5":
		proxy["type"] = "socks5"
	case "http":
		proxy["type"] = "http"
	default:
		return nil
	}

	return proxy
}

func (s *Server) generateV2rayLinks(c *gin.Context, tunnels []models.Tunnel) {
	var links []string

	for _, tunnel := range tunnels {
		for _, outbound := range tunnel.Outbounds {
			link := s.outboundToV2rayLink(tunnel, outbound)
			if link != "" {
				links = append(links, link)
			}
		}
	}

	content := base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
	c.Header("Content-Disposition", "attachment; filename=wui.txt")
	c.Data(http.StatusOK, "text/plain", []byte(content))
}

func (s *Server) outboundToV2rayLink(tunnel models.Tunnel, outbound models.Outbound) string {
	var cfg map[string]interface{}
	if outbound.Config != "" {
		json.Unmarshal([]byte(outbound.Config), &cfg)
	}

	switch outbound.Protocol {
	case "vmess":
		vmess := map[string]interface{}{
			"v":    "2",
			"ps":   fmt.Sprintf("%s-%s", tunnel.Name, outbound.Name),
			"add":  outbound.Address,
			"port": outbound.Port,
			"id":   cfg["uuid"],
			"aid":  cfg["alterId"],
			"scy":  cfg["security"],
			"net":  "tcp",
			"type": "none",
		}
		if ws, ok := cfg["streamSettings"].(map[string]interface{}); ok {
			if ws["network"] == "ws" {
				vmess["net"] = "ws"
				if wsOpts, ok := ws["wsSettings"].(map[string]interface{}); ok {
					vmess["path"] = wsOpts["path"]
				}
			}
			if ws["security"] == "tls" {
				vmess["tls"] = "tls"
			}
		}
		data, _ := json.Marshal(vmess)
		return "vmess://" + base64.StdEncoding.EncodeToString(data)

	case "vless":
		link := fmt.Sprintf("vless://%s@%s:%d?encryption=none",
			cfg["uuid"], outbound.Address, outbound.Port)
		if flow, ok := cfg["flow"].(string); ok && flow != "" {
			link += "&flow=" + flow
		}
		link += "#" + fmt.Sprintf("%s-%s", tunnel.Name, outbound.Name)
		return link

	case "trojan":
		return fmt.Sprintf("trojan://%s@%s:%d#%s-%s",
			cfg["password"], outbound.Address, outbound.Port, tunnel.Name, outbound.Name)

	case "shadowsocks":
		return fmt.Sprintf("ss://%s@%s:%d#%s-%s",
			base64.StdEncoding.EncodeToString([]byte(cfg["method"].(string)+":"+cfg["password"].(string))),
			outbound.Address, outbound.Port, tunnel.Name, outbound.Name)
	}

	return ""
}

func (s *Server) toYAML(data map[string]interface{}) string {
	var yaml strings.Builder

	for key, value := range data {
		switch v := value.(type) {
		case string:
			yaml.WriteString(fmt.Sprintf("%s: %s\n", key, v))
		case int:
			yaml.WriteString(fmt.Sprintf("%s: %d\n", key, v))
		case bool:
			yaml.WriteString(fmt.Sprintf("%s: %t\n", key, v))
		case []string:
			yaml.WriteString(fmt.Sprintf("%s:\n", key))
			for _, item := range v {
				yaml.WriteString(fmt.Sprintf("  - %s\n", item))
			}
		case []map[string]interface{}:
			yaml.WriteString(fmt.Sprintf("%s:\n", key))
			for _, item := range v {
				yaml.WriteString("  - ")
				first := true
				for k, val := range item {
					if !first {
						yaml.WriteString("    ")
					}
					switch val := val.(type) {
					case string:
						yaml.WriteString(fmt.Sprintf("%s: %s\n", k, val))
					case int:
						yaml.WriteString(fmt.Sprintf("%s: %d\n", k, val))
					case bool:
						yaml.WriteString(fmt.Sprintf("%s: %t\n", k, val))
					default:
						yaml.WriteString(fmt.Sprintf("%s: %v\n", k, val))
					}
					first = false
				}
			}
		}
	}

	return yaml.String()
}

func (s *Server) handleGetSubscriptionToken(c *gin.Context) {
	userID := c.GetUint("userId")
	token := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(userID)) + ":" + "wui"))

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"token":    token,
		"clashUrl": "/api/subscription?format=clash&token=" + token,
		"v2rayUrl": "/api/subscription?format=v2ray&token=" + token,
	}))
}
