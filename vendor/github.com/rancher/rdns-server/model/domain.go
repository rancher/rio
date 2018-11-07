package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Domain struct {
	Fqdn       string     `json:"fqdn,omitempty"`
	Hosts      []string   `json:"hosts,omitempty"`
	Text       string     `json:"text,omitempty"`
	Expiration *time.Time `json:"expiration,omitempty"`
}

func (d *Domain) String() string {
	if d.Text != "" {
		return fmt.Sprintf("{Fqdn: %s, Text: %s, Expiration: %s}", d.Fqdn, d.Text, d.Expiration.String())
	}
	return fmt.Sprintf("{Fqdn: %s, Hosts: %s, Expiration: %s}", d.Fqdn, d.Hosts, d.Expiration.String())
}

type DomainOptions struct {
	Fqdn  string   `json:"fqdn"`
	Hosts []string `json:"hosts"`
	Text  string   `json:"text"`
}

func (d *DomainOptions) String() string {
	if d.Text != "" {
		return fmt.Sprintf("{Fqdn: %s, Text: %s}", d.Fqdn, d.Text)
	}
	return fmt.Sprintf("{Fqdn: %s, Hosts: %s}", d.Fqdn, d.Hosts)
}

func ParseDomainOptions(r *http.Request) (*DomainOptions, error) {
	var opts DomainOptions
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	return &opts, err
}
