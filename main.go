// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/pflag"

	"go.xrstf.de/kubesort/pkg/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

func printVersion() {
	// handle empty values in case `go install` was used
	if BuildCommit == "" {
		fmt.Printf("kubesort dev, built with %s\n",
			runtime.Version(),
		)
	} else {
		fmt.Printf("kubesort %s (%s), built with %s on %s\n",
			BuildTag,
			BuildCommit[:10],
			runtime.Version(),
			BuildDate,
		)
	}
}

type globalOptions struct {
	flattenLists bool
	version      bool
}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.flattenLists, "flatten", "f", o.flattenLists, "Unwrap List kinds into standalone objects")
	fs.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately")
}

func main() {
	opts := globalOptions{}

	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if opts.version {
		printVersion()
		return
	}

	args := pflag.Args()
	if len(args) == 0 {
		log.Fatal("No input file(s) provided.")
	}

	allObjects := []*unstructured.Unstructured{}

	for _, arg := range args {
		objects, err := yaml.Decode(arg)
		if err != nil {
			log.Fatalf("Failed to load %q: %v", arg, err)
		}

		allObjects = append(allObjects, objects...)
	}

	if opts.flattenLists {
		allObjects = flattenLists(allObjects)
	}

	allObjects = sortObjects(allObjects)

	for _, obj := range allObjects {
		encoded, err := yaml.Encode(obj)
		if err != nil {
			log.Fatalf("Failed to encode object: %v", err)
		}

		fmt.Printf("---\n%s\n", string(encoded))
	}
}

func flattenLists(input []*unstructured.Unstructured) []*unstructured.Unstructured {
	result := []*unstructured.Unstructured{}

	for i, obj := range input {
		if isList(obj) {
			if err := obj.EachListItem(func(o kruntime.Object) error {
				result = append(result, o.(*unstructured.Unstructured))
				return nil
			}); err != nil {
				log.Fatalf("Failed to flatten list: %v", err)
			}
		} else {
			result = append(result, input[i])
		}
	}

	return result
}

func isList(obj *unstructured.Unstructured) bool {
	if !strings.HasSuffix(obj.GetKind(), "List") {
		return false
	}

	_, ok := obj.Object["items"]

	return ok
}

func sortObjects(objects []*unstructured.Unstructured) []*unstructured.Unstructured {
	slices.SortStableFunc(objects, func(a, b *unstructured.Unstructured) int {
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

	return objects
}

func isCRD(obj *unstructured.Unstructured) bool {
	return obj.GetKind() == "CustomResourceDefinition"
}

func isClusterScoped(obj *unstructured.Unstructured) bool {
	return obj.GetNamespace() == ""
}
