// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/spf13/pflag"

	"go.xrstf.de/kubesort/pkg/sort"
	"go.xrstf.de/kubesort/pkg/types"
	"go.xrstf.de/kubesort/pkg/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kruntime "k8s.io/apimachinery/pkg/runtime"
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
	configFile   string
}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.configFile, "config", "c", o.configFile, "Load configuration from this file")
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

	config, err := types.LoadConfig(opts.configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	allObjects := []*unstructured.Unstructured{}

	for _, arg := range args {
		objects, err := yaml.Decode(arg)
		if err != nil {
			log.Fatalf("Failed to load %q: %v", arg, err)
		}

		allObjects = append(allObjects, objects...)
	}

	if opts.flattenLists || config.FlattenLists {
		allObjects = flattenLists(allObjects)
	}

	allObjects, err = sort.Objects(allObjects, config.ObjectRules)
	if err != nil {
		log.Fatalf("Failed to sort objects: %v", err)
	}

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
