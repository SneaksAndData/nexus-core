package configurations

import (
	"context"
	"github.com/aws/smithy-go/ptr"
	"os"
	"reflect"
	"testing"
)

type NestedConfig struct {
	Prop1 string `mapstructure:"prop1"`
}

type TestConfig struct {
	Prop1       string        `mapstructure:"prop1,omitempty"`
	Prop2       *int          `mapstructure:"prop2,omitempty"`
	NestedProp2 *NestedConfig `mapstructure:"nested-prop2,omitempty"`
}

func getExpectedConfig() *TestConfig {
	return &TestConfig{
		Prop1: "prop1",
		Prop2: ptr.Int(1),
		NestedProp2: &NestedConfig{
			Prop1: "prop1",
		},
	}
}

func Test_LoadConfig(t *testing.T) {
	var expected = getExpectedConfig()

	var result = LoadConfig[TestConfig](context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}

func Test_LoadConfigFromEnv(t *testing.T) {
	prop2 := "prop2"
	_ = os.Setenv("NEXUS__NESTED_PROP2__PROP1", prop2)

	var expected = getExpectedConfig()
	expected.NestedProp2.Prop1 = prop2

	var result = LoadConfig[TestConfig](context.TODO())
	if !reflect.DeepEqual(*expected, result) {
		t.Errorf("LoadConfig failed, expected %v, got %v", *expected, result)
	}
}
