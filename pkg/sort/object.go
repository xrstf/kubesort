// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package sort

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.xrstf.de/kubesort/pkg/jsonpath"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type SortingRule struct {
	Kinds        []string `yaml:"kinds,omitempty"`
	Path         string   `yaml:"path"`
	ByKey        string   `yaml:"byKey,omitempty"`
	ByValue      *bool    `yaml:"byValue,omitempty"`
	RBACRules    *bool    `yaml:"rbacRules,omitempty"`
	RBACSubjects *bool    `yaml:"rbacSubjects,omitempty"`
}

func (r SortingRule) Validate() error {
	var methods []string
	if r.ByKey != "" {
		methods = append(methods, "byKey")
	}
	if r.ByValue != nil {
		methods = append(methods, "byValue")
	}
	if r.RBACRules != nil {
		methods = append(methods, "rbacRules")
	}
	if r.RBACSubjects != nil {
		methods = append(methods, "rbacSubjects")
	}

	switch len(methods) {
	case 0:
		return errors.New("no sorting method specified")
	case 1:
		return nil
	default:
		return fmt.Errorf("cannot specify multiple sorting methods: %v", methods)
	}
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

		return sortSlice(list, rule)
	})
	if err != nil {
		return nil, err
	}

	return patched.(map[string]any), nil
}

func sortSlice(items []any, rule SortingRule) ([]any, error) {
	if rule.ByKey != "" {
		return sortSliceByKey(items, rule.ByKey), nil
	}

	if rule.ByValue != nil && *rule.ByValue {
		return sortSliceByValue(items), nil
	}

	if rule.RBACRules != nil && *rule.RBACRules {
		return sortRBACRules(items), nil
	}

	if rule.RBACSubjects != nil && *rule.RBACSubjects {
		return sortRBACSubjects(items), nil
	}

	return nil, errors.New("no supporting sorting mechanism configured")
}

func sortSliceByValue(items []any) []any {
	slices.SortFunc(items, func(a, b any) int {
		aValue, ok := a.(string)
		if !ok {
			return -1
		}

		bValue, ok := b.(string)
		if !ok {
			return 1
		}

		return strings.Compare(aValue, bValue)
	})

	return items
}

func sortSliceByKey(items []any, keyField string) []any {
	slices.SortFunc(items, func(a, b any) int {
		aKey, ok := getField(a, keyField)
		if !ok {
			return -1
		}

		bKey, ok := getField(b, keyField)
		if !ok {
			return 1
		}

		return strings.Compare(aKey, bKey)
	})

	return items
}

func getField(val any, fieldName string) (string, bool) {
	asMap, ok := val.(map[string]any)
	if !ok {
		return "", false
	}

	value, ok := asMap[fieldName]
	if !ok {
		return "", false
	}

	asString, ok := value.(string)
	if !ok {
		return "", false
	}

	return asString, true
}

func sortRBACRules(rules []any) []any {
	slices.SortStableFunc(rules, func(a, b any) int {
		aRule := rbacv1.PolicyRule{}
		if marshalAs(a, &aRule) != nil {
			return -1
		}

		bRule := rbacv1.PolicyRule{}
		if marshalAs(b, &bRule) != nil {
			return 1
		}

		aHasCore := slices.Contains(aRule.APIGroups, "")
		bHasCore := slices.Contains(bRule.APIGroups, "")

		if aHasCore != bHasCore {
			if aHasCore {
				return -1
			} else {
				return 1
			}
		}

		if diff := compareStringSlices(aRule.APIGroups, bRule.APIGroups); diff != 0 {
			return diff
		}

		if diff := compareStringSlices(aRule.Resources, bRule.Resources); diff != 0 {
			return diff
		}

		if diff := compareStringSlices(aRule.ResourceNames, bRule.ResourceNames); diff != 0 {
			return diff
		}

		if diff := compareStringSlices(aRule.NonResourceURLs, bRule.NonResourceURLs); diff != 0 {
			return diff
		}

		if diff := compareStringSlices(aRule.Verbs, bRule.Verbs); diff != 0 {
			return diff
		}

		return 0
	})

	return rules
}

func compareStringSlices(a, b []string) int {
	return strings.Compare(strings.Join(a, ";"), strings.Join(b, ";"))
}

func sortRBACSubjects(subjects []any) []any {
	slices.SortFunc(subjects, func(a, b any) int {
		aSubject := rbacv1.Subject{}
		if marshalAs(a, &aSubject) != nil {
			return -1
		}

		bSubject := rbacv1.Subject{}
		if marshalAs(b, &bSubject) != nil {
			return 1
		}

		if aSubject.Kind != bSubject.Kind {
			return strings.Compare(aSubject.Kind, bSubject.Kind)
		}

		if aSubject.Namespace != bSubject.Namespace {
			return strings.Compare(aSubject.Namespace, bSubject.Namespace)
		}

		return strings.Compare(aSubject.Name, bSubject.Name)
	})

	return subjects
}

func marshalAs(data any, dest any) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	if err := json.NewDecoder(&buf).Decode(dest); err != nil {
		return err
	}

	return nil
}
