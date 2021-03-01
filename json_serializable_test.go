package iotago_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicJSONArrayDeserialization(t *testing.T) {
	jsonData := `{"array": [{"type": 0, "name": "Alice"}, {"type": 1, "color": "violet"}]}`

	type envelope struct {
		Array []*json.RawMessage `json:"array"`
	}

	type typeenvelope struct {
		Type int `json:"type"`
	}

	type objwithname struct {
		Name string `json:"name"`
	}

	type objwithcolor struct {
		Color string `json:"color"`
	}

	env := &envelope{}
	assert.NoError(t, json.Unmarshal([]byte(jsonData), env))

	for _, ele := range env.Array {
		eleJson, err := ele.MarshalJSON()
		assert.NoError(t, err)
		envty := &typeenvelope{}
		assert.NoError(t, json.Unmarshal(eleJson, envty))
		var obj interface{}
		switch envty.Type {
		case 0:
			obj = &objwithname{}
		case 1:
			obj = &objwithcolor{}
		}
		assert.NoError(t, json.Unmarshal(eleJson, obj))
		fmt.Printf("%v\n", obj)
	}

}
