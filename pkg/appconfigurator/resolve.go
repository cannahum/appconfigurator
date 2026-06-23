package appconfigurator

import (
	"fmt"
	"os"
	"reflect"
)

const resolveTag = "resolve"

// Resolve walks a non-nil pointer to a struct and resolves fields by their
// `resolve:"..."` tag:
//
//	env  — a string field whose value is an environment variable NAME; replaced
//	       with that variable's value (empty if unset).
//	file — an appconfigurator.File field; the file at its Path is read into Data.
//
// A field whose env reference is empty is left untouched (supports values that
// are omitted in some environments). Untagged fields are never touched — a plain
// string path you don't want read simply carries no tag. Nested structs and
// non-nil struct pointers are recursed.
//
// Future resolvers (e.g. "secret" for AWS Secrets Manager) slot in here.
func Resolve(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("appconfigurator: Resolve requires a non-nil pointer, got %T", v)
	}
	return resolve(rv.Elem())
}

func resolve(rv reflect.Value) error {
	switch rv.Kind() {
	case reflect.Pointer:
		if !rv.IsNil() {
			return resolve(rv.Elem())
		}
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			fv := rv.Field(i)
			if !fv.CanSet() {
				continue // unexported
			}
			if kind, tagged := rt.Field(i).Tag.Lookup(resolveTag); tagged {
				if err := resolveField(fv, kind, rt.Field(i).Name); err != nil {
					return err
				}
				continue
			}
			if fv.Kind() == reflect.Struct || fv.Kind() == reflect.Pointer {
				if err := resolve(fv); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func resolveField(fv reflect.Value, kind, name string) error {
	switch kind {
	case "env":
		if fv.Kind() != reflect.String {
			return fmt.Errorf("appconfigurator: field %s: resolve:%q requires a string field", name, kind)
		}
		if ref := fv.String(); ref != "" {
			fv.SetString(os.Getenv(ref))
		}
		return nil
	case "file":
		f, ok := fv.Addr().Interface().(*File)
		if !ok {
			return fmt.Errorf("appconfigurator: field %s: resolve:\"file\" requires an appconfigurator.File field", name)
		}
		if err := f.load(); err != nil {
			return fmt.Errorf("appconfigurator: field %s: %w", name, err)
		}
		return nil
	default:
		return fmt.Errorf("appconfigurator: field %s: unknown resolver %q", name, kind)
	}
}
