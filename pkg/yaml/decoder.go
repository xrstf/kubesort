// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// 5 MB, same as chunk size in decoder
	bufSize = 5 * 1024 * 1024
)

func Decode(source string) ([]*unstructured.Unstructured, error) {
	if source == "-" {
		// thank you https://stackoverflow.com/a/26567513
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeCharDevice != 0 {
			return nil, errors.New("no data provided on stdin")
		}

		return DecodeReader(os.Stdin)
	}

	stat, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}

	if stat.IsDir() {
		return nil, fmt.Errorf("%s is a directory", source)
	}

	return DecodeFile(source)
}

func DecodeFile(source string) ([]*unstructured.Unstructured, error) {
	f, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return DecodeReader(f)
}

func DecodeReader(source io.ReadCloser) ([]*unstructured.Unstructured, error) {
	docSplitter := yamlutil.NewDocumentDecoder(source)
	defer docSplitter.Close()

	result := []*unstructured.Unstructured{}

	for i := 1; true; i++ {
		buf := make([]byte, bufSize)
		read, err := docSplitter.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("document %d is larger than the internal buffer", i)
		}

		object, err := parseDocument(buf[:read])
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			fmt.Println(string(buf[:read]))
			return nil, fmt.Errorf("document %d is invalid: %w", i, err)
		}

		if object == nil || len(object.Object) == 0 {
			continue
		}

		result = append(result, object)
	}

	return result, nil
}

func parseDocument(data []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}

	err := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024).Decode(obj)
	if err != nil {
		return nil, fmt.Errorf("document is not valid YAML: %w", err)
	}

	return obj, nil
}
