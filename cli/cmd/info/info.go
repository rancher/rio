package info

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/version"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Info(app *cli.App) cli.Command {
	return cli.Command{
		Name:   "info",
		Usage:  "Display System-Wide Information",
		Action: clicontext.DefaultAction(info),
	}
}

func info(ctx *clicontext.CLIContext) error {
	builder := &strings.Builder{}

	info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
	if err != nil {
		return err
	}

	clusterDomains, err := ctx.Project.ClusterDomains().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(clusterDomains.Items) == 0 {
		return fmt.Errorf("no system information is generated")
	}

	clusterDomain := clusterDomains.Items[0]

	var addresses []string
	for _, d := range clusterDomain.Spec.Addresses {
		addresses = append(addresses, d.IP)
	}

	builder.WriteString(fmt.Sprintf("Rio Version: %s (%s)\n", info.Status.Version, info.Status.GitCommit))
	builder.WriteString(fmt.Sprintf("Rio CLI Version: %s (%s)\n", version.Version, version.GitCommit))
	builder.WriteString(fmt.Sprintf("Cluster Domain: %s\n", clusterDomain.Name))
	builder.WriteString(fmt.Sprintf("Cluster Domain IPs: %s\n", strings.Join(addresses, ",")))
	builder.WriteString(fmt.Sprintf("System Namespace: %s\n", info.Status.SystemNamespace))
	builder.WriteString(fmt.Sprintf("Wildcard certificates: %v\n", clusterDomain.Status.HTTPSSupported))

	var keys []string
	for k := range info.Status.SystemComponentReadyMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		builder.WriteString(fmt.Sprintf("%v status: %v\n", k, info.Status.SystemComponentReadyMap[k]))
	}
	fmt.Println(builder.String())

	return nil
}
