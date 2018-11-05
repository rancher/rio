package acme

import (
	"net/http"

	"fmt"

	"crypto"
	"encoding/json"
)

// NewAccount registers a new account with the acme service
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.3
func (c Client) NewAccount(privateKey crypto.Signer, onlyReturnExisting, termsOfServiceAgreed bool, contact ...string) (Account, error) {
	newAccountReq := struct {
		OnlyReturnExisting   bool     `json:"onlyReturnExisting"`
		TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
		Contact              []string `json:"contact"`
	}{
		OnlyReturnExisting:   onlyReturnExisting,
		TermsOfServiceAgreed: termsOfServiceAgreed,
		Contact:              contact,
	}

	account := Account{}
	resp, err := c.post(c.dir.NewAccount, "", privateKey, newAccountReq, &account, http.StatusOK, http.StatusCreated)
	if err != nil {
		return account, err
	}

	account.URL = resp.Header.Get("Location")
	account.PrivateKey = privateKey

	if account.Thumbprint == "" {
		account.Thumbprint, err = JWKThumbprint(account.PrivateKey.Public())
		if err != nil {
			return account, fmt.Errorf("acme: error computing account thumbprint: %v", err)
		}
	}

	// Shouldn't be necessary anymore, but kept just in case
	// https://github.com/letsencrypt/boulder/pull/3811
	if account.Status == "" {
		if _, err := c.post(account.URL, account.URL, privateKey, struct{}{}, &account, http.StatusOK); err != nil {
			return account, fmt.Errorf("acme: error fetching existing account information: %v", err)
		}
	}

	return account, nil
}

// UpdateAccount updates an existing account with the acme service.
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.3.2
func (c Client) UpdateAccount(account Account, termsOfServiceAgreed bool, contact ...string) (Account, error) {
	updateAccountReq := struct {
		TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
		Contact              []string `json:"contact"`
	}{
		TermsOfServiceAgreed: termsOfServiceAgreed,
		Contact:              contact,
	}

	_, err := c.post(account.URL, account.URL, account.PrivateKey, updateAccountReq, &account, http.StatusOK)
	if err != nil {
		return account, err
	}

	if account.Thumbprint == "" {
		account.Thumbprint, err = JWKThumbprint(account.PrivateKey.Public())
		if err != nil {
			return account, fmt.Errorf("acme: error computing account thumbprint: %v", err)
		}
	}

	return account, nil
}

// AccountKeyChange rolls over an account to a new key.
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.3.6
func (c Client) AccountKeyChange(account Account, newPrivateKey crypto.Signer) (Account, error) {
	if c.dir.KeyChange == "" {
		return account, ErrUnsupported
	}

	oldJwkKeyPub, err := jwkEncode(account.PrivateKey.Public())
	if err != nil {
		return account, fmt.Errorf("acme: error encoding new private key: %v", err)
	}

	keyChangeReq := struct {
		Account string          `json:"account"`
		OldKey  json.RawMessage `json:"oldKey"`
	}{
		Account: account.URL,
		OldKey:  []byte(oldJwkKeyPub),
	}

	innerJws, err := jwsEncodeJSON(keyChangeReq, newPrivateKey, c.dir.KeyChange, "", "")
	if err != nil {
		return account, fmt.Errorf("acme: error encoding inner jws: %v", err)
	}

	if _, err := c.post(c.dir.KeyChange, account.URL, account.PrivateKey, json.RawMessage(innerJws), nil, http.StatusOK); err != nil {
		return account, err
	}

	account.PrivateKey = newPrivateKey

	return account, nil
}

// DeactivateAccount deactivates a given account.
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.3.7
func (c Client) DeactivateAccount(account Account) (Account, error) {
	deactivateReq := struct {
		Status string `json:"status"`
	}{
		Status: "deactivated",
	}

	_, err := c.post(account.URL, account.URL, account.PrivateKey, deactivateReq, &account, http.StatusOK)

	return account, err
}
