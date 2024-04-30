// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package yaml

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	sigsyaml "sigs.k8s.io/yaml"
)

func Encode(obj *unstructured.Unstructured) ([]byte, error) {
	return sigsyaml.Marshal(obj)
}
