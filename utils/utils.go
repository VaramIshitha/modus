/*
 * Copyright 2024 Hypermode, Inc.
 */

package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func MapValues[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

func ParseNameAndVersion(s string) (name string, version string) {
	i := strings.LastIndex(s, "@")
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func ConvertToMap(data any) (map[string]any, error) {
	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(j))
	dec.UseNumber()

	var m map[string]any
	err = dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func ConvertToStruct[T any](data map[string]any) (T, error) {
	var result T

	j, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(j, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func EnvVarFlagEnabled(envVarName string) bool {
	v := os.Getenv(envVarName)
	b, err := strconv.ParseBool(v)
	return err == nil && b
}

func HypermodeDebugEnabled() bool {
	return EnvVarFlagEnabled("HYPERMODE_DEBUG")
}

func HypermodeTraceEnabled() bool {
	return EnvVarFlagEnabled("HYPERMODE_TRACE")
}