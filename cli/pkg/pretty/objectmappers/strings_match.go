package objectmappers

import (
	"strings"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewStringMatch(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &StringMatchStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			return ParseStringMatch(str), nil
		},
	}
}

type StringMatchStringer struct {
	v1.StringMatch
}

func (s StringMatchStringer) MaybeString() interface{} {
	switch {
	case s.Exact != "":
		return s.Exact
	case s.Prefix != "":
		return s.Prefix + "*"
	case s.Regexp != "":
		return "regex(" + s.Regexp + ")"
	default:
		return ""
	}
}

func ParseStringMatch(str string) *v1.StringMatch {
	if strings.HasSuffix(str, "*") {
		return &v1.StringMatch{
			Prefix: str[:len(str)-1],
		}
	} else if IsRegexp(str) {
		return &v1.StringMatch{
			Regexp: strings.TrimSuffix(strings.SplitN(str, "(", 2)[1], ")"),
		}
	}

	return &v1.StringMatch{
		Exact: str,
	}
}

func IsRegexp(str string) bool {
	return (strings.HasPrefix(str, "regex(") || strings.HasPrefix(str, "regexp(")) && strings.HasSuffix(str, ")")
}

func stringMatchToString(v interface{}) string {
	m := convert.ToMapInterface(v)

	exact, _ := m["exact"].(string)
	prefix, _ := m["prefix"].(string)
	regexp, _ := m["regexp"].(string)

	switch {
	case exact != "":
		return exact
	case prefix != "":
		return prefix + "*"
	case regexp != "":
		return "regex(" + regexp + ")"
	default:
		return ""
	}
}
