package kubeapi

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/rotisserie/eris"
)

var (
	apiVersionRegexp *regexp.Regexp

	MalformedVersionError = func(version string) error {
		return eris.Errorf("Failed to parse kubernetes api version from %v", version)
	}

	InvalidMajorVersionError = eris.New("Major version cannot be zero")

	InvalidPrereleaseVersionError = eris.New("Prerelease version cannot be zero")
)

type PrereleaseModifier int

const (
	Alpha PrereleaseModifier = iota + 1
	Beta
	GA
)

func (m PrereleaseModifier) String() string {
	switch m {
	case Alpha:
		return "alpha"
	case Beta:
		return "beta"
	default:
		return ""
	}
}

func init() {
	apiVersionRegexp = regexp.MustCompile(`^v([0-9]+)((alpha|beta)([0-9]+))?$`)
}

// Version models the structure of Kubernetes API version tags, but does NOT handle version as documented by k8s.
// Priority is assigned by implied order of release, such that alpha/beta v2 tags are "greater than" the GA v1 tag.
// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definition-versioning/#version-priority
type Version interface {
	Major() int
	Prerelease() int
	PrereleaseModifier() PrereleaseModifier
	String() string
	GreaterThan(other Version) bool
	LessThan(other Version) bool
	Equal(other Version) bool
}

type version struct {
	major, prerelease int
	modifier          PrereleaseModifier
}

func ParseVersion(v string) (Version, error) {
	matches := apiVersionRegexp.FindStringSubmatch(v)
	if matches == nil {
		return nil, MalformedVersionError(v)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}
	if major == 0 {
		return nil, InvalidMajorVersionError
	}

	// Prerelease v info is optional
	var prerelease int
	if matches[3] != "" && matches[4] != "" {
		prerelease, err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, err
		}
		if prerelease == 0 {
			return nil, InvalidPrereleaseVersionError
		}
	}

	var modifier PrereleaseModifier
	switch matches[3] {
	case "alpha":
		modifier = Alpha
	case "beta":
		modifier = Beta
	default:
		modifier = GA
	}

	return &version{
		major:      major,
		prerelease: prerelease,
		modifier:   modifier,
	}, nil
}

func (v *version) Major() int {
	return v.major
}

func (v *version) Prerelease() int {
	return v.prerelease
}

func (v *version) PrereleaseModifier() PrereleaseModifier {
	return v.modifier
}

func (v *version) String() string {
	sb := strings.Builder{}
	sb.WriteString("v")
	sb.WriteString(strconv.Itoa(v.Major()))

	switch v.PrereleaseModifier() {
	case Alpha:
		sb.WriteString("alpha")
		sb.WriteString(strconv.Itoa(v.Prerelease()))
	case Beta:
		sb.WriteString("beta")
		sb.WriteString(strconv.Itoa(v.Prerelease()))
	}

	return sb.String()
}

func (v *version) GreaterThan(other Version) bool {
	if v.Major() < other.Major() {
		return false
	}

	if v.Major() == other.Major() {
		if v.PrereleaseModifier() < other.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == other.PrereleaseModifier() {
			if v.Prerelease() <= other.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *version) LessThan(other Version) bool {
	if v.Major() > other.Major() {
		return false
	}

	if v.Major() == other.Major() {
		if v.PrereleaseModifier() > other.PrereleaseModifier() {
			return false
		}

		if v.PrereleaseModifier() == other.PrereleaseModifier() {
			if v.Prerelease() >= other.Prerelease() {
				return false
			}
		}
	}

	return true
}

func (v *version) Equal(other Version) bool {
	return v.String() == other.String()
}
