package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const tailscale_api_base_url = "https://api.tailscale.com/api/v2"

const default_tailnet_name = "-"

var oauth_token OauthToken

type OauthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

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
		var token OauthToken
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

// This function will always make sure to return a valid token
func GetToken(client_id string, client_secret string) string {
	token := GetOauthToken(client_id, client_secret)
	if IsOauthTokenValid(token) {
		return token
	}
	oauth_token.AccessToken = ""
	return GetToken(client_id, client_secret)
}

func main() {
	client_id := flag.String("client_id", "", "Client ID of the Tailscale Oauth Client")
	client_secret := flag.String("client_secret", "", "Client Secret of the Tailscale Oauth Client")
	flag.Parse()
	if *client_id == "" || *client_secret == "" {
		log.Fatal("Client ID and Secret are required")
	}
	token := GetToken(*client_id, *client_secret)
	fmt.Println(token)
}
