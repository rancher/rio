package lookup

import (
	"regexp"

	"github.com/rancher/rio/cli/pkg/types"
	namer "github.com/rancher/rio/pkg/name"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	dns1035                      string = "[a-z]([-a-z0-9]*[a-z0-9])?"
	FourPartsNameType                   = NameType("fourParts")
	FullDomainNameTypeNameType          = NameType("domainName")
	SingleNameNameType                  = NameType("singleName")
	StackScopedNameType                 = NameType("stackScoped")
	ThreePartsNameType                  = NameType("threeParts")
	VersionedSingleNameNameType         = NameType("versionedSingleName")
	VersionedStackScopedNameType        = NameType("versionedStackScoped")
)

type NameType string

var (
	nameTypes = map[NameType]nameType{
		FullDomainNameTypeNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "\\." + dns1035 + ".*" + "$"),
			lookup: resolveFullDomain,
		},
		SingleNameNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "$"),
			lookup: resolveSingleName,
		},
		VersionedSingleNameNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + ":" + dns1035 + "$"),
			lookup: resolveStackScoped,
		},
		StackScopedNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "/" + dns1035 + "$"),
			lookup: resolveStackScoped,
		},
		VersionedStackScopedNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "/" + dns1035 + ":" + dns1035 + "$"),
			lookup: resolveStackScoped,
		},
		ThreePartsNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "/" + dns1035 + "/" + dns1035 + "$"),
			lookup: resolvePod,
		},
		FourPartsNameType: {
			Regexp: regexp.MustCompile("^" + dns1035 + "/" + dns1035 + "/" + dns1035 + "/" + dns1035 + "$"),
			lookup: resolvePod,
		},
	}
)

type nameType struct {
	types  map[string]bool
	Regexp *regexp.Regexp
	lookup func(defaultStackName, name, typeName string) types.Resource
}

func (n nameType) Lookup(lookup ClientLookup, name, typeName string) (types.Resource, error) {
	if !n.types[typeName] {
		return types.Resource{}, errors.NewNotFound(schema.GroupResource{}, name)
	}
	r := n.lookup(lookup.GetDefaultStackName(), name, typeName)
	r, err := lookup.ByID(r.Namespace, r.Name, typeName)
	r.LookupName = name
	return r, err
}

func (n nameType) Matches(name string) bool {
	return n.Regexp.MatchString(name)
}

func RegisterType(typeName string, supportedNameTypes ...NameType) {
	for _, nameType := range supportedNameTypes {
		if nameTypes[nameType].types == nil {
			t := nameTypes[nameType]
			t.types = map[string]bool{}
			nameTypes[nameType] = t
		}
		nameTypes[nameType].types[typeName] = true
	}
}

func resolveFullDomain(defaultStackName, name, typeName string) types.Resource {
	return types.Resource{
		Namespace: defaultStackName,
		Name:      namer.PublicDomain(name),
		Type:      typeName,
	}
}

func resolveSingleName(defaultStackName, name, typeName string) types.Resource {
	return types.Resource{
		Namespace: defaultStackName,
		Name:      name,
		Type:      typeName,
	}
}

func resolveStackScoped(defaultStackName, name, typeName string) types.Resource {
	stackScoped := ParseStackScoped(defaultStackName, name)
	return types.Resource{
		Namespace: stackScoped.StackName,
		Name:      stackScoped.ResourceName,
		Type:      typeName,
	}
}

func resolvePod(defaultStackName, name, typeName string) types.Resource {
	container, _ := ParseContainer(defaultStackName, name)
	return types.Resource{
		Namespace: container.Service.StackName,
		Name:      container.K8sPodName(),
		Type:      typeName,
	}
}
