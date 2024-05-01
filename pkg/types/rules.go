package types

import (
	"go.xrstf.de/kubesort/pkg/sort"
	"k8s.io/utils/ptr"
)

var (
	templatePodSpecHolders = []string{"Deployment", "DaemonSet", "StatefulSet"}

	defaultObjectRules = []sort.SortingRule{
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.containers",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.containers[].env",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.containers[].volumeMounts",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.containers[].ports",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.initContainers[].env",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.initContainers[].volumeMounts",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.initContainers[].ports",
			ByKey: "name",
		},
		{
			Kinds: templatePodSpecHolders,
			Path:  "spec.template.spec.volumes",
			ByKey: "name",
		},

		{
			Kinds:   []string{"Role", "ClusterRole"},
			Path:    "rules[].apiGroups",
			ByValue: ptr.To(true),
		},
		{
			Kinds:   []string{"Role", "ClusterRole"},
			Path:    "rules[].verbs",
			ByValue: ptr.To(true),
		},
		{
			Kinds:   []string{"Role", "ClusterRole"},
			Path:    "rules[].resources",
			ByValue: ptr.To(true),
		},
		{
			Kinds:   []string{"Role", "ClusterRole"},
			Path:    "rules[].resourceNames",
			ByValue: ptr.To(true),
		},
		{
			Kinds:   []string{"Role", "ClusterRole"},
			Path:    "rules[].nonResourceURLs",
			ByValue: ptr.To(true),
		},
		// do this one after sorting each rule, so it can generate stable sorting keys
		{
			Kinds:     []string{"Role", "ClusterRole"},
			Path:      "rules",
			RBACRules: ptr.To(true),
		},
		{
			Kinds:        []string{"RoleBinding", "ClusterRoleBinding"},
			Path:         "subjects",
			RBACSubjects: ptr.To(true),
		},
	}
)
