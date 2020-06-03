package pkg

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"

	"github.com/rancher/rio/pkg/constants"

	"github.com/rancher/rio/pkg/constructors"

	v1 "k8s.io/api/core/v1"

	"github.com/go-acme/lego/v3/registration"
)

type User struct {
	Name         string
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
	URL          string
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.Key
}

func FromSecret(secret *v1.Secret) (*User, error) {
	block, _ := pem.Decode(secret.Data["privateKey"])
	x509Encoded := block.Bytes
	privateKey, err := x509.ParseECPrivateKey(x509Encoded)
	if err != nil {
		return nil, err
	}

	var reg registration.Resource
	if err := json.Unmarshal(secret.Data["registration"], &reg); err != nil {
		return nil, err
	}

	return &User{
		Name:         secret.Name,
		Email:        string(secret.Data["email"]),
		Registration: &reg,
		Key:          privateKey,
		URL:          string(secret.Data["url"]),
	}, nil
}

func SetSecret(namespace string, user *User) (*v1.Secret, error) {
	x509Encoded, _ := x509.MarshalECPrivateKey(user.Key.(*ecdsa.PrivateKey))
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	reg, err := json.Marshal(user.Registration)
	if err != nil {
		return nil, err
	}

	secret := constructors.NewSecret(namespace, constants.LetsEncryptAccountSecretName, v1.Secret{
		Data: map[string][]byte{
			"email":        []byte(user.Email),
			"privateKey":   pemEncoded,
			"url":          []byte(user.URL),
			"registration": reg,
		},
	})

	return secret, nil
}
