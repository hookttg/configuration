package configuration

import (
	"fmt"
	"reflect"
)

const DefaultProviderName = `DefaultProvider`

// NewDefaultProvider creates new provider which sets values from `default` tag
func NewDefaultProvider() defaultProvider {
	return defaultProvider{}
}

type defaultProvider struct{}

func (defaultProvider) Name() string {
	return DefaultProviderName
}

func (defaultProvider) Init(_ any) error {
	return nil
}

func (dp defaultProvider) Provide(field reflect.StructField, v reflect.Value) error {
	valStr := field.Tag.Get("default")
	if len(valStr) == 0 {
		//return fmt.Errorf("defaultProvider: %w", ErrEmptyValue)
		return
	}

	return SetField(field, v, valStr)
}
