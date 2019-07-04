package info

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/install"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/constants"
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

	clusterDomain, err := ctx.Project.ClusterDomains(ctx.SystemNamespace).Get(constants.ClusterDomainName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	var addresses []string
	for _, d := range clusterDomain.Spec.Addresses {
		addresses = append(addresses, d.IP)
	}

	builder.WriteString(fmt.Sprintf("Rio Version: %s (%s)\n", info.Status.Version, info.Status.GitCommit))
	builder.WriteString(fmt.Sprintf("Rio CLI Version: %s (%s)\n", version.Version, version.GitCommit))
	builder.WriteString(fmt.Sprintf("Cluster Domain: %s\n", clusterDomain.Status.ClusterDomain))
	builder.WriteString(fmt.Sprintf("Cluster Domain IPs: %s\n", strings.Join(addresses, ",")))
	builder.WriteString(fmt.Sprintf("System Namespace: %s\n", info.Status.SystemNamespace))
	builder.WriteString("\n")
	builder.WriteString("System Components:\n")
	builder.WriteString(fmt.Sprintf("Autoscaler status: %v\n", info.Status.SystemComponentReadyMap[install.Autoscaler]))
	builder.WriteString(fmt.Sprintf("BuildController status: %v\n", info.Status.SystemComponentReadyMap[install.BuildController]))
	builder.WriteString(fmt.Sprintf("CertManager status: %v\n", info.Status.SystemComponentReadyMap[install.CertManager]))
	builder.WriteString(fmt.Sprintf("Grafana status: %v\n", info.Status.SystemComponentReadyMap[install.Grafana]))
	builder.WriteString(fmt.Sprintf("IstioCitadel status: %v\n", info.Status.SystemComponentReadyMap[install.IstioCitadel]))
	builder.WriteString(fmt.Sprintf("IstioPilot status: %v\n", info.Status.SystemComponentReadyMap[install.IstioPilot]))
	builder.WriteString(fmt.Sprintf("IstioTelemetry status: %v\n", info.Status.SystemComponentReadyMap[install.IstioTelemetry]))
	builder.WriteString(fmt.Sprintf("Kiali status: %v\n", info.Status.SystemComponentReadyMap[install.Kiali]))
	builder.WriteString(fmt.Sprintf("Prometheus status: %v\n", info.Status.SystemComponentReadyMap[install.Prometheus]))
	builder.WriteString(fmt.Sprintf("Registry status: %v\n", info.Status.SystemComponentReadyMap[install.Registry]))
	builder.WriteString(fmt.Sprintf("Webhook status: %v", info.Status.SystemComponentReadyMap[install.Webhook]))
	fmt.Println(builder.String())

	return nil
}
