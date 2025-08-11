package util

import (
	"reflect"
	"testing"
)

type cloneSourceOther struct {
	Prop1 string `json:"prop1"`
}

type cloneSource struct {
	Prop1 string            `json:"prop1"`
	Prop2 int               `json:"prop2"`
	Prop3 *cloneSourceOther `json:"prop3"`
}

func Test_DeepClone(t *testing.T) {
	toClone := cloneSource{
		Prop1: "value1",
		Prop2: 2,
		Prop3: &cloneSourceOther{
			Prop1: "value2",
		},
	}

	cloned, err := DeepClone(toClone)

	if err != nil {
		t.Errorf("deep clone failed: %v", err)
	}

	if !reflect.DeepEqual(*cloned, toClone) {
		t.Errorf("deep clone failed: object content differs, expected %v, got %v", toClone, cloned)
	}
}

func Test_CoalesceCollection(t *testing.T) {
	c1 := []string{"a", "b", "c"}
	var c2 []string

	c3 := CoalesceCollection[string](c1, c2)

	if !reflect.DeepEqual(c3, c1) {
		t.Errorf("failed to coalesce a collection: object content differs, expected %v, got %v", c1, c3)
	}
}
