package v3

import (
	"github.com/rancher/norman/condition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const UserConditionInitialRolesPopulated condition.Cond = "InitialRolesPopulated"

type Token struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Token           string            `json:"token" norman:"writeOnly,noupdate"`
	UserPrincipal   Principal         `json:"userPrincipal" norman:"type=reference[principal]"`
	GroupPrincipals []Principal       `json:"groupPrincipals" norman:"type=array[reference[principal]]"`
	ProviderInfo    map[string]string `json:"providerInfo,omitempty"`
	UserID          string            `json:"userId" norman:"type=reference[user]"`
	AuthProvider    string            `json:"authProvider"`
	TTLMillis       int64             `json:"ttl"`
	LastUpdateTime  string            `json:"lastUpdateTime"`
	IsDerived       bool              `json:"isDerived"`
	Description     string            `json:"description"`
	Expired         bool              `json:"expired"`
	ExpiresAt       string            `json:"expiresAt"`
	Current         bool              `json:"current"`
}

type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DisplayName        string     `json:"displayName,omitempty"`
	Description        string     `json:"description"`
	Username           string     `json:"username,omitempty"`
	Password           string     `json:"password,omitempty" norman:"writeOnly,noupdate"`
	MustChangePassword bool       `json:"mustChangePassword,omitempty"`
	PrincipalIDs       []string   `json:"principalIds,omitempty" norman:"type=array[reference[principal]]"`
	Me                 bool       `json:"me,omitempty"`
	Enabled            *bool      `json:"enabled,omitempty" norman:"default=true"`
	Spec               UserSpec   `json:"spec,omitempty"`
	Status             UserStatus `json:"status"`
}

type UserStatus struct {
	Conditions []UserCondition `json:"conditions"`
}

type UserCondition struct {
	// Type of user condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}

type UserSpec struct{}

// UserAttribute will have a CRD (and controller) generated for it, but will not be exposed in the API.
type UserAttribute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	UserName        string
	GroupPrincipals map[string]Principals // the value is a []Principal, but code generator cannot handle slice as a value
}

type Principals struct {
	Items []Principal
}

type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DisplayName string `json:"displayName,omitempty"`
}

type GroupMember struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	GroupName   string `json:"groupName,omitempty" norman:"type=reference[group]"`
	PrincipalID string `json:"principalId,omitempty" norman:"type=reference[principal]"`
}

type Principal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	DisplayName    string            `json:"displayName,omitempty"`
	LoginName      string            `json:"loginName,omitempty"`
	ProfilePicture string            `json:"profilePicture,omitempty"`
	ProfileURL     string            `json:"profileURL,omitempty"`
	PrincipalType  string            `json:"principalType,omitempty"`
	Me             bool              `json:"me,omitempty"`
	MemberOf       bool              `json:"memberOf,omitempty"`
	Provider       string            `json:"provider,omitempty"`
	ExtraInfo      map[string]string `json:"extraInfo,omitempty"`
}

type SearchPrincipalsInput struct {
	Name          string `json:"name" norman:"type=string,required,notnullable"`
	PrincipalType string `json:"principalType,omitempty" norman:"type=enum,options=user|group"`
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword" norman:"type=string,required"`
	NewPassword     string `json:"newPassword" norman:"type=string,required"`
}

type SetPasswordInput struct {
	NewPassword string `json:"newPassword" norman:"type=string,required"`
}

type AuthConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Type                string   `json:"type" norman:"noupdate"`
	Enabled             bool     `json:"enabled,omitempty"`
	AccessMode          string   `json:"accessMode,omitempty" norman:"required,notnullable,type=enum,options=required|restricted|unrestricted"`
	AllowedPrincipalIDs []string `json:"allowedPrincipalIds,omitempty" norman:"type=array[reference[principal]]"`
}

type LocalConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`
}

type GithubConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`

	Hostname     string `json:"hostname,omitempty" norman:"default=github.com" norman:"required"`
	TLS          bool   `json:"tls,omitempty" norman:"notnullable,default=true" norman:"required"`
	ClientID     string `json:"clientId,omitempty" norman:"required"`
	ClientSecret string `json:"clientSecret,omitempty" norman:"required,type=password"`
}

type GithubConfigTestOutput struct {
	RedirectURL string `json:"redirectUrl"`
}

type GithubConfigApplyInput struct {
	GithubConfig GithubConfig `json:"githubConfig,omitempty"`
	Code         string       `json:"code,omitempty"`
	Enabled      bool         `json:"enabled,omitempty"`
}

type AzureADConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`

	Endpoint          string `json:"endpoint,omitempty" norman:"default=https://login.microsoftonline.com/,required,notnullable"`
	GraphEndpoint     string `json:"graphEndpoint,omitempty" norman:"required,notnullable"`
	TokenEndpoint     string `json:"tokenEndpoint,omitempty" norman:"required,notnullable"`
	AuthEndpoint      string `json:"authEndpoint,omitempty" norman:"required,notnullable"`
	TenantID          string `json:"tenantId,omitempty" norman:"required,notnullable"`
	ApplicationID     string `json:"applicationId,omitempty" norman:"required,notnullable"`
	ApplicationSecret string `json:"applicationSecret,omitempty" norman:"required,notnullable,type=password"`
	RancherURL        string `json:"rancherUrl,omitempty" norman:"required,notnullable"`
}

type AzureADConfigTestOutput struct {
	RedirectURL string `json:"redirectUrl"`
}

type AzureADConfigApplyInput struct {
	Config AzureADConfig `json:"config,omitempty"`
	Code   string        `json:"code,omitempty"`
}

type ActiveDirectoryConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`

	Servers                      []string `json:"servers,omitempty"                     norman:"type=array[string],required"`
	Port                         int64    `json:"port,omitempty"                        norman:"default=389"`
	TLS                          bool     `json:"tls,omitempty"                         norman:"default=false"`
	Certificate                  string   `json:"certificate,omitempty"`
	DefaultLoginDomain           string   `json:"defaultLoginDomain,omitempty"`
	ServiceAccountUsername       string   `json:"serviceAccountUsername,omitempty"      norman:"required"`
	ServiceAccountPassword       string   `json:"serviceAccountPassword,omitempty"      norman:"type=password,required"`
	UserDisabledBitMask          int64    `json:"userDisabledBitMask,omitempty"         norman:"default=2"`
	UserSearchBase               string   `json:"userSearchBase,omitempty"              norman:"required"`
	UserSearchAttribute          string   `json:"userSearchAttribute,omitempty"         norman:"default=sAMAccountName|sn|givenName,required"`
	UserSearchFilter             string   `json:"userSearchFilter,omitempty"`
	UserLoginAttribute           string   `json:"userLoginAttribute,omitempty"          norman:"default=sAMAccountName,required"`
	UserObjectClass              string   `json:"userObjectClass,omitempty"             norman:"default=person,required"`
	UserNameAttribute            string   `json:"userNameAttribute,omitempty"           norman:"default=name,required"`
	UserEnabledAttribute         string   `json:"userEnabledAttribute,omitempty"        norman:"default=userAccountControl,required"`
	GroupSearchBase              string   `json:"groupSearchBase,omitempty"`
	GroupSearchAttribute         string   `json:"groupSearchAttribute,omitempty"        norman:"default=sAMAccountName,required"`
	GroupSearchFilter            string   `json:"groupSearchFilter,omitempty"`
	GroupObjectClass             string   `json:"groupObjectClass,omitempty"            norman:"default=group,required"`
	GroupNameAttribute           string   `json:"groupNameAttribute,omitempty"          norman:"default=name,required"`
	GroupDNAttribute             string   `json:"groupDNAttribute,omitempty"            norman:"default=distinguishedName,required"`
	GroupMemberUserAttribute     string   `json:"groupMemberUserAttribute,omitempty"    norman:"default=distinguishedName,required"`
	GroupMemberMappingAttribute  string   `json:"groupMemberMappingAttribute,omitempty" norman:"default=member,required"`
	ConnectionTimeout            int64    `json:"connectionTimeout,omitempty"           norman:"default=5000,notnullable,required"`
	NestedGroupMembershipEnabled *bool    `json:"nestedGroupMembershipEnabled,omitempty" norman:"default=false"`
}

type ActiveDirectoryTestAndApplyInput struct {
	ActiveDirectoryConfig ActiveDirectoryConfig `json:"activeDirectoryConfig,omitempty"`
	Username              string                `json:"username"`
	Password              string                `json:"password"`
	Enabled               bool                  `json:"enabled,omitempty"`
}

type LdapConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`

	Servers                         []string `json:"servers,omitempty"                     norman:"type=array[string],notnullable,required"`
	Port                            int64    `json:"port,omitempty"                        norman:"default=389,notnullable,required"`
	TLS                             bool     `json:"tls,omitempty"                         norman:"default=false,notnullable,required"`
	Certificate                     string   `json:"certificate,omitempty"`
	ServiceAccountDistinguishedName string   `json:"serviceAccountDistinguishedName,omitempty"      norman:"required"`
	ServiceAccountPassword          string   `json:"serviceAccountPassword,omitempty"      norman:"type=password,required"`
	UserDisabledBitMask             int64    `json:"userDisabledBitMask,omitempty"`
	UserSearchBase                  string   `json:"userSearchBase,omitempty"              norman:"notnullable,required"`
	UserSearchAttribute             string   `json:"userSearchAttribute,omitempty"         norman:"default=uid|sn|givenName,notnullable,required"`
	UserLoginAttribute              string   `json:"userLoginAttribute,omitempty"          norman:"default=uid,notnullable,required"`
	UserObjectClass                 string   `json:"userObjectClass,omitempty"             norman:"default=inetOrgPerson,notnullable,required"`
	UserNameAttribute               string   `json:"userNameAttribute,omitempty"           norman:"default=cn,notnullable,required"`
	UserMemberAttribute             string   `json:"userMemberAttribute,omitempty"           norman:"default=memberOf,notnullable,required"`
	UserEnabledAttribute            string   `json:"userEnabledAttribute,omitempty"`
	GroupSearchBase                 string   `json:"groupSearchBase,omitempty"`
	GroupSearchAttribute            string   `json:"groupSearchAttribute,omitempty"        norman:"default=cn,notnullable,required"`
	GroupObjectClass                string   `json:"groupObjectClass,omitempty"            norman:"default=groupOfNames,notnullable,required"`
	GroupNameAttribute              string   `json:"groupNameAttribute,omitempty"          norman:"default=cn,notnullable,required"`
	GroupDNAttribute                string   `json:"groupDNAttribute,omitempty"            norman:"default=entryDN,notnullable"`
	GroupMemberUserAttribute        string   `json:"groupMemberUserAttribute,omitempty"    norman:"default=entryDN,notnullable"`
	GroupMemberMappingAttribute     string   `json:"groupMemberMappingAttribute,omitempty" norman:"default=member,notnullable,required"`
	ConnectionTimeout               int64    `json:"connectionTimeout,omitempty"           norman:"default=5000,notnullable,required"`
	NestedGroupMembershipEnabled    bool     `json:"nestedGroupMembershipEnabled"    norman:"default=false"`
}

type LdapTestAndApplyInput struct {
	LdapConfig `json:"ldapConfig,omitempty"`
	Username   string `json:"username"`
	Password   string `json:"password" norman:"type=password,required"`
}

type OpenLdapConfig struct {
	LdapConfig `json:",inline" mapstructure:",squash"`
}

type OpenLdapTestAndApplyInput struct {
	LdapTestAndApplyInput `json:",inline" mapstructure:",squash"`
}

type FreeIpaConfig struct {
	LdapConfig `json:",inline" mapstructure:",squash"`
}

type FreeIpaTestAndApplyInput struct {
	LdapTestAndApplyInput `json:",inline" mapstructure:",squash"`
}

type SamlConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AuthConfig        `json:",inline" mapstructure:",squash"`

	IDPMetadataContent string `json:"idpMetadataContent" norman:"required"`
	SpCert             string `json:"spCert"             norman:"required"`
	SpKey              string `json:"spKey"              norman:"required"`
	GroupsField        string `json:"groupsField"        norman:"required"`
	DisplayNameField   string `json:"displayNameField"   norman:"required"`
	UserNameField      string `json:"userNameField"      norman:"required"`
	UIDField           string `json:"uidField"           norman:"required"`
	RancherAPIHost     string `json:"rancherApiHost"     norman:"required"`
}

type SamlConfigTestInput struct {
	FinalRedirectURL string `json:"finalRedirectUrl"`
}

type SamlConfigTestOutput struct {
	IdpRedirectURL string `json:"idpRedirectUrl"`
}

type PingConfig struct {
	SamlConfig `json:",inline" mapstructure:",squash"`
}

type ADFSConfig struct {
	SamlConfig `json:",inline" mapstructure:",squash"`
}

type KeyCloakConfig struct {
	SamlConfig `json:",inline" mapstructure:",squash"`
}
