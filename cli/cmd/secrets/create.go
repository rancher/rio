package secrets

import (
	"fmt"
	"io/ioutil"

	"github.com/rancher/wrangler/pkg/kv"

	v1 "k8s.io/api/core/v1"

	"github.com/rancher/rio/cli/pkg/stack"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/constructors"
)

type Create struct {
	T_Type     string   `desc:"Create type" default:"Opaque"`
	F_FromFile []string `desc:"Creating secrets from files"`
	D_Data     []string `desc:"Creating secrets from key-pair data"`
}

func (s *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("exact one argument is required")
	}

	namespace, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])
	secret := constructors.NewSecret(namespace, name, v1.Secret{
		Type:       v1.SecretType(s.T_Type),
		Data:       make(map[string][]byte),
		StringData: make(map[string]string),
	})
	for _, f := range s.F_FromFile {
		k, v := kv.Split(f, "=")
		content, err := ioutil.ReadFile(v)
		if err != nil {
			return err
		}
		secret.Data[k] = content
	}

	for _, d := range s.D_Data {
		k, v := kv.Split(d, "=")
		secret.StringData[k] = v
	}
	if _, err := ctx.Core.Secrets(namespace).Create(secret); err != nil {
		return err
	}
	fmt.Printf("%s/%s\n", secret.Namespace, secret.Name)

	return nil
}
