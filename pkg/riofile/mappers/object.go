package mappers

import (
	"github.com/rancher/mapper"
	"github.com/rancher/mapper/mappers"
)

func NewObject(kind, apiVersion string) mapper.Mapper {
	return mapper.Mappers{
		mappers.SetValue{Field: "kind", InternalValue: kind},
		mappers.SetValue{Field: "apiVersion", InternalValue: apiVersion},
		mappers.Drop{
			Field: "kind",
		},
		mappers.Drop{
			Field: "apiVersion",
		},
		mappers.Move{
			From: "metadata/labels",
			To:   "labels",
		},
		mappers.Move{
			From: "metadata/annotations",
			To:   "annotations",
		},
		mappers.Drop{
			Field: "metadata",
		},
		mappers.Drop{
			IgnoreDefinition: true,
			Field:            "status",
		},
		&mappers.Embed{
			Field:    "spec",
			Optional: true,
		},
		LabelCleaner{},
	}
}
