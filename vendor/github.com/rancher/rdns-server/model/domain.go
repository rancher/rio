package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Domain struct {
	Fqdn       string              `json:"fqdn,omitempty"`
	Hosts      []string            `json:"hosts,omitempty"`
	SubDomain  map[string][]string `json:"subdomain,omitempty"`
	Text       string              `json:"text,omitempty"`
	CNAME      string              `json:"cname,omitempty"`
	Expiration *time.Time          `json:"expiration,omitempty"`
}

func (d *Domain) String() string {
	if d.CNAME != "" {
		return fmt.Sprintf("{Fqdn: %s, CNAME: %s, Expiration: %s}", d.Fqdn, d.CNAME, d.Expiration.Format(time.RFC3339Nano))
	}
	if d.Text != "" {
		return fmt.Sprintf("{Fqdn: %s, Text: %s, Expiration: %s}", d.Fqdn, d.Text, d.Expiration.Format(time.RFC3339Nano))
	}
	if len(d.SubDomain) > 0 {
		return fmt.Sprintf("{Fqdn: %s, Hosts: %s, SubDomain: %s, Expiration: %s}", d.Fqdn, d.Hosts, mapToString(d.SubDomain), d.Expiration.Format(time.RFC3339Nano))
	}
	return fmt.Sprintf("{Fqdn: %s, Hosts: %s, Expiration: %s}", d.Fqdn, d.Hosts, d.Expiration.Format(time.RFC3339Nano))
}

type DomainOptions struct {
	Fqdn      string              `json:"fqdn"`
	Hosts     []string            `json:"hosts"`
	SubDomain map[string][]string `json:"subdomain"`
	Text      string              `json:"text"`
	CNAME     string              `json:"cname"`
	Normal    bool                `json:"normal"`
}

func (d *DomainOptions) String() string {
	if d.CNAME != "" {
		return fmt.Sprintf("{Fqdn: %s, CNAME: %s}", d.Fqdn, d.CNAME)
	}
	if d.Text != "" {
		return fmt.Sprintf("{Fqdn: %s, Text: %s}", d.Fqdn, d.Text)
	}
	if len(d.SubDomain) > 0 {
		return fmt.Sprintf("{Fqdn: %s, Hosts: %s, SubDomain: %s}", d.Fqdn, d.Hosts, mapToString(d.SubDomain))
	}
	return fmt.Sprintf("{Fqdn: %s, Hosts: %s}", d.Fqdn, d.Hosts)
}

func ParseDomainOptions(r *http.Request) (*DomainOptions, error) {
	var opts DomainOptions
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	return &opts, err
}

func mapToString(m map[string][]string) string {
	b, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(b)
}
