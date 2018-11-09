package acme

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"
)

// OptionFunc function prototype for passing options to NewClient
type OptionFunc func(client *Client) error

// WithHTTPTimeout sets a timeout on the http client used by the Client
func WithHTTPTimeout(duration time.Duration) OptionFunc {
	return func(client *Client) error {
		client.httpClient.Timeout = duration
		return nil
	}
}

// WithInsecureSkipVerify sets InsecureSkipVerify on the http client transport tls client config used by the Client
func WithInsecureSkipVerify() OptionFunc {
	return func(client *Client) error {
		client.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		return nil
	}
}

// WithUserAgentSuffix appends a user agent suffix for http requests to acme resources
func WithUserAgentSuffix(userAgentSuffix string) OptionFunc {
	return func(client *Client) error {
		client.userAgentSuffix = userAgentSuffix
		return nil
	}
}

// WithAcceptLanguage sets an Accept-Language header on http requests
func WithAcceptLanguage(acceptLanguage string) OptionFunc {
	return func(client *Client) error {
		client.acceptLanguage = acceptLanguage
		return nil
	}
}

func WithRetryCount(retryCount int) OptionFunc {
	return func(client *Client) error {
		if retryCount < 1 {
			return errors.New("retryCount must be > 0")
		}
		client.retryCount = retryCount
		return nil
	}
}
