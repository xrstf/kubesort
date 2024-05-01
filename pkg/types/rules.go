package types

import (
	"go.xrstf.de/kubesort/pkg/sort"
)

var (
	defaultObjectRules = []sort.SortingRule{
		{
			Path:  "spec.template.spec.containers",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.containers[].env",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.containers[].volumeMounts",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.containers[].ports",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.initContainers[].env",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.initContainers[].volumeMounts",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.initContainers[].ports",
			ByKey: "name",
		},
		{
			Path:  "spec.template.spec.volumes",
			ByKey: "name",
		},
	}
)
