package config

import (
	"strconv"
	"strings"
	"testing"
)

func TestMCPConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     MCPConfig
		wantErr string // substring of the error; empty means valid
	}{
		{"https base url", MCPConfig{ResourceID: "https://mcp.example.com/mcp", PublicBaseURL: "https://mcp.example.com"}, ""},
		{"http localhost", MCPConfig{ResourceID: "http://localhost:9090/mcp", PublicBaseURL: "http://localhost:9090"}, ""},
		{"explicit port, no userinfo", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com:8443"}, ""},
		{"ipv6 host with port", MCPConfig{ResourceID: "r", PublicBaseURL: "https://[::1]:9090"}, ""},
		{"path prefix behind a proxy", MCPConfig{ResourceID: "r", PublicBaseURL: "https://example.com/mcp-service"}, ""},
		{"percent-encoded character in path prefix", MCPConfig{ResourceID: "r", PublicBaseURL: "https://example.com/a%3Fb"}, ""},
		{"percent-encoded quote in path prefix", MCPConfig{ResourceID: "r", PublicBaseURL: "https://example.com/team%22west"}, ""},
		{"punctuation allowed in a path", MCPConfig{ResourceID: "r", PublicBaseURL: "https://example.com/a,b;c:d@e=f+g"}, ""},

		{"both missing", MCPConfig{}, "MCP_RESOURCE_ID, MCP_PUBLIC_BASE_URL"},
		{"public base url missing", MCPConfig{ResourceID: "r"}, "MCP_PUBLIC_BASE_URL"},
		{"resource id missing", MCPConfig{PublicBaseURL: "https://mcp.example.com"}, "MCP_RESOURCE_ID"},

		{"no scheme", MCPConfig{ResourceID: "r", PublicBaseURL: "mcp.example.com"}, "absolute http(s) URL"},
		{"host and port without scheme", MCPConfig{ResourceID: "r", PublicBaseURL: "localhost:9090"}, "absolute http(s) URL"},
		{"unsupported scheme", MCPConfig{ResourceID: "r", PublicBaseURL: "ftp://mcp.example.com"}, "absolute http(s) URL"},
		{"no host", MCPConfig{ResourceID: "r", PublicBaseURL: "https://"}, "absolute http(s) URL"},
		{"port only, no hostname", MCPConfig{ResourceID: "r", PublicBaseURL: "https://:9090"}, "absolute http(s) URL"},

		{"username and password", MCPConfig{ResourceID: "r", PublicBaseURL: "https://admin:s3cret@mcp.example.com"}, "username or password"},
		{"username only", MCPConfig{ResourceID: "r", PublicBaseURL: "https://admin@mcp.example.com"}, "username or password"},
		{"empty userinfo", MCPConfig{ResourceID: "r", PublicBaseURL: "https://@mcp.example.com"}, "username or password"},
		{"credentials with a port", MCPConfig{ResourceID: "r", PublicBaseURL: "https://admin:s3cret@mcp.example.com:8443"}, "username or password"},
		{"credentials and a trailing slash", MCPConfig{ResourceID: "r", PublicBaseURL: "https://admin:s3cret@mcp.example.com/"}, "username or password"},
		{"credentials and a bad scheme", MCPConfig{ResourceID: "r", PublicBaseURL: "ftp://admin:s3cret@mcp.example.com"}, "absolute http(s) URL"},

		{"trailing slash", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com/"}, "must not end with a slash"},
		{"query string", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com?x=1"}, "query or fragment"},
		{"fragment", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com#top"}, "query or fragment"},
		{"lone question mark", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com?"}, "query or fragment"},
		{"lone hash", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com#"}, "query or fragment"},
		{"lone question mark after path prefix", MCPConfig{ResourceID: "r", PublicBaseURL: "https://example.com/mcp-service?"}, "query or fragment"},

		{"literal quote in path", MCPConfig{ResourceID: "r", PublicBaseURL: `https://mcp.example.com/team"west`}, "canonical form"},
		{"space in path", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com/team west"}, "canonical form"},
		{"backslash in path", MCPConfig{ResourceID: "r", PublicBaseURL: `https://mcp.example.com/team\west`}, "canonical form"},
		{"non-ASCII letter in path", MCPConfig{ResourceID: "r", PublicBaseURL: "https://mcp.example.com/équipe"}, "canonical form"},
		{"uppercase scheme", MCPConfig{ResourceID: "r", PublicBaseURL: "HTTPS://mcp.example.com"}, "canonical form"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
			// The value can carry credentials, and a startup error is logged, so
			// no error may repeat it, or the password inside it. The check is for
			// the quoted form the messages used to print: the example URL in the
			// text can legitimately contain a short value such as "https://".
			if tt.cfg.PublicBaseURL != "" && strings.Contains(err.Error(), strconv.Quote(tt.cfg.PublicBaseURL)) {
				t.Errorf("error %q repeats the configured value", err)
			}
			if strings.Contains(err.Error(), "s3cret") {
				t.Errorf("error %q leaks the password", err)
			}
		})
	}
}
