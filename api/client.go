package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

var (
	errRedirect = errors.New("redirect")
)

// Config is used to configure the creation of the client.
type Config struct {
	// Address is the address of the Vault server. This should be a complete
	// URL such as "http://vault.example.com". If you need a custom SSL
	// cert or want to enable insecure mode, you need to specify a custom
	// HttpClient.
	Address string

	// HttpClient is the HTTP client to use. http.DefaultClient will be
	// used if not specified. The HTTP client must have the cookie jar set
	// to be able to store cookies, otherwise authentication (login) will
	// not work properly. If the jar is nil, a default empty cookie jar
	// will be set.
	HttpClient *http.Client
}

// DefaultConfig returns a default configuration for the client. It is
// safe to modify the return value of this function.
// The default Address is https://127.0.0.1:8200
func DefaultConfig() *Config {
	return &Config{
		Address:    "https://127.0.0.1:8200",
		HttpClient: &http.Client{},
	}
}

// Client is the client to the Vault API. Create a client with NewClient.
type Client struct {
	addr   *url.URL
	config *Config
}

// NewClient returns a new client for the given configuration.
func NewClient(c *Config) (*Client, error) {
	u, err := url.Parse(c.Address)
	if err != nil {
		return nil, err
	}

	if c.HttpClient == nil {
		c.HttpClient = http.DefaultClient
	}

	// If no cookie jar is set on the client, we set a default empty
	// cookie jar.
	if c.HttpClient.Jar == nil {
		jar, err := cookiejar.New(&cookiejar.Options{})
		if err != nil {
			return nil, err
		}

		c.HttpClient.Jar = jar
	}

	// Ensure redirects are not automatically followed
	c.HttpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errRedirect
	}

	client := &Client{
		addr:   u,
		config: c,
	}

	return client, nil
}

// NewRequest creates a new raw request object to query the Vault server
// configured for this client. This is an advanced method and generally
// doesn't need to be called externally.
func (c *Client) NewRequest(method, path string) *Request {
	return &Request{
		Method: method,
		URL: &url.URL{
			Scheme: c.addr.Scheme,
			Host:   c.addr.Host,
			Path:   path,
		},
		Params: make(map[string][]string),
	}
}

// RawRequest performs the raw request given. This request may be against
// a Vault server not configured with this client. This is an advanced operation
// that generally won't need to be called externally.
func (c *Client) RawRequest(r *Request) (*Response, error) {
	redirectCount := 0
START:
	req, err := r.ToHTTP()
	if err != nil {
		return nil, err
	}

	var result *Response
	resp, err := c.config.HttpClient.Do(req)
	if resp != nil {
		result = &Response{Response: resp}
	}
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Err == errRedirect {
			err = nil
		} else if strings.Contains(err.Error(), "tls: oversized") {
			err = fmt.Errorf(
				"%s\n\n"+
					"This error usually means that the server is running with TLS disabled\n"+
					"but the client is configured to use TLS. Please either enable TLS\n"+
					"on the server or run the client with -address set to an address\n"+
					"that uses the http protocol:\n\n"+
					"    vault <command> -address http://<address>\n\n"+
					"You can also set the VAULT_ADDR environment variable:\n\n\n"+
					"    VAULT_ADDR=http://<address> vault <command>\n\n"+
					"where <address> is replaced by the actual address to the server.",
				err)
		}
	}
	if err != nil {
		return result, err
	}

	// Check for a redirect, only allowing for a single redirect
	if (resp.StatusCode == 302 || resp.StatusCode == 307) && redirectCount == 0 {
		// Parse the updated location
		respLoc, err := resp.Location()
		if err != nil {
			return result, err
		}

		// Ensure a protocol downgrade doesn't happen
		if req.URL.Scheme == "https" && respLoc.Scheme != "https" {
			return result, fmt.Errorf("redirect would cause protocol downgrade")
		}

		// Copy the cookies so that our client auth transfers
		cookies := c.config.HttpClient.Jar.Cookies(r.URL)
		c.config.HttpClient.Jar.SetCookies(respLoc, cookies)

		// Update the request
		r.URL = respLoc

		// Reset the request body if any
		if err := r.ResetJSONBody(); err != nil {
			return result, err
		}

		// Retry the request
		redirectCount++
		goto START
	}

	if err := result.Error(); err != nil {
		return result, err
	}

	return result, nil
}
