package schema

import (
	"net/http"

	"k8s.io/api/core/v1"

	"github.com/rancher/norman/types"
	m "github.com/rancher/norman/types/mapper"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/rancher/types/factory"
	"github.com/rancher/types/mapper"
)

var (
	Version = types.APIVersion{
		Version: "v3",
		Group:   "management.cattle.io",
		Path:    "/v3",
	}

	Schemas = factory.Schemas(&Version).
		Init(nativeNodeTypes).
		Init(nodeTypes).
		Init(authzTypes).
		Init(clusterTypes).
		Init(catalogTypes).
		Init(authnTypes).
		Init(tokens).
		Init(schemaTypes).
		Init(userTypes).
		Init(projectNetworkPolicyTypes).
		Init(logTypes).
		Init(globalTypes).
		Init(rkeTypes).
		Init(alertTypes).
		Init(composeType)

	TokenSchemas = factory.Schemas(&Version).
			Init(tokens)
)

func rkeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.AddMapperForType(&Version, v3.BaseService{}, m.Drop{Field: "image"})
}

func schemaTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		MustImport(&Version, v3.DynamicSchema{})
}

func catalogTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v3.Catalog{},
			&m.Move{From: "catalogKind", To: "kind"},
			&m.Embed{Field: "status"},
			&m.Drop{Field: "helmVersionCommits"},
		).
		MustImportAndCustomize(&Version, v3.Catalog{}, func(schema *types.Schema) {
			schema.ResourceActions = map[string]types.Action{
				"refresh": {},
			}
			schema.CollectionActions = map[string]types.Action{
				"refresh": {},
			}
		}).
		AddMapperForType(&Version, v3.Template{},
			m.DisplayName{},
		).
		MustImport(&Version, v3.Template{}, struct {
			VersionLinks map[string]string
		}{}).
		MustImport(&Version, v3.TemplateVersion{}).
		MustImport(&Version, v3.TemplateContent{})
}

func nativeNodeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		TypeName("internalNodeStatus", v1.NodeStatus{}).
		TypeName("internalNodeSpec", v1.NodeSpec{}).
		AddMapperForType(&Version, v1.NodeStatus{},
			&mapper.NodeAddressMapper{},
			&mapper.OSInfo{},
			&m.Drop{Field: "addresses"},
			&m.Drop{Field: "daemonEndpoints"},
			&m.Drop{Field: "images"},
			&m.Drop{Field: "nodeInfo"},
			&m.Move{From: "conditions", To: "nodeConditions"},
			&m.Drop{Field: "phase"},
			&m.SliceToMap{Field: "volumesAttached", Key: "devicePath"},
		).
		AddMapperForType(&Version, v1.NodeSpec{},
			&m.Drop{Field: "externalID"},
			&m.Drop{Field: "configSource"},
			&m.Move{From: "providerID", To: "providerId"},
			&m.Move{From: "podCIDR", To: "podCidr"},
			m.Access{Fields: map[string]string{
				"podCidr":       "r",
				"providerId":    "r",
				"taints":        "ru",
				"unschedulable": "ru",
			}}).
		AddMapperForType(&Version, v1.Node{},
			&mapper.NodeAddressAnnotationMapper{}).
		MustImportAndCustomize(&Version, v1.NodeSpec{}, func(schema *types.Schema) {
			schema.CodeName = "InternalNodeSpec"
			schema.CodeNamePlural = "InternalNodeSpecs"
		}).
		MustImportAndCustomize(&Version, v1.NodeStatus{}, func(schema *types.Schema) {
			schema.CodeName = "InternalNodeStatus"
			schema.CodeNamePlural = "InternalNodeStatuses"
		}, struct {
			IPAddress         string
			ExternalIPAddress string `json:"externalIpAddress,omitempty"`
			Hostname          string
			Info              NodeInfo
		}{})
}

func clusterTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v3.Cluster{},
			&m.Embed{Field: "status"},
			m.DisplayName{},
		).
		AddMapperForType(&Version, v3.ClusterStatus{},
			m.Drop{Field: "serviceAccountToken"},
		).
		AddMapperForType(&Version, v3.ClusterEvent{}, &m.Move{
			From: "type",
			To:   "eventType",
		}).
		AddMapperForType(&Version, v3.ClusterRegistrationToken{},
			&m.Embed{Field: "status"},
		).
		AddMapperForType(&Version, v3.RancherKubernetesEngineConfig{},
			m.Drop{Field: "systemImages"},
		).
		MustImport(&Version, v3.Cluster{}).
		MustImport(&Version, v3.ClusterEvent{}).
		MustImport(&Version, v3.ClusterRegistrationToken{}).
		MustImport(&Version, v3.GenerateKubeConfigOutput{}).
		MustImport(&Version, v3.ImportClusterYamlInput{}).
		MustImport(&Version, v3.ImportYamlOutput{}).
		MustImport(&Version, v3.ExportOutput{}).
		MustImportAndCustomize(&Version, v3.ETCDService{}, func(schema *types.Schema) {
			schema.MustCustomizeField("extraArgs", func(field types.Field) types.Field {
				field.Default = map[string]interface{}{
					"election-timeout":   "5000",
					"heartbeat-interval": "500"}
				return field
			})
		}).
		MustImportAndCustomize(&Version, v3.Cluster{}, func(schema *types.Schema) {
			schema.MustCustomizeField("name", func(field types.Field) types.Field {
				field.Type = "dnsLabel"
				field.Nullable = true
				field.Required = false
				return field
			})
			schema.ResourceActions["generateKubeconfig"] = types.Action{
				Output: "generateKubeConfigOutput",
			}
			schema.ResourceActions["importYaml"] = types.Action{
				Input:  "importClusterYamlInput",
				Output: "importYamlOutput",
			}
			schema.ResourceActions["exportYaml"] = types.Action{
				Output: "exportOutput",
			}
		})
}

func authzTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		MustImport(&Version, v3.ProjectStatus{}).
		AddMapperForType(&Version, v3.Project{}, m.DisplayName{},
			&m.Embed{Field: "status"}).
		AddMapperForType(&Version, v3.GlobalRole{}, m.DisplayName{}).
		AddMapperForType(&Version, v3.RoleTemplate{}, m.DisplayName{}).
		AddMapperForType(&Version,
			v3.PodSecurityPolicyTemplateProjectBinding{},
			&mapper.NamespaceIDMapper{}).
		AddMapperForType(&Version, v3.ProjectRoleTemplateBinding{},
			&mapper.NamespaceIDMapper{},
		).
		MustImport(&Version, v3.SetPodSecurityPolicyTemplateInput{}).
		MustImport(&Version, v3.ImportYamlOutput{}).
		MustImportAndCustomize(&Version, v3.Project{}, func(schema *types.Schema) {
			schema.ResourceActions = map[string]types.Action{
				"setpodsecuritypolicytemplate": {
					Input:  "setPodSecurityPolicyTemplateInput",
					Output: "project",
				},
				"exportYaml": {},
			}
		}).
		MustImportAndCustomize(&Version, v3.GlobalRole{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
		}).
		MustImport(&Version, v3.GlobalRoleBinding{}).
		MustImport(&Version, v3.RoleTemplate{}).
		MustImport(&Version, v3.PodSecurityPolicyTemplate{}).
		MustImportAndCustomize(&Version, v3.PodSecurityPolicyTemplateProjectBinding{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet, http.MethodPost}
			schema.ResourceMethods = []string{}
		}).
		MustImport(&Version, v3.ClusterRoleTemplateBinding{}).
		MustImport(&Version, v3.ProjectRoleTemplateBinding{}).
		MustImport(&Version, v3.GlobalRoleBinding{})
}

func nodeTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v3.NodeSpec{}, &m.Embed{Field: "internalNodeSpec"}).
		AddMapperForType(&Version, v3.NodeStatus{},
			&m.Drop{Field: "nodeTemplateSpec"},
			&m.Embed{Field: "internalNodeStatus"},
			&m.SliceMerge{From: []string{"conditions", "nodeConditions"}, To: "conditions"}).
		AddMapperForType(&Version, v3.Node{},
			&m.Embed{Field: "status"},
			&m.Move{From: "rkeNode/user", To: "sshUser"},
			&m.ReadOnly{Field: "sshUser"},
			&m.Drop{Field: "rkeNode"},
			&m.Drop{Field: "labels"},
			&m.Drop{Field: "annotations"},
			&m.Move{From: "nodeLabels", To: "labels"},
			&m.Move{From: "nodeAnnotations", To: "annotations"},
			&m.Drop{Field: "desiredNodeLabels"},
			&m.Drop{Field: "desiredNodeAnnotations"},
			&m.Drop{Field: "desiredNodeUnschedulable"},
			&m.Drop{Field: "nodeDrainInput"},
			&m.AnnotationField{Field: "publicEndpoints", List: true},
			m.Copy{From: "namespaceId", To: "clusterName"},
			m.DisplayName{}).
		AddMapperForType(&Version, v3.NodeDriver{}, m.DisplayName{}).
		AddMapperForType(&Version, v3.NodeTemplate{}, m.DisplayName{}).
		MustImport(&Version, v3.PublicEndpoint{}).
		MustImport(&Version, v3.NodePool{}).
		MustImport(&Version, v3.NodeDrainInput{}).
		MustImportAndCustomize(&Version, v3.Node{}, func(schema *types.Schema) {
			labelField := schema.ResourceFields["labels"]
			labelField.Create = true
			labelField.Update = true
			schema.ResourceFields["labels"] = labelField
			unschedulable := schema.ResourceFields["unschedulable"]
			unschedulable.Create = false
			unschedulable.Update = false
			schema.ResourceFields["unschedulable"] = unschedulable
			clusterField := schema.ResourceFields["clusterId"]
			clusterField.Type = "reference[cluster]"
			schema.ResourceFields["clusterId"] = clusterField
			schema.ResourceActions["cordon"] = types.Action{}
			schema.ResourceActions["uncordon"] = types.Action{}
			schema.ResourceActions["stopDrain"] = types.Action{}
			schema.ResourceActions["drain"] = types.Action{
				Input: "nodeDrainInput",
			}
		}, struct {
			PublicEndpoints string `json:"publicEndpoints" norman:"type=array[publicEndpoint],nocreate,noupdate"`
		}{}).
		MustImportAndCustomize(&Version, v3.NodeDriver{}, func(schema *types.Schema) {
			schema.ResourceActions["activate"] = types.Action{
				Output: "nodeDriver",
			}
			schema.ResourceActions["deactivate"] = types.Action{
				Output: "nodeDriver",
			}
		}).
		MustImportAndCustomize(&Version, v3.NodeTemplate{}, func(schema *types.Schema) {
			delete(schema.ResourceFields, "namespaceId")
		})
}

func tokens(schemas *types.Schemas) *types.Schemas {
	return schemas.
		MustImportAndCustomize(&Version, v3.Token{}, func(schema *types.Schema) {
			schema.CollectionActions = map[string]types.Action{
				"logout": {},
			}
		})
}

func authnTypes(schemas *types.Schemas) *types.Schemas {
	return schemas.
		AddMapperForType(&Version, v3.User{}, m.DisplayName{},
			&m.Embed{Field: "status"}).
		AddMapperForType(&Version, v3.Group{}, m.DisplayName{}).
		MustImport(&Version, v3.Group{}).
		MustImport(&Version, v3.GroupMember{}).
		AddMapperForType(&Version, v3.Principal{}, m.DisplayName{}).
		MustImportAndCustomize(&Version, v3.Principal{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}
			schema.CollectionActions = map[string]types.Action{
				"search": {
					Input:  "searchPrincipalsInput",
					Output: "collection",
				},
			}
		}).
		MustImport(&Version, v3.SearchPrincipalsInput{}).
		MustImport(&Version, v3.ChangePasswordInput{}).
		MustImport(&Version, v3.SetPasswordInput{}).
		MustImportAndCustomize(&Version, v3.User{}, func(schema *types.Schema) {
			schema.ResourceActions = map[string]types.Action{
				"setpassword": {
					Input:  "setPasswordInput",
					Output: "user",
				},
			}
			schema.CollectionActions = map[string]types.Action{
				"changepassword": {
					Input: "changePasswordInput",
				},
			}
		}).
		MustImportAndCustomize(&Version, v3.AuthConfig{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
		}).
		// Local Config
		MustImportAndCustomize(&Version, v3.LocalConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet}
		}).
		//Github Config
		MustImportAndCustomize(&Version, v3.GithubConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.ResourceActions = map[string]types.Action{
				"disable": {},
				"configureTest": {
					Input:  "githubConfig",
					Output: "githubConfigTestOutput",
				},
				"testAndApply": {
					Input: "githubConfigApplyInput",
				},
			}
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
		}).
		MustImport(&Version, v3.GithubConfigTestOutput{}).
		MustImport(&Version, v3.GithubConfigApplyInput{}).
		//AzureAD Config
		MustImportAndCustomize(&Version, v3.AzureADConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.ResourceActions = map[string]types.Action{
				"disable": {},
				"configureTest": {
					Input:  "azureADConfig",
					Output: "azureADConfigTestOutput",
				},
				"testAndApply": {
					Input: "azureADConfigApplyInput",
				},
			}
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
		}).
		MustImport(&Version, v3.AzureADConfigTestOutput{}).
		MustImport(&Version, v3.AzureADConfigApplyInput{}).
		// Active Directory Config
		MustImportAndCustomize(&Version, v3.ActiveDirectoryConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.ResourceActions = map[string]types.Action{
				"disable": {},
				"testAndApply": {
					Input: "activeDirectoryTestAndApplyInput",
				},
			}
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
		}).
		MustImport(&Version, v3.ActiveDirectoryTestAndApplyInput{}).
		// OpenLdap Config
		MustImportAndCustomize(&Version, v3.OpenLdapConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.ResourceActions = map[string]types.Action{
				"disable": {},
				"testAndApply": {
					Input: "openLdapTestAndApplyInput",
				},
			}
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
		}).
		MustImport(&Version, v3.OpenLdapTestAndApplyInput{}).
		// FreeIpa Config
		AddMapperForType(&Version, v3.FreeIpaConfig{}, m.Drop{Field: "nestedGroupMembershipEnabled"}).
		MustImportAndCustomize(&Version, v3.FreeIpaConfig{}, func(schema *types.Schema) {
			schema.BaseType = "authConfig"
			schema.ResourceActions = map[string]types.Action{
				"disable": {},
				"testAndApply": {
					Input: "freeIpaTestAndApplyInput",
				},
			}
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
			schema.MustCustomizeField("groupObjectClass", func(f types.Field) types.Field {
				f.Default = "groupofnames"
				return f
			})
			schema.MustCustomizeField("userNameAttribute", func(f types.Field) types.Field {
				f.Default = "givenName"
				return f
			})
			schema.MustCustomizeField("userObjectClass", func(f types.Field) types.Field {
				f.Default = "inetorgperson"
				return f
			})
			schema.MustCustomizeField("groupDNAttribute", func(f types.Field) types.Field {
				f.Default = "entrydn"
				return f
			})
			schema.MustCustomizeField("groupMemberUserAttribute", func(f types.Field) types.Field {
				f.Default = "entrydn"
				return f
			})
		}).
		MustImport(&Version, v3.FreeIpaTestAndApplyInput{}).
		// Saml Config
		// Ping-Saml Config
		// KeyCloak-Saml Configs
		MustImportAndCustomize(&Version, v3.PingConfig{}, configSchema).
		MustImportAndCustomize(&Version, v3.ADFSConfig{}, configSchema).
		MustImportAndCustomize(&Version, v3.KeyCloakConfig{}, configSchema).
		MustImport(&Version, v3.SamlConfigTestInput{}).
		MustImport(&Version, v3.SamlConfigTestOutput{})
}

func configSchema(schema *types.Schema) {
	schema.BaseType = "authConfig"
	schema.ResourceActions = map[string]types.Action{
		"disable": {},
		"testAndEnable": {
			Input:  "samlConfigTestInput",
			Output: "samlConfigTestOutput",
		},
	}
	schema.CollectionMethods = []string{}
	schema.ResourceMethods = []string{http.MethodGet, http.MethodPut}
}

func userTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		MustImportAndCustomize(&Version, v3.Preference{}, func(schema *types.Schema) {
			schema.MustCustomizeField("name", func(f types.Field) types.Field {
				f.Required = true
				return f
			})
			schema.MustCustomizeField("namespaceId", func(f types.Field) types.Field {
				f.Required = false
				return f
			})
		}).
		MustImportAndCustomize(&Version, v3.UserAttribute{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{}
			schema.ResourceMethods = []string{}
		})
}

func projectNetworkPolicyTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		MustImportAndCustomize(&Version, v3.ProjectNetworkPolicy{}, func(schema *types.Schema) {
			schema.CollectionMethods = []string{http.MethodGet}
			schema.ResourceMethods = []string{http.MethodGet}
		})
}

func logTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		AddMapperForType(&Version, v3.ClusterLogging{},
			&m.Embed{Field: "status"},
			m.DisplayName{}).
		AddMapperForType(&Version, v3.ProjectLogging{},
			m.DisplayName{}).
		MustImport(&Version, v3.ClusterLogging{}).
		MustImport(&Version, v3.ProjectLogging{})
}

func globalTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		AddMapperForType(&Version, v3.ListenConfig{},
			m.DisplayName{},
			m.Drop{Field: "caKey"},
			m.Drop{Field: "caCert"},
		).
		MustImport(&Version, v3.ListenConfig{}).
		MustImportAndCustomize(&Version, v3.Setting{}, func(schema *types.Schema) {
			schema.MustCustomizeField("name", func(f types.Field) types.Field {
				f.Required = true
				return f
			})
		})
}

func alertTypes(schema *types.Schemas) *types.Schemas {
	return schema.
		AddMapperForType(&Version, v3.Notifier{},
			m.DisplayName{}).
		AddMapperForType(&Version, v3.ClusterAlert{},
			&m.Embed{Field: "status"},
			m.DisplayName{}).
		AddMapperForType(&Version, v3.ProjectAlert{},
			&m.Embed{Field: "status"},
			m.DisplayName{}).
		MustImport(&Version, v3.Notification{}).
		MustImportAndCustomize(&Version, v3.Notifier{}, func(schema *types.Schema) {
			schema.CollectionActions = map[string]types.Action{
				"send": {
					Input: "notification",
				},
			}
			schema.ResourceActions = map[string]types.Action{
				"send": {
					Input: "notification",
				},
			}
		}).
		MustImportAndCustomize(&Version, v3.ClusterAlert{}, func(schema *types.Schema) {

			schema.ResourceActions = map[string]types.Action{
				"activate":   {},
				"deactivate": {},
				"mute":       {},
				"unmute":     {},
			}
		}).
		MustImportAndCustomize(&Version, v3.ProjectAlert{}, func(schema *types.Schema) {

			schema.ResourceActions = map[string]types.Action{
				"activate":   {},
				"deactivate": {},
				"mute":       {},
				"unmute":     {},
			}
		})

}

func composeType(schemas *types.Schemas) *types.Schemas {
	return schemas.MustImport(&Version, v3.ComposeConfig{})
}
