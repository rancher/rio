package model

import (
	"encoding/json"
	"net/http"
	"time"
)

type MigrateRecord struct {
	Fqdn       string              `json:"fqdn"`
	Hosts      []string            `json:"hosts"`
	SubDomain  map[string][]string `json:"subdomain"`
	Text       string              `json:"text"`
	Token      string              `json:"token"`
	Expiration *time.Time          `json:"expiration"`
}

type MigrateFrozen struct {
	Path       string     `json:"path"`
	Expiration *time.Time `json:"expiration"`
}

type MigrateToken struct {
	Path       string     `json:"path"`
	Token      string     `json:"token"`
	Expiration *time.Time `json:"expiration"`
}

func ParseMigrateRecord(r *http.Request) (*MigrateRecord, error) {
	var opts MigrateRecord
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	return &opts, err
}

func ParseMigrateFrozen(r *http.Request) (*MigrateFrozen, error) {
	var opts MigrateFrozen
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	return &opts, err
}

func ParseMigrateToken(r *http.Request) (*MigrateToken, error) {
	var opts MigrateToken
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&opts)
	return &opts, err
}
