# Tailscale MCP Server

An experimental Tailscale MCP server created by reverse engineering the Tailscale client.

---

### Windows Only Support

Currently, this server only supports Windows platforms.

### Running the Server

1. Build the Go binary
2. Add the following configuration to your MCP config:

```json
{
    "mcpServers": {
        "tailscale-mcp": {
            "command": "$absolute-path-of-tailscale-mcp.exe",
            "args": ["-client_id=$CLIENT_ID","-client_secret=$CLIENT_SECRET"]
        }
    }
}

```