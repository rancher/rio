package webhook

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"

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
	request2 "k8s.io/apiserver/pkg/endpoints/request"
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
	hostportVerb    = "rio-hostport"
	hostnetworkVerb = "rio-hostnetwork"
	privilegedVerb  = "rio-privileged"
	hostmountVerb   = "rio-hostmount"
	servicemeshVerb = "rio-servicemesh"
)

func SetupValidatingWebhook(rContext *types.Context) error {
	secret, err := rContext.Core.Core().V1().Secret().Get(rContext.Namespace, constants.AuthWebhookSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	devHost := os.Getenv("LOCALHOST")
	if devHost == "" {
		devHost = "localhost"
	}
	if len(secret.Data) == 0 {
		hostname := fmt.Sprintf("%s.%s.svc", constants.AuthWebhookServiceName, rContext.Namespace)
		if constants.DevMode != "" {
			hostname = devHost
		}
		webhookCa, err := tls.GenerateRootCAWithDefaults(hostname)
		if err != nil {
			return err
		}
		secret.Data = map[string][]byte{
			corev1.TLSPrivateKeyKey: []byte(webhookCa.Cred.EncodePrivateKeyPEM()),
			corev1.TLSCertKey:       []byte(webhookCa.Cred.EncodeCertificatePEM()),
			"ca":                    []byte(webhookCa.Cred.EncodeCertificatePEM()),
		}
		if _, err := rContext.Core.Core().V1().Secret().Update(secret); err != nil {
			return err
		}
	}

	if constants.DevMode != "" {
		tlsKeyPath = fmt.Sprintf("%s/.local/ssl/tls.key", os.Getenv("HOME"))
		tlsCrtPath = fmt.Sprintf("%s/.local/ssl/tls.crt", os.Getenv("HOME"))
		if err := os.MkdirAll(filepath.Dir(tlsKeyPath), 0755); err != nil {
			return err
		}
		if err := ioutil.WriteFile(tlsKeyPath, secret.Data[corev1.TLSPrivateKeyKey], 0755); err != nil {
			return err
		}

		if err := ioutil.WriteFile(tlsCrtPath, secret.Data[corev1.TLSCertKey], 0755); err != nil {
			return err
		}
	}

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
				Name: "auth-webhook.rio.io",
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
						Namespace: rContext.Namespace,
						Name:      "auth-webhook",
					},
					CABundle: secret.Data["ca"],
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
							Resources:   []string{rioService},
							APIVersions: []string{"v1"},
						},
					},
				},
			},
		},
	}

	if constants.DevMode != "" {
		validatingWebhook.Webhooks[0].ClientConfig = v1beta1.WebhookClientConfig{
			URL:      &[]string{fmt.Sprintf("https://%s%s", devHost, constants.DevWebhookPort)}[0],
			CABundle: secret.Data["ca"],
		}
	}

	webhook, err := rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Get(validatingWebhook.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if _, err = rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Create(validatingWebhook); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		webhook.Webhooks = validatingWebhook.Webhooks
		webhook.Name = validatingWebhook.Name
		if _, err := rContext.K8s.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations().Update(webhook); err != nil {
			return err
		}
	}

	return nil
}

func Run(kc string, rContext *types.Context, port string) error {
	k8sAPI, err := k8s.InitializeAPI(kc)
	if err != nil {
		log.Fatalf("failed to initialize Kubernetes API: %s", err)
	}

	rbacRestGetter := rbacRestGetter{
		Interface: rContext.RBAC.Rbac(),
	}

	ruleResolver := rbacregistryvalidation.NewDefaultRuleResolver(rbacRestGetter, rbacRestGetter, rbacRestGetter, rbacRestGetter)

	cred, err := tls.ReadPEMCreds(tlsKeyPath, tlsCrtPath)
	if err != nil {
		return fmt.Errorf("failed to read TLS secrets: %s", err)
	}

	h := handler{
		systemNamespace:      rContext.Namespace,
		ruleSolver:           ruleResolver,
		ignoreServiceAccount: fmt.Sprintf(rioAdminAccount, rContext.Namespace),
	}

	s, err := webhook.NewServer(k8sAPI, port, cred, h.ValidateAuth, constants.AuthWebhookServiceName)
	if err != nil {
		log.Fatalf("failed to initialize the webhook server: %s", err)
	}

	go s.Start()
	return nil
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

	var service riov1.Service
	err := yaml.UnmarshalStrict(request.Object.Raw, &service)
	if err != nil {
		return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
	}
	// todo: compare spec, it can be removed after status is moved to subresource
	if request.Operation == admissionv1beta1.Update {
		var oldService riov1.Service
		err := yaml.UnmarshalStrict(request.OldObject.Raw, &oldService)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service, parsing old object: %v", err)
		}
		if reflect.DeepEqual(service.Spec, oldService.Spec) {
			admissionResponse.Allowed = true
			return admissionResponse, nil
		}
	}

	var globalRules, rules []rbacv1.PolicyRule

	for _, p := range service.Spec.GlobalPermissions {
		policyRules, err := h.permToPolicyRule(p, true, "")
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
		globalRules = append(globalRules, policyRules...)
	}

	for _, p := range service.Spec.Permissions {
		policyRules, err := h.permToPolicyRule(p, false, service.Namespace)
		if err != nil {
			return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
		}
		rules = append(rules, policyRules...)
	}

	rules = append(rules, convertSecurityPolicyRule(service)...)

	var userInfo = &user.DefaultInfo{
		Name:   request.UserInfo.Username,
		UID:    request.UserInfo.UID,
		Groups: request.UserInfo.Groups,
		Extra:  toExtra(request.UserInfo.Extra),
	}

	globaleCtx := request2.WithNamespace(request2.WithUser(context.Background(), userInfo), "")
	if err := rbacregistryvalidation.ConfirmNoEscalation(globaleCtx, h.ruleSolver, globalRules); err != nil {
		return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
	}

	ctx := request2.WithNamespace(request2.WithUser(context.Background(), userInfo), service.Namespace)
	if err := rbacregistryvalidation.ConfirmNoEscalation(ctx, h.ruleSolver, rules); err != nil {
		return admissionResponse, fmt.Errorf("failed to validate rio service: %v", err)
	}

	admissionResponse.Allowed = true
	return admissionResponse, nil
}

func convertSecurityPolicyRule(service riov1.Service) []rbacv1.PolicyRule {
	var policyRules []rbacv1.PolicyRule

	var useHostPort, useHostNetwork, usePriviledged, useHostMount, useServiceMesh bool

	// checking hostport
	ports := service.Spec.Ports
	for _, c := range service.Spec.Sidecars {
		ports = append(ports, c.Ports...)
	}
	for _, p := range ports {
		if p.HostPort {
			useHostPort = true
			break
		}
	}

	// todo: add privileged, hostmount

	containers := []riov1.Container{service.Spec.Container}
	for _, c := range service.Spec.Sidecars {
		containers = append(containers, c.Container)
	}

	if service.Spec.HostNetwork {
		useHostNetwork = true
	}

	if !service.Spec.DisableServiceMesh {
		useServiceMesh = true
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

	if useHostMount {
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
