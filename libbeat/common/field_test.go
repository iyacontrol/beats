// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package common

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/go-ucfg/yaml"
)

func TestFieldsHasKey(t *testing.T) {
	tests := []struct {
		key    string
		fields Fields
		result bool
	}{
		{
			key:    "test.find",
			fields: Fields{},
			result: false,
		},
		{
			key: "test.find",
			fields: Fields{
				Field{Name: "test"},
				Field{Name: "find"},
			},
			result: false,
		},
		{
			key: "test.find",
			fields: Fields{
				Field{
					Name: "test", Fields: Fields{
						Field{
							Name: "find",
						},
					},
				},
			},
			result: true,
		},
		{
			key: "test",
			fields: Fields{
				Field{
					Name: "test", Fields: Fields{
						Field{
							Name: "find",
						},
					},
				},
			},
			result: false,
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.result, test.fields.HasKey(test.key))
	}
}

func TestDynamicYaml(t *testing.T) {
	tests := []struct {
		input  []byte
		output Field
		error  bool
	}{
		{
			input: []byte(`
name: test
dynamic: true`),
			output: Field{
				Name:    "test",
				Dynamic: DynamicType{true},
			},
		},
		{
			input: []byte(`
name: test
dynamic: "true"`),
			output: Field{
				Name:    "test",
				Dynamic: DynamicType{true},
			},
		},
		{
			input: []byte(`
name: test
dynamic: "blue"`),
			error: true,
		},
		{
			input: []byte(`
name: test
dynamic: "strict"`),
			output: Field{
				Name:    "test",
				Dynamic: DynamicType{"strict"},
			},
		},
	}

	for _, test := range tests {
		keys := Field{}

		cfg, err := yaml.NewConfig(test.input)
		assert.NoError(t, err)
		err = cfg.Unpack(&keys)

		if err != nil {
			assert.True(t, test.error)
		} else {
			assert.Equal(t, test.output.Dynamic, keys.Dynamic)
		}
	}
}

func TestGetKeys(t *testing.T) {
	tests := []struct {
		fields Fields
		keys   []string
	}{
		{
			fields: Fields{
				Field{
					Name: "test", Fields: Fields{
						Field{
							Name: "find",
						},
					},
				},
			},
			keys: []string{"test.find"},
		},
		{
			fields: Fields{
				Field{
					Name: "a", Fields: Fields{
						Field{
							Name: "b",
						},
					},
				},
				Field{
					Name: "a", Fields: Fields{
						Field{
							Name: "c",
						},
					},
				},
			},
			keys: []string{"a.b", "a.c"},
		},
		{
			fields: Fields{
				Field{
					Name: "a",
				},
				Field{
					Name: "b",
				},
				Field{
					Name: "c",
				},
			},
			keys: []string{"a", "b", "c"},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.keys, test.fields.GetKeys())
	}
}

func TestFieldValidate(t *testing.T) {
	tests := []struct {
		cfg   MapStr
		field Field
		err   bool
		name  string
	}{
		{
			cfg:   MapStr{"object_type": "scaled_float", "object_type_mapping_type": "float", "scaling_factor": 10},
			field: Field{ObjectType: "scaled_float", ObjectTypeMappingType: "float", ScalingFactor: 10},
			err:   false,
			name:  "top level object type config",
		}, {
			cfg: MapStr{"object_type_params": []MapStr{
				{"object_type": "scaled_float", "object_type_mapping_type": "float", "scaling_factor": 100}}},
			field: Field{ObjectTypeParams: []ObjectTypeCfg{{ObjectType: "scaled_float", ObjectTypeMappingType: "float", ScalingFactor: 100}}},
			err:   false,
			name:  "multiple object type configs",
		}, {
			cfg: MapStr{
				"object_type":        "scaled_float",
				"object_type_params": []MapStr{{"object_type": "scaled_float", "object_type_mapping_type": "float"}}},
			err:  true,
			name: "invalid config mixing object_type and object_type_params",
		}, {
			cfg: MapStr{
				"object_type_mapping_type": "float",
				"object_type_params":       []MapStr{{"object_type": "scaled_float", "object_type_mapping_type": "float"}}},
			err:  true,
			name: "invalid config mixing object_type_mapping_type and object_type_params",
		}, {
			cfg: MapStr{
				"scaling_factor":     100,
				"object_type_params": []MapStr{{"object_type": "scaled_float", "object_type_mapping_type": "float"}}},
			err:  true,
			name: "invalid config mixing scaling_factor and object_type_params",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg, err := NewConfigFrom(test.cfg)
			require.NoError(t, err)
			var f Field
			err = cfg.Unpack(&f)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.field, f)
			}
		})
	}

}
