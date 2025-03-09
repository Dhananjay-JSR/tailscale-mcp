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

func main() {
	client_id := flag.String("client_id", "", "Client ID of the Tailscale Oauth Client")
	client_secret := flag.String("client_secret", "", "Client Secret of the Tailscale Oauth Client")
	flag.Parse()
	if *client_id == "" || *client_secret == "" {
		log.Fatal("Client ID and Secret are required")
	}
	data := GetOauthToken(*client_id, *client_secret)
	fmt.Println(data)
}
