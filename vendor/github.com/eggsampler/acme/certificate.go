package acme

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
)

// FetchCertificates downloads a certificate chain from a url given in an order certificate.
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.4.2
func (c Client) FetchCertificates(certificateURL string) ([]*x509.Certificate, error) {
	resp, raw, err := c.getRaw(certificateURL, http.StatusOK)
	if err != nil {
		return nil, err
	}

	var certs []*x509.Certificate
	for {
		var p *pem.Block
		p, raw = pem.Decode(raw)
		if p == nil {
			break
		}
		cert, err := x509.ParseCertificate(p.Bytes)
		if err != nil {
			return certs, fmt.Errorf("acme: error parsing certificate: %v", err)
		}
		certs = append(certs, cert)
	}

	up := fetchLink(resp, "up")
	if up != "" {
		upCerts, err := c.FetchCertificates(up)
		if err != nil {
			return certs, fmt.Errorf("acme: error fetching up cert: %v", err)
		}
		if len(upCerts) != 0 {
			certs = append(certs, upCerts...)
		}
	}

	return certs, nil
}

// RevokeCertificate revokes a given certificate given the certificate key or account key, and a reason.
// More details: https://tools.ietf.org/html/draft-ietf-acme-acme-10#section-7.6
func (c Client) RevokeCertificate(account Account, cert *x509.Certificate, key crypto.Signer, reason int) error {
	revokeReq := struct {
		Certificate string `json:"certificate"`
		Reason      int    `json:"reason"`
	}{
		Certificate: base64.RawURLEncoding.EncodeToString(cert.Raw),
		Reason:      reason,
	}

	kid := ""
	if key == account.PrivateKey {
		kid = account.URL
	}

	if _, err := c.post(c.dir.RevokeCert, kid, key, revokeReq, nil, http.StatusOK); err != nil {
		return err
	}

	return nil
}
