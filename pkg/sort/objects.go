// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package sort

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Objects(objects []*unstructured.Unstructured, objectRules []SortingRule) ([]*unstructured.Unstructured, error) {
	sortedObjects := make([]*unstructured.Unstructured, 0, len(objects))
	for i := range objects {
		sorted, err := Object(objects[i], objectRules)
		if err != nil {
			return nil, fmt.Errorf("failed to sort object: %w", err)
		}
		sortedObjects = append(sortedObjects, sorted)
	}

	slices.SortStableFunc(sortedObjects, func(a, b *unstructured.Unstructured) int {
		// CRDs always come first
		aCRD := isCRD(a)
		bCRD := isCRD(b)

		if aCRD != bCRD {
			if aCRD {
				return -1
			} else {
				return 1
			}
		}

		// cluster-scoped resources are next (this includes Namespaces themselves)
		aClusterScoped := isClusterScoped(a)
		bClusterScoped := isClusterScoped(b)

		if aClusterScoped != bClusterScoped {
			if aClusterScoped {
				return -1
			} else {
				return 1
			}
		}

		// next we compare GVK (split APIVersion to make sure core API groups get sorted before others (because it's ""))
		aGV, err := schema.ParseGroupVersion(a.GetAPIVersion())
		if err != nil {
			return -1
		}

		bGV, err := schema.ParseGroupVersion(b.GetAPIVersion())
		if err != nil {
			return -1
		}

		if aGV.Group != bGV.Group {
			return strings.Compare(aGV.Group, bGV.Group)
		}

		if aGV.Version != bGV.Version {
			return strings.Compare(aGV.Version, bGV.Version)
		}

		if a.GetKind() != b.GetKind() {
			return strings.Compare(a.GetKind(), b.GetKind())
		}

		// next we sort by namespace
		if a.GetNamespace() != b.GetNamespace() {
			return strings.Compare(a.GetNamespace(), b.GetNamespace())
		}

		// and finally by name
		if a.GetName() != b.GetName() {
			return strings.Compare(a.GetName(), b.GetName())
		}

		log.Print("Found two objects that are the same?")

		return 0
	})

	return sortedObjects, nil
}

func isCRD(obj *unstructured.Unstructured) bool {
	return obj.GetKind() == "CustomResourceDefinition"
}

func isClusterScoped(obj *unstructured.Unstructured) bool {
	return obj.GetNamespace() == ""
}
