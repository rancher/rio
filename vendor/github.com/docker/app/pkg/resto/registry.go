package resto

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"github.com/docker/docker/api/types"
	dd "github.com/docker/docker/distribution"
	"github.com/docker/docker/registry"
	digest "github.com/opencontainers/go-digest"
)

type myreference string

func (m myreference) String() string {
	return string(m)
}

func (m myreference) Name() string {
	return string(m)
}

// MediaTypeConfig is the media type used for configuration files.
const MediaTypeConfig = "application/vndr.docker.config"

// ConfigManifest is a Manifest type holding arbitrary data.
type ConfigManifest struct {
	mediaType string
	payload   []byte
}

// References returns the objects this Manifest refers to.
func (c *ConfigManifest) References() []distribution.Descriptor {
	return nil
}

// Payload returns the mediatype and payload of this manifest.
func (c *ConfigManifest) Payload() (string, []byte, error) {
	return c.mediaType, c.payload, nil
}

// NewConfigManifest creates and returns an new ConfigManifest.
func NewConfigManifest(mediaType string, payload []byte) *ConfigManifest {
	return &ConfigManifest{mediaType, payload}
}

func init() {
	distribution.RegisterManifestSchema(MediaTypeConfig, func(b []byte) (distribution.Manifest, distribution.Descriptor, error) {
		return &ConfigManifest{
				mediaType: MediaTypeConfig,
				payload:   b,
			},
			distribution.Descriptor{
				MediaType: MediaTypeConfig,
				Size:      int64(len(b)),
				Digest:    digest.SHA256.FromBytes(b),
			}, nil
	})
}

// NewRepository instantiates a distribution.Repository pointing to the given target, with credentials
func NewRepository(ctx context.Context, endpoint string, repository string, opts RegistryOptions) (distribution.Repository, error) {
	named := myreference(repository)
	authConfig := &types.AuthConfig{
		Username: opts.Username,
		Password: opts.Password,
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	apiendpoint := registry.APIEndpoint{
		Mirror:    false,
		URL:       url,
		Version:   2,
		TLSConfig: &tls.Config{InsecureSkipVerify: opts.Insecure},
	}
	repoInfo, err := registry.ParseRepositoryInfo(named)
	if err != nil {
		return nil, err
	}
	repo, _, err := dd.NewV2Repository(ctx, repoInfo, apiendpoint, nil, authConfig, "push", "pull")
	if err == nil {
		return repo, nil
	}
	if !strings.Contains(err.Error(), "HTTP response to HTTPS client") {
		return nil, err
	}
	if !opts.CleartextCredentials {
		// Don't use credentials over insecure connection unless instnucted to
		authConfig.Username = ""
		authConfig.Password = ""
	}
	endpointHTTP := strings.Replace(endpoint, "https://", "http://", 1)
	urlHTTP, err := url.Parse(endpointHTTP)
	if err != nil {
		return nil, err
	}
	apiendpoint.URL = urlHTTP
	repo, _, err = dd.NewV2Repository(ctx, repoInfo, apiendpoint, nil, authConfig, "push", "pull")
	return repo, err
}

// NewTransportCatalog returns a transport suitable for a catalog operation
func NewTransportCatalog(endpoint string, opts RegistryOptions) (http.RoundTripper, error) {
	// taken from docker/distribution/registry.go
	direct := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	base := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                direct.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: opts.Insecure},
		DisableKeepAlives:   true,
	}

	authTransport := transport.NewTransport(base)
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	challengeManager, _, err := registry.PingV2Registry(endpointURL, authTransport)
	if err != nil {
		return nil, err
	}
	scope := auth.RegistryScope{
		Name:    "catalog",
		Actions: []string{"*"},
	}
	authConfig := &types.AuthConfig{
		Username: opts.Username,
		Password: opts.Password,
	}
	creds := registry.NewStaticCredentialStore(authConfig)
	tokenHandlerOptions := auth.TokenHandlerOptions{
		Transport:   authTransport,
		Credentials: creds,
		Scopes:      []auth.Scope{scope},
		ClientID:    registry.AuthClientID,
	}
	tokenHandler := auth.NewTokenHandlerWithOptions(tokenHandlerOptions)
	basicHandler := auth.NewBasicHandler(creds)
	tr := transport.NewTransport(base, auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler))
	return tr, nil
}
