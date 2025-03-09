package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Dhananjay-JSR/tailscale-mcp/internal"
	"github.com/Microsoft/go-winio"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"golang.org/x/sys/windows"
)

const tailscale_api_base_url = "https://api.tailscale.com/api/v2"

const default_tailnet_name = "-"

var oauth_token internal.OauthToken

func GetOauthToken(client_id string, client_secret string) string {
	if oauth_token.AccessToken == "" {
		requestBody := strings.NewReader("grant_type=client_credentials")
		req, err := http.NewRequest(http.MethodPost, tailscale_api_base_url+"/oauth/token", requestBody)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(client_id+":"+client_secret)))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		oauth_token.AccessToken = string(body)
		var token internal.OauthToken
		json.Unmarshal(body, &token)
		oauth_token = token

	}
	return oauth_token.AccessToken
}

func IsOauthTokenValid(token string) bool {
	req, err := http.NewRequest(http.MethodGet, tailscale_api_base_url+"/tailnet/"+default_tailnet_name+"/devices", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != 200 {
		return false
	}
	return true
}

// This function will always make sure to return a valid token ( LAME CODE but works!! )
func GetToken(client_id string, client_secret string) string {
	token := GetOauthToken(client_id, client_secret)
	if IsOauthTokenValid(token) {
		return token
	}
	oauth_token.AccessToken = ""
	return GetToken(client_id, client_secret)
}

type ListDevicesResponse struct {
	Devices []internal.DeviceInfo `json:"devices"`
}

func getConnector(ctx context.Context, network string, addr string) (net.Conn, error) {
	if runtime.GOOS != "windows" {
		log.Fatal("This is a windows only feature")
	}
	socketPath := "\\\\.\\pipe\\ProtectedPrefix\\Administrators\\Tailscale\\tailscaled"
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	return winio.DialPipeAccessImpLevel(ctx, socketPath, windows.GENERIC_READ|windows.GENERIC_WRITE, winio.PipeImpLevelIdentification)
}

func DaemonClient(ctx context.Context, method string, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	local_url := "http://local-tailscaled.sock/localapi/v0"
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: getConnector,
		},
	}
	req, err := http.NewRequest(method, local_url+path, body)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return client.Do(req)
}

// API CALLS
func ListAllTailnetDevices(token string) ListDevicesResponse {
	req, err := http.NewRequest(http.MethodGet, tailscale_api_base_url+"/tailnet/"+default_tailnet_name+"/devices", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var devices ListDevicesResponse
	json.Unmarshal(body, &devices)
	return devices
}

func BuildServerConfigToServer(ctx context.Context, dns_name string, port uint16, proxy_url string) internal.ServeConfig {
	server_config := internal.ServeConfig{
		TCP: map[uint16]*internal.TCPPortHandler{
			port: {
				HTTPS: true,
			},
		},
		Web: map[internal.HostPort]*internal.WebServerConfig{
			internal.HostPort(dns_name + ":" + strconv.Itoa(int(port))): {
				Handlers: map[string]*internal.HTTPHandler{
					"/": {
						Proxy: proxy_url,
					},
				},
			},
		},
		AllowFunnel: map[internal.HostPort]bool{
			internal.HostPort(dns_name + ":" + strconv.Itoa(int(port))): true,
		},
	}
	return server_config
}

func GetClientStatus(ctx context.Context) internal.TailscaleStatus {
	res, err := DaemonClient(ctx, http.MethodGet, "/status", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var status internal.TailscaleStatus
	json.Unmarshal(body, &status)
	return status
}

func SetServeConfig(ctx context.Context, server_config internal.ServeConfig) {
	json_server_config, err := json.Marshal(server_config)
	if err != nil {
		log.Fatal(err)
	}
	res, err := DaemonClient(ctx, http.MethodPost, "/serve-config", bytes.NewReader(json_server_config), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func main() {

	client_id := flag.String("client_id", "", "Client ID of the Tailscale Oauth Client")
	client_secret := flag.String("client_secret", "", "Client Secret of the Tailscale Oauth Client")
	flag.Parse()
	if *client_id == "" || *client_secret == "" {
		log.Fatal("Client ID and Secret are required")
	}
	s := server.NewMCPServer(
		"Tailscale MCP Demo",
		"0.0.1",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	listDevices := mcp.NewTool("list-devices", mcp.WithDescription("List All Devices on Your Tailnet"))

	host_local_address_to_public := mcp.NewTool("local-address-to-public-mapper", mcp.WithDescription("Expose your local address to the public"), mcp.WithBoolean("active", mcp.Required(), mcp.Description("Select Whether to Expose your local address to the public or not")), mcp.WithString("PORT", mcp.Description("Enter the local address to expose to the public Example Input 3000 , 4000 , 5000")))
	s.AddTool(listDevices, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		token := GetToken(*client_id, *client_secret)
		devices := ListAllTailnetDevices(token)
		return mcp.NewToolResultText(
			fmt.Sprintf("Devices: %v", devices.Devices),
		), nil
	})

	s.AddTool(host_local_address_to_public, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		active := request.Params.Arguments["active"].(bool)
		local_address := request.Params.Arguments["PORT"].(string)
		if active {
			dns_name := strings.TrimSuffix(GetClientStatus(ctx).Self.DNSName, ".")
			server_config := BuildServerConfigToServer(ctx, dns_name, 443, "http://localhost:"+local_address)
			SetServeConfig(ctx, server_config)
			return mcp.NewToolResultText(
				fmt.Sprintf("Server is now exposing %v to the public , You can access it at https://%v", local_address, dns_name),
			), nil
		} else {
			SetServeConfig(ctx, internal.ServeConfig{})
			return mcp.NewToolResultText(
				"Server is no longer exposing to the public",
			), nil
		}
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
