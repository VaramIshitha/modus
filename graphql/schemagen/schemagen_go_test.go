/*
 * Copyright 2024 Hypermode, Inc.
 */

package schemagen

import (
	"context"
	"strings"
	"testing"

	"hypruntime/languages"
	"hypruntime/manifestdata"
	"hypruntime/plugins/metadata"
	"hypruntime/utils"

	"github.com/hypermodeAI/manifest"
	"github.com/stretchr/testify/require"
)

func Test_GetGraphQLSchema_Go(t *testing.T) {

	manifest := &manifest.HypermodeManifest{
		Models: map[string]manifest.ModelInfo{},
		Hosts:  map[string]manifest.HostInfo{},
		Collections: map[string]manifest.CollectionInfo{
			"collection1": {
				SearchMethods: map[string]manifest.SearchMethodInfo{
					"search1": {
						Embedder: "myEmbedder",
					},
				},
			},
		},
	}
	manifestdata.SetManifest(manifest)

	md := metadata.NewPluginMetadata()
	md.SDK = "functions-go"

	md.FnExports.AddFunction("add").
		WithParameter("a", "int32").
		WithParameter("b", "int32").
		WithResult("int32")

	md.FnExports.AddFunction("sayHello").
		WithParameter("name", "string").
		WithResult("string")

	md.FnExports.AddFunction("currentTime").
		WithResult("time.Time")

	md.FnExports.AddFunction("transform").
		WithParameter("items", "map[string]string").
		WithResult("map[string]string")

	md.FnExports.AddFunction("testDefaultIntParams").
		WithParameter("a", "int32").
		WithParameter("b", "int32", 0).
		WithParameter("c", "int32", 1)

	md.FnExports.AddFunction("testDefaultStringParams").
		WithParameter("a", "string").
		WithParameter("b", "string", "").
		WithParameter("c", "string", `a"b`).
		WithParameter("d", "*string").
		WithParameter("e", "*string", nil).
		WithParameter("f", "*string", "").
		WithParameter("g", "*string", "test")

	md.FnExports.AddFunction("testDefaultArrayParams").
		WithParameter("a", "[]int32").
		WithParameter("b", "[]int32", []int32{}).
		WithParameter("c", "[]int32", []int32{1, 2, 3}).
		WithParameter("d", "*[]int32").
		WithParameter("e", "*[]int32", nil).
		WithParameter("f", "*[]int32", []int32{}).
		WithParameter("g", "*[]int32", []int32{1, 2, 3})

	md.FnExports.AddFunction("getPerson").
		WithResult("testdata.Person")

	md.FnExports.AddFunction("getPeople").
		WithResult("[]testdata.Person")

	md.FnExports.AddFunction("addPerson").
		WithParameter("person", "testdata.Person")

	md.FnExports.AddFunction("getProductMap").
		WithResult("map[string]testdata.Product")

	md.FnExports.AddFunction("doNothing")

	md.FnExports.AddFunction("testPointers").
		WithParameter("a", "*int32").
		WithParameter("b", "[]*int32").
		WithParameter("c", "*[]int32").
		WithParameter("d", "[]*testdata.Person").
		WithParameter("e", "*[]testdata.Person").
		WithResult("*testdata.Person")
	md.Types.AddType("*int32")
	md.Types.AddType("[]*int32")
	md.Types.AddType("*[]int32")
	md.Types.AddType("[]*testdata.Person")
	md.Types.AddType("*[]testdata.Person")
	md.Types.AddType("*testdata.Person")

	// This should be excluded from the final schema
	md.FnExports.AddFunction("myEmbedder").
		WithParameter("text", "string").
		WithResult("[]float64")

	// Generated input object from the output object
	md.FnExports.AddFunction("testObj1").
		WithParameter("obj", "testdata.Obj1").
		WithResult("testdata.Obj1")
	md.Types.AddType("testdata.Obj1").
		WithField("id", "int32").
		WithField("name", "string")

	// Separate input and output objects defined
	md.FnExports.AddFunction("testObj2").
		WithParameter("obj", "testdata.Obj2Input").
		WithResult("testdata.Obj2")
	md.Types.AddType("testdata.Obj2").
		WithField("id", "int32").
		WithField("name", "string")
	md.Types.AddType("testdata.Obj2Input").
		WithField("name", "string")

	// Generated input object without output object
	md.FnExports.AddFunction("testObj3").
		WithParameter("obj", "testdata.Obj3")
	md.Types.AddType("testdata.Obj3").
		WithField("name", "string")

	// Single input object defined without output object
	md.FnExports.AddFunction("testObj4").
		WithParameter("obj", "testdata.Obj4Input")
	md.Types.AddType("testdata.Obj4Input").
		WithField("name", "string")

	md.Types.AddType("[]int32")
	md.Types.AddType("[]float64")
	md.Types.AddType("[]testdata.Person")
	md.Types.AddType("map[string]string")
	md.Types.AddType("map[string]testdata.Product")

	md.Types.AddType("testdata.Company").
		WithField("name", "string")

	md.Types.AddType("testdata.Product").
		WithField("name", "string").
		WithField("price", "float64").
		WithField("manufacturer", "testdata.Company").
		WithField("components", "[]testdata.Product")

	md.Types.AddType("testdata.Person").
		WithField("name", "string").
		WithField("age", "int32").
		WithField("addresses", "[]testdata.Address")

	md.Types.AddType("testdata.Address").
		WithField("street", "string").
		WithField("city", "string").
		WithField("state", "string").
		WithField("country", "string").
		WithField("postalCode", "string").
		WithField("location", "testdata.Coordinates")

	md.Types.AddType("testdata.Coordinates").
		WithField("lat", "float64").
		WithField("lon", "float64")

	// This should be excluded from the final schema
	md.Types.AddType("testdata.Header").
		WithField("name", "string").
		WithField("values", "[]string")

	result, err := GetGraphQLSchema(context.Background(), md)

	t.Log(result.Schema)

	expectedSchema := `
# Hypermode GraphQL Schema (auto-generated)

type Query {
  add(a: Int!, b: Int!): Int!
  addPerson(person: PersonInput!): Void
  currentTime: Timestamp!
  doNothing: Void
  getPeople: [Person!]
  getPerson: Person!
  getProductMap: [StringProductPair!]
  sayHello(name: String!): String!
  testDefaultArrayParams(a: [Int!], b: [Int!] = [], c: [Int!] = [1,2,3], d: [Int!], e: [Int!] = null, f: [Int!] = [], g: [Int!] = [1,2,3]): Void
  testDefaultIntParams(a: Int!, b: Int! = 0, c: Int! = 1): Void
  testDefaultStringParams(a: String!, b: String! = "", c: String! = "a\"b", d: String, e: String = null, f: String = "", g: String = "test"): Void
  testObj1(obj: Obj1Input!): Obj1!
  testObj2(obj: Obj2Input!): Obj2!
  testObj3(obj: Obj3Input!): Void
  testObj4(obj: Obj4Input!): Void
  testPointers(a: Int, b: [Int], c: [Int!], d: [PersonInput], e: [PersonInput!]): Person
  transform(items: [StringStringPairInput!]): [StringStringPair!]
}

scalar Timestamp
scalar Void

input AddressInput {
  street: String!
  city: String!
  state: String!
  country: String!
  postalCode: String!
  location: CoordinatesInput!
}

input CoordinatesInput {
  lat: Float!
  lon: Float!
}

input Obj1Input {
  id: Int!
  name: String!
}

input Obj2Input {
  name: String!
}

input Obj3Input {
  name: String!
}

input Obj4Input {
  name: String!
}

input PersonInput {
  name: String!
  age: Int!
  addresses: [AddressInput!]
}

input StringStringPairInput {
  key: String!
  value: String!
}

type Address {
  street: String!
  city: String!
  state: String!
  country: String!
  postalCode: String!
  location: Coordinates!
}

type Company {
  name: String!
}

type Coordinates {
  lat: Float!
  lon: Float!
}

type Obj1 {
  id: Int!
  name: String!
}

type Obj2 {
  id: Int!
  name: String!
}

type Person {
  name: String!
  age: Int!
  addresses: [Address!]
}

type Product {
  name: String!
  price: Float!
  manufacturer: Company!
  components: [Product!]
}

type StringProductPair {
  key: String!
  value: Product!
}

type StringStringPair {
  key: String!
  value: String!
}
`[1:]

	require.Nil(t, err)
	require.Equal(t, expectedSchema, result.Schema)
}

func Test_ConvertType_Go(t *testing.T) {

	lti := languages.GoLang().TypeInfo()

	testCases := []struct {
		sourceType          string
		forInput            bool
		expectedGraphQLType string
		sourceTypeDefs      []*metadata.TypeDefinition
		expectedTypeDefs    []*TypeDefinition
	}{
		// Plain non-nullable types
		{"string", false, "String!", nil, nil},
		{"string", true, "String!", nil, nil},
		{"bool", false, "Boolean!", nil, nil},
		{"bool", true, "Boolean!", nil, nil},
		{"int8", false, "Int!", nil, nil},
		{"int8", true, "Int!", nil, nil},
		{"int16", false, "Int!", nil, nil},
		{"int16", true, "Int!", nil, nil},
		{"int32", false, "Int!", nil, nil},
		{"int32", true, "Int!", nil, nil},
		{"uint8", false, "Int!", nil, nil},
		{"uint8", true, "Int!", nil, nil},
		{"uint16", false, "Int!", nil, nil},
		{"uint16", true, "Int!", nil, nil},
		{"float32", false, "Float!", nil, nil},
		{"float32", true, "Float!", nil, nil},
		{"float64", false, "Float!", nil, nil},
		{"float64", true, "Float!", nil, nil},

		// Slice types
		{"[]string", false, "[String!]", nil, nil},
		{"[]string", true, "[String!]", nil, nil},
		{"[][]string", false, "[[String!]]", nil, nil},
		{"[][]string", true, "[[String!]]", nil, nil},
		{"[]*string", false, "[String]", nil, nil},
		{"[]*string", true, "[String]", nil, nil},

		// Custom scalar types
		{"time.Time", false, "Timestamp!", nil, []*TypeDefinition{{Name: "Timestamp"}}},
		{"time.Time", true, "Timestamp!", nil, []*TypeDefinition{{Name: "Timestamp"}}},
		{"int64", false, "Int64!", nil, []*TypeDefinition{{Name: "Int64"}}},
		{"int64", true, "Int64!", nil, []*TypeDefinition{{Name: "Int64"}}},
		{"uint32", false, "UInt!", nil, []*TypeDefinition{{Name: "UInt"}}},
		{"uint32", true, "UInt!", nil, []*TypeDefinition{{Name: "UInt"}}},
		{"uint64", false, "UInt64!", nil, []*TypeDefinition{{Name: "UInt64"}}},
		{"uint64", true, "UInt64!", nil, []*TypeDefinition{{Name: "UInt64"}}},

		// Custom types
		{"testdata.User", false, "User!",
			[]*metadata.TypeDefinition{{
				Name: "User",
				Fields: []*metadata.Field{
					{Name: "firstName", Type: "string"},
					{Name: "lastName", Type: "string"},
					{Name: "age", Type: "uint8"},
				},
			}},
			[]*TypeDefinition{{
				Name: "User",
				Fields: []*NameTypePair{
					{"firstName", "String!"},
					{"lastName", "String!"},
					{"age", "Int!"},
				},
			}}},
		{"testdata.User", true, "UserInput!",
			[]*metadata.TypeDefinition{{
				Name: "User",
				Fields: []*metadata.Field{
					{Name: "firstName", Type: "string"},
					{Name: "lastName", Type: "string"},
					{Name: "age", Type: "uint8"},
				},
			}},
			[]*TypeDefinition{{
				Name: "UserInput",
				Fields: []*NameTypePair{
					{"firstName", "String!"},
					{"lastName", "String!"},
					{"age", "Int!"},
				},
			}}},

		{"*bool", false, "Boolean", nil, nil},
		{"*bool", true, "Boolean", nil, nil},
		{"*int", false, "Int", nil, nil},
		{"*int", true, "Int", nil, nil},
		{"*float64", false, "Float", nil, nil},
		{"*float64", true, "Float", nil, nil},
		{"*string", false, "String", nil, nil},
		{"*string", true, "String", nil, nil},
		{"*testdata.Foo", false, "Foo", // scalar
			[]*metadata.TypeDefinition{{Name: "testdata.Foo"}},
			[]*TypeDefinition{{Name: "Foo"}}},
		{"*testdata.Foo", true, "Foo", // scalar
			[]*metadata.TypeDefinition{{Name: "testdata.Foo"}},
			[]*TypeDefinition{{Name: "Foo"}}},

		// Map types
		{"map[string]string", false, "[StringStringPair!]", nil, []*TypeDefinition{{
			Name: "StringStringPair",
			Fields: []*NameTypePair{
				{"key", "String!"},
				{"value", "String!"},
			},
			IsMapType: true,
		}}},
		{"map[string]string", true, "[StringStringPairInput!]", nil, []*TypeDefinition{{
			Name: "StringStringPairInput",
			Fields: []*NameTypePair{
				{"key", "String!"},
				{"value", "String!"},
			},
			IsMapType: true,
		}}},
		{"map[string]*string", false, "[StringNullableStringPair!]", nil, []*TypeDefinition{{
			Name: "StringNullableStringPair",
			Fields: []*NameTypePair{
				{"key", "String!"},
				{"value", "String"},
			},
			IsMapType: true,
		}}},
		{"map[string]*string", true, "[StringNullableStringPairInput!]", nil, []*TypeDefinition{{
			Name: "StringNullableStringPairInput",
			Fields: []*NameTypePair{
				{"key", "String!"},
				{"value", "String"},
			},
			IsMapType: true,
		}}},
		{"map[int32]string", false, "[IntStringPair!]", nil, []*TypeDefinition{{
			Name: "IntStringPair",
			Fields: []*NameTypePair{
				{"key", "Int!"},
				{"value", "String!"},
			},
			IsMapType: true,
		}}},
		{"map[int32]string", true, "[IntStringPairInput!]", nil, []*TypeDefinition{{
			Name: "IntStringPairInput",
			Fields: []*NameTypePair{
				{"key", "Int!"},
				{"value", "String!"},
			},
			IsMapType: true,
		}}},
		{"map[string]map[string]float32", false, "[StringNullableStringFloatPairListPair!]", nil, []*TypeDefinition{
			{
				Name: "StringNullableStringFloatPairListPair",
				Fields: []*NameTypePair{
					{"key", "String!"},
					{"value", "[StringFloatPair!]"},
				},
				IsMapType: true,
			},
			{
				Name: "StringFloatPair",
				Fields: []*NameTypePair{
					{"key", "String!"},
					{"value", "Float!"},
				},
				IsMapType: true,
			},
		}},
		{"map[string]map[string]float32", true, "[StringNullableStringFloatPairListPairInput!]", nil, []*TypeDefinition{
			{
				Name: "StringNullableStringFloatPairListPairInput",
				Fields: []*NameTypePair{
					{"key", "String!"},
					{"value", "[StringFloatPairInput!]"},
				},
				IsMapType: true,
			},
			{
				Name: "StringFloatPairInput",
				Fields: []*NameTypePair{
					{"key", "String!"},
					{"value", "Float!"},
				},
				IsMapType: true,
			},
		}},
	}

	for _, tc := range testCases {
		testName := strings.ReplaceAll(tc.sourceType, " ", "")
		if tc.forInput {
			testName += "_input"
		}
		t.Run(testName, func(t *testing.T) {

			types := make(metadata.TypeMap, len(tc.sourceTypeDefs))
			for _, td := range tc.sourceTypeDefs {
				types[td.Name] = td
			}

			typeDefs, errors := transformTypes(types, lti, tc.forInput)
			require.Empty(t, errors)

			result, err := convertType(tc.sourceType, lti, typeDefs, false, tc.forInput)

			require.Nil(t, err)
			require.Equal(t, tc.expectedGraphQLType, result)

			if tc.expectedTypeDefs == nil {
				require.Empty(t, typeDefs)
			} else {
				require.ElementsMatch(t, tc.expectedTypeDefs, utils.MapValues(typeDefs))
			}
		})
	}
}