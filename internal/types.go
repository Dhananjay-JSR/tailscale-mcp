package internal

import "tailscale.com/tailcfg"

type TailscaleStatus struct {
	Version        string          `json:"Version"`
	TUN            bool            `json:"TUN"`
	BackendState   string          `json:"BackendState"`
	HaveNodeKey    bool            `json:"HaveNodeKey"`
	AuthURL        string          `json:"AuthURL"`
	TailscaleIPs   []string        `json:"TailscaleIPs"`
	Self           Node            `json:"Self"`
	Health         []string        `json:"Health"`
	MagicDNSSuffix string          `json:"MagicDNSSuffix"`
	CurrentTailnet Tailnet         `json:"CurrentTailnet"`
	CertDomains    interface{}     `json:"CertDomains"` // Can be nil
	Peer           map[string]Node `json:"Peer"`
	User           map[string]User `json:"User"`
	ClientVersion  ClientVersion   `json:"ClientVersion"`
}

type Node struct {
	ID             string                 `json:"ID"`
	PublicKey      string                 `json:"PublicKey"`
	HostName       string                 `json:"HostName"`
	DNSName        string                 `json:"DNSName"`
	OS             string                 `json:"OS"`
	UserID         int64                  `json:"UserID"`
	TailscaleIPs   []string               `json:"TailscaleIPs"`
	AllowedIPs     []string               `json:"AllowedIPs"`
	Addrs          []string               `json:"Addrs"`
	CurAddr        string                 `json:"CurAddr"`
	Relay          string                 `json:"Relay"`
	RxBytes        int64                  `json:"RxBytes"`
	TxBytes        int64                  `json:"TxBytes"`
	Created        string                 `json:"Created"`
	LastWrite      string                 `json:"LastWrite"`
	LastSeen       string                 `json:"LastSeen"`
	LastHandshake  string                 `json:"LastHandshake"`
	Online         bool                   `json:"Online"`
	ExitNode       bool                   `json:"ExitNode"`
	ExitNodeOption bool                   `json:"ExitNodeOption"`
	Active         bool                   `json:"Active"`
	PeerAPIURL     []string               `json:"PeerAPIURL"`
	Capabilities   []string               `json:"Capabilities,omitempty"`
	CapMap         map[string]interface{} `json:"CapMap,omitempty"`
	InNetworkMap   bool                   `json:"InNetworkMap"`
	InMagicSock    bool                   `json:"InMagicSock"`
	InEngine       bool                   `json:"InEngine"`
	KeyExpiry      string                 `json:"KeyExpiry"`
	SSHHostKeys    []string               `json:"sshHostKeys,omitempty"`
}

type Tailnet struct {
	Name            string `json:"Name"`
	MagicDNSSuffix  string `json:"MagicDNSSuffix"`
	MagicDNSEnabled bool   `json:"MagicDNSEnabled"`
}

type User struct {
	ID            int64    `json:"ID"`
	LoginName     string   `json:"LoginName"`
	DisplayName   string   `json:"DisplayName"`
	ProfilePicURL string   `json:"ProfilePicURL"`
	Roles         []string `json:"Roles"`
}

type ClientVersion struct {
	RunningLatest bool `json:"RunningLatest"`
}

type DeviceInfo struct {
	Addresses                 []string `json:"addresses"`
	Authorized                bool     `json:"authorized"`
	BlocksIncomingConnections bool     `json:"blocksIncomingConnections"`
	ClientVersion             string   `json:"clientVersion"`
	Created                   string   `json:"created"`
	Expires                   string   `json:"expires"`
	Hostname                  string   `json:"hostname"`
	ID                        string   `json:"id"`
	IsExternal                bool     `json:"isExternal"`
	KeyExpiryDisabled         bool     `json:"keyExpiryDisabled"`
	LastSeen                  string   `json:"lastSeen"`
	MachineKey                string   `json:"machineKey"`
	Name                      string   `json:"name"`
	NodeID                    string   `json:"nodeId"`
	NodeKey                   string   `json:"nodeKey"`
	OS                        string   `json:"os"`
	TailnetLockError          string   `json:"tailnetLockError"`
	TailnetLockKey            string   `json:"tailnetLockKey"`
	UpdateAvailable           bool     `json:"updateAvailable"`
	User                      string   `json:"user"`
}

type OauthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type TCPPortHandler struct {
	HTTPS bool `json:",omitempty"`
	HTTP  bool `json:",omitempty"`

	TCPForward   string `json:",omitempty"`
	TerminateTLS string `json:",omitempty"`
}

type HostPort string

type HTTPHandler struct {
	Path  string `json:",omitempty"`
	Proxy string `json:",omitempty"`

	Text string `json:",omitempty"`
}

type ServiceConfig struct {
	TCP map[uint16]*TCPPortHandler `json:",omitempty"`

	Web map[HostPort]*WebServerConfig `json:",omitempty"`
	Tun bool                          `json:",omitempty"`
}

type WebServerConfig struct {
	Handlers map[string]*HTTPHandler
}

type ServeConfig struct {
	TCP map[uint16]*TCPPortHandler `json:",omitempty"`

	Web map[HostPort]*WebServerConfig `json:",omitempty"`

	Services map[tailcfg.ServiceName]*ServiceConfig `json:",omitempty"`

	AllowFunnel map[HostPort]bool `json:",omitempty"`

	Foreground map[string]*ServeConfig `json:",omitempty"`

	ETag string `json:"-"`
}
