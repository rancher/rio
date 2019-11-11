package webhook

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/rancher/rio/pkg/config"

	"github.com/linkerd/linkerd2/controller/k8s"
	"github.com/linkerd/linkerd2/controller/webhook"
	"github.com/linkerd/linkerd2/pkg/tls"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/types"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/tools/record"
	rbacregistryvalidation "k8s.io/kubernetes/pkg/registry/rbac/validation"
	"sigs.k8s.io/yaml"
)

var (
	tlsKeyPath = "/var/run/rio/ssl/tls.key"
	tlsCrtPath = "/var/run/rio/ssl/tls.crt"

	rioAdminAccount = "system:serviceaccount:%s:rio-controller-serviceaccount"

	rioAPIGroup     = "rio.cattle.io"
	rioService      = "services"
	rioStack        = "stacks"
	hostportVerb    = "rio-hostport"
	hostnetworkVerb = "rio-hostnetwork"
	privilegedVerb  = "rio-privileged"
	hostmountVerb   = "rio-hostpath"
	servicemeshVerb = "rio-servicemesh"
)

type Webhook struct {
	rContext   *types.Context
	kubeconfig string
	listenHost string
	port       string
}

func New(rContext *types.Context, kc string) Webhook {
	w := Webhook{
		rContext:   rContext,
		kubeconfig: kc,
		listenHost: fmt.Sprintf("%s.%s.svc", constants.AuthWebhookServiceName, rContext.Namespace),
		port:       ":443",
	}
	if config.ConfigController.WebhookHost != "" {
		w.listenHost = config.ConfigController.WebhookHost
	}

	if config.ConfigController.WebhookPort != "" {
		w.port = config.ConfigController.WebhookPort
	}

	return w
}

func (w Webhook) Setup() error {
	if !config.ConfigController.RunAPIValidatorWebhook {
		return nil
	}
	if err := w.setup(); err != nil {
		return err
	}
	return w.run()
}

type handler struct {
	systemNamespace      string
	ruleSolver           rbacregistryvalidation.AuthorizationRuleResolver
	ignoreServiceAccount string
}

func (h handler) ValidateAuth(api *k8s.API, request *admissionv1beta1.AdmissionRequest, _ record.EventRecorder) (*admissionv1beta1.AdmissionResponse, error) {
	admissionResponse := &admissionv1beta1.AdmissionResponse{Allowed: false}
	if request.UserInfo.Username == h.ignoreServiceAccount {
		admissionResponse.Allowed = true
		return admissionResponse, nil
	}

	if request.Operation == admissionv1beta1.Update && request.SubResource == "status" {
		admissionResponse.Allowed = true
		return admissionResponse, nil
	}

	var globalRules, rules []rbacv1.PolicyRule
	if request.Kind.Kind == "Service" {
		var service riov1.Service
		err := yaml.Unmarshal(request.Object.Raw, &service)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
		globalRules, rules, err = h.convertServiceToRule(service)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
	} else if request.Kind.Kind == "Stack" {
		var stack riov1.Stack
		err := yaml.Unmarshal(request.Object.Raw, &stack)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
		globalRules, rules, err = h.convertStackToRule(stack)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
	}

	var userInfo = &user.DefaultInfo{
		Name:   request.UserInfo.Username,
		UID:    request.UserInfo.UID,
		Groups: request.UserInfo.Groups,
		Extra:  toExtra(request.UserInfo.Extra),
	}

	globaleCtx := k8srequest.WithNamespace(k8srequest.WithUser(context.Background(), userInfo), "")
	if err := rbacregistryvalidation.ConfirmNoEscalation(globaleCtx, h.ruleSolver, globalRules); err != nil {
		return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
	}

	ctx := k8srequest.WithNamespace(k8srequest.WithUser(context.Background(), userInfo), request.Namespace)
	if err := rbacregistryvalidation.ConfirmNoEscalation(ctx, h.ruleSolver, rules); err != nil {
		return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
	}

	admissionResponse.Allowed = true
	return admissionResponse, nil
}

func (w Webhook) setup() error {
	caBundle, err := w.reconcileSecret()
	if err != nil {
		return err
	}

	return w.reconcileWebhook(caBundle)
}

func (w Webhook) reconcileSecret() ([]byte, error) {
	secret, err := w.rContext.Core.Core().V1().Secret().Get(w.rContext.Namespace, constants.AuthWebhookSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if len(secret.Data) == 0 {
		hostname := w.listenHost
		webhookCa, err := tls.GenerateRootCAWithDefaults(hostname)
		if err != nil {
			return nil, err
		}
		secret.Data = map[string][]byte{
			corev1.TLSPrivateKeyKey: []byte(webhookCa.Cred.EncodePrivateKeyPEM()),
			corev1.TLSCertKey:       []byte(webhookCa.Cred.EncodeCertificatePEM()),
			"ca":                    []byte(webhookCa.Cred.EncodeCertificatePEM()),
		}
		if _, err := w.rContext.Core.Core().V1().Secret().Update(secret); err != nil {
			return nil, err
		}
	}

	// if WebhookHost is not empty set up webhook locally
	if config.ConfigController.WebhookHost != "" {
		tlsKeyPath = fmt.Sprintf("%s/.local/ssl/tls.key", os.Getenv("HOME"))
		tlsCrtPath = fmt.Sprintf("%s/.local/ssl/tls.crt", os.Getenv("HOME"))
		if err := os.MkdirAll(filepath.Dir(tlsKeyPath), 0755); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(tlsKeyPath, secret.Data[corev1.TLSPrivateKeyKey], 0755); err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(tlsCrtPath, secret.Data[corev1.TLSCertKey], 0755); err != nil {
			return nil, err
		}
	}
	return secret.Data["ca"], nil
}

func (w Webhook) reconcileWebhook(caBundle []byte) error {
	failPolicy := v1beta1.Fail
	validatingWebhook := &v1beta1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.AuthWebhookServiceName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1beta1",
			Kind:       "ValidatingWebhookConfiguration",
		},
		Webhooks: []v1beta1.ValidatingWebhook{
			{
				Name: "api-validator.rio.io",
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "rio.cattle.io/is-system",
							Operator: metav1.LabelSelectorOpDoesNotExist,
						},
					},
				},
				ClientConfig: v1beta1.WebhookClientConfig{
					Service: &v1beta1.ServiceReference{
						Namespace: w.rContext.Namespace,
						Name:      constants.AuthWebhookServiceName,
					},
					CABundle: caBundle,
				},
				FailurePolicy: &failPolicy,
				Rules: []v1beta1.RuleWithOperations{
					{
						Operations: []v1beta1.OperationType{
							v1beta1.Create,
							v1beta1.Update,
						},
						Rule: v1beta1.Rule{
							APIGroups:   []string{rioAPIGroup},
							Resources:   []string{rioService, rioStack},
							APIVersions: []string{"v1"},
						},
					},
				},
			},
		},
	}

	// if WebhookHost is not empty set up webhook locally
	if config.ConfigController.WebhookHost != "" {
		validatingWebhook.Webhooks[0].ClientConfig = v1beta1.WebhookClientConfig{
			URL:      &[]string{fmt.Sprintf("https://%s%s", w.listenHost, constants.DevWebhookPort)}[0],
			CABundle: caBundle,
		}
		ignorePolicy := v1beta1.Ignore
		validatingWebhook.Webhooks[0].FailurePolicy = &ignorePolicy
	}

	webhook, err := w.rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(validatingWebhook.Name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		_, err = w.rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(validatingWebhook)
		return err
	} else if err == nil {
		webhook.Webhooks = validatingWebhook.Webhooks
		webhook.Name = validatingWebhook.Name
		if _, err := w.rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(webhook); err != nil {
			return err
		}
	}
	return err
}

func (w Webhook) run() error {
	k8sAPI, err := k8s.InitializeAPI(w.kubeconfig)
	if err != nil {
		log.Fatalf("failed to initialize Kubernetes API: %s", err)
	}

	rbacRestGetter := rbacRestGetter{
		Interface: w.rContext.RBAC.Rbac(),
	}

	ruleResolver := rbacregistryvalidation.NewDefaultRuleResolver(rbacRestGetter, rbacRestGetter, rbacRestGetter, rbacRestGetter)

	cred, err := tls.ReadPEMCreds(tlsKeyPath, tlsCrtPath)
	if err != nil {
		return fmt.Errorf("failed to read TLS secrets: %s", err)
	}

	h := handler{
		systemNamespace:      w.rContext.Namespace,
		ruleSolver:           ruleResolver,
		ignoreServiceAccount: fmt.Sprintf(rioAdminAccount, w.rContext.Namespace),
	}

	s, err := webhook.NewServer(k8sAPI, w.port, cred, h.ValidateAuth, constants.AuthWebhookServiceName)
	if err != nil {
		log.Fatalf("failed to initialize the webhook server: %s", err)
	}

	go s.Start()
	return nil
}

func (h handler) convertServiceToRule(service riov1.Service) (globalRules []rbacv1.PolicyRule, rules []rbacv1.PolicyRule, err error) {
	for _, p := range service.Spec.GlobalPermissions {
		policyRules, err := h.permToPolicyRule(p, true, "")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate rio service: %v", err)
		}
		globalRules = append(globalRules, policyRules...)
	}

	for _, p := range service.Spec.Permissions {
		policyRules, err := h.permToPolicyRule(p, false, service.Namespace)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate rio service: %v", err)
		}
		rules = append(rules, policyRules...)
	}

	rules = append(rules, convertSecurityPolicyRule(service)...)
	return globalRules, rules, nil
}

func (h handler) convertStackToRule(stack riov1.Stack) (globalRules []rbacv1.PolicyRule, rules []rbacv1.PolicyRule, err error) {
	for _, p := range stack.Spec.Permissions {
		policyRules, err := h.permToPolicyRule(p, false, stack.Namespace)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to validate rio service: %v", err)
		}
		rules = append(rules, policyRules...)
	}
	return globalRules, rules, nil
}

func convertSecurityPolicyRule(service riov1.Service) []rbacv1.PolicyRule {
	var policyRules []rbacv1.PolicyRule

	var useHostPort, useHostNetwork, usePriviledged, useHostPath, useServiceMesh bool

	containers := []riov1.Container{service.Spec.Container}
	for _, c := range service.Spec.Sidecars {
		containers = append(containers, c.Container)
	}

	for _, con := range containers {
		for _, p := range con.Ports {
			if p.HostPort {
				useHostPort = true
				break
			}
		}

		for _, v := range con.Volumes {
			if v.HostPath != "" {
				useHostPath = true
				break
			}
		}
	}

	if service.Spec.HostNetwork {
		useHostNetwork = true
	}

	if service.Spec.ServiceMesh != nil {
		useServiceMesh = true
	}

	if service.Spec.ContainerSecurityContext != nil && service.Spec.ContainerSecurityContext.Privileged != nil && *service.Spec.ContainerSecurityContext.Privileged {
		usePriviledged = true
	}

	if useHostPort {
		rule := newPolicyRule(service)
		rule.Verbs = []string{hostportVerb}
		policyRules = append(policyRules, rule)
	}

	if useHostNetwork {
		rule := newPolicyRule(service)
		rule.Verbs = []string{hostnetworkVerb}
		policyRules = append(policyRules, rule)
	}

	if useHostPath {
		rule := newPolicyRule(service)
		rule.Verbs = []string{hostmountVerb}
		policyRules = append(policyRules, rule)
	}

	if usePriviledged {
		rule := newPolicyRule(service)
		rule.Verbs = []string{privilegedVerb}
		policyRules = append(policyRules, rule)
	}

	if useServiceMesh {
		rule := newPolicyRule(service)
		rule.Verbs = []string{servicemeshVerb}
		policyRules = append(policyRules, rule)
	}

	return policyRules
}

func newPolicyRule(service riov1.Service) rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		APIGroups:     []string{rioAPIGroup},
		Resources:     []string{rioService},
		ResourceNames: []string{service.Name},
	}
}

func (h handler) permToPolicyRule(perm riov1.Permission, global bool, namespace string) ([]rbacv1.PolicyRule, error) {
	if perm.Role != "" {
		rolebindings := rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     perm.Role,
		}
		if global {
			rolebindings.Kind = "ClusterRole"
		} else {
			rolebindings.Kind = "Role"
		}

		rules, err := h.ruleSolver.GetRoleReferenceRules(rolebindings, namespace)
		if err != nil {
			return nil, err
		}
		return rules, nil
	}

	var policyRule rbacv1.PolicyRule
	policyRule.Verbs = perm.Verbs
	if perm.URL == "" {
		if perm.ResourceName != "" {
			policyRule.ResourceNames = []string{perm.ResourceName}
		}

		policyRule.APIGroups = []string{perm.APIGroup}

		if perm.Resource != "" {
			policyRule.Resources = []string{perm.Resource}
		}
	} else {
		policyRule.NonResourceURLs = []string{perm.URL}
	}

	return []rbacv1.PolicyRule{policyRule}, nil
}

func toExtra(extra map[string]v1.ExtraValue) map[string][]string {
	r := map[string][]string{}
	for k, v := range extra {
		r[k] = v
	}
	return r
}
