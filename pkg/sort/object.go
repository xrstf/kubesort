// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package sort

import (
	"regexp"
	"slices"
	"strings"

	"go.xrstf.de/kubesort/pkg/jsonpath"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type SortingRule struct {
	Kinds []string `yaml:"kinds"`
	Path  string   `yaml:"path"`
	ByKey string   `yaml:"byKey"`
}

func (r SortingRule) JSONPath() jsonpath.Path {
	path := jsonpath.Path{}

	parts := strings.Split(r.Path, ".")
	for _, part := range parts {
		if strings.HasSuffix(part, "[]") {
			part = strings.TrimSuffix(part, "[]")
			path = append(path, jsonpath.KeyStep(part), wildcardStep{})
		} else {
			path = append(path, jsonpath.KeyStep(part))
		}
	}

	return path
}

func (r SortingRule) Matches(obj *unstructured.Unstructured) bool {
	if len(r.Kinds) == 0 {
		return true
	}

	return slices.Contains(r.Kinds, obj.GetKind())
}

var pathRegex = regexp.MustCompile(`^(\.[^.]+)+$`)

type wildcardStep struct{}

func (wildcardStep) Keep(key any, value any) (bool, error) {
	return true, nil
}

func Object(obj *unstructured.Unstructured, rules []SortingRule) (*unstructured.Unstructured, error) {
	data := obj.Object

	for _, rule := range rules {
		if !rule.Matches(obj) {
			continue
		}

		patched, err := applyRule(data, rule)
		if err != nil {
			return nil, err
		}

		data = patched
	}

	obj.Object = data

	return obj, nil
}

func applyRule(obj map[string]any, rule SortingRule) (map[string]any, error) {
	patched, err := jsonpath.Patch(obj, rule.JSONPath(), func(exists bool, key, val any) (any, error) {
		if !exists {
			return nil, nil
		}

		list, ok := val.([]any)
		if !ok {
			return val, nil
		}

		return sortSliceByKey(list, rule.ByKey), nil
	})
	if err != nil {
		return nil, err
	}

	return patched.(map[string]any), nil
}

func sortSliceByKey(items []any, keyField string) []any {
	slices.SortFunc(items, func(a, b any) int {
		aMap, ok := a.(map[string]any)
		if !ok {
			return -1
		}

		bMap, ok := b.(map[string]any)
		if !ok {
			return 1
		}

		aKey := aMap[keyField].(string)
		bKey := bMap[keyField].(string)

		return strings.Compare(aKey, bKey)
	})

	return items
}
