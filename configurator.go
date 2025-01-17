// Package configuration provides ability to initialize your custom configuration struct from: flags, environment variables, `default` tag, files (json, yaml)
package configuration

import (
	"fmt"
	"log"
	"reflect"
)

// New creates a new instance of the Configurator.
func New(
	cfgPtr any, // must be a pointer to a struct
	providers ...Provider, // providers will be executed in order of their declaration
) *Configurator {
	return &Configurator{
		configPtr:      cfgPtr,
		providers:      providers,
		registeredTags: map[string]struct{}{},
		loggerFn:       log.Printf,
		onErrorFn: func(err error) {
			if err != nil {
				log.Fatal(err)
			}
		},
		loggingEnabled: false,
	}
}

type Configurator struct {
	configPtr      any
	providers      []Provider
	registeredTags map[string]struct{}
	onErrorFn      func(err error)
	loggerFn       func(format string, v ...any)
	loggingEnabled bool
}

func (c *Configurator) SetOptions(options ...ConfiguratorOption) *Configurator {
	for _, o := range options {
		o(c)
	}
	return c
}

// InitValues sets values into struct field using given set of providers
// respecting their order: first defined -> first executed
func (c *Configurator) InitValues() error {
	if reflect.TypeOf(c.configPtr).Kind() != reflect.Ptr {
		return ErrNotAPointer
	}

	if len(c.providers) == 0 {
		return ErrNoProviders
	}

	for _, p := range c.providers {
		if _, ok := c.registeredTags[p.Name()]; ok {
			return ErrProviderNameCollision
		}
		c.registeredTags[p.Name()] = struct{}{}

		if err := p.Init(c.configPtr); err != nil {
			return fmt.Errorf("cannot init [%s] provider: %v", p.Name(), err)
		}
	}

	c.fillUp(c.configPtr)
	return nil
}

func (c *Configurator) fillUp(i any) {
	var (
		t = reflect.TypeOf(i)
		v = reflect.ValueOf(i)
	)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		var (
			tField = t.Field(i)
			vField = v.Field(i)
		)

		if tField.Type.Kind() == reflect.Struct {
			c.fillUp(vField.Addr().Interface())
			continue
		}

		if tField.Type.Kind() == reflect.Ptr && tField.Type.Elem().Kind() == reflect.Struct {
			vField.Set(reflect.New(tField.Type.Elem()))
			c.fillUp(vField.Interface())
			continue
		}

		c.applyProviders(tField, vField)
	}
}

func (c *Configurator) applyProviders(field reflect.StructField, v reflect.Value) {
	if !field.IsExported() {
		return
	}

	for _, provider := range c.providers {
		if provider.Provide(field, v) == nil {
			return
		}
	}

	c.onErrorFn(fmt.Errorf("configurator: field [%s] with tags [%v] cannot be set", field.Name, field.Tag))
}

// FromEnvAndDefault is a shortcut for `New(cfg, NewEnvProvider(), NewDefaultProvider()).InitValues()`.
func FromEnvAndDefault(cfg any) error {
	return New(cfg, NewEnvProvider(), NewDefaultProvider()).InitValues()
}
