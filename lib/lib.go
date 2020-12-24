package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gobuffalo/flect"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"sigs.k8s.io/yaml"
)

type SchemaChecker struct {
	// project root directory
	RootDir  string
	Registry map[string]reflect.Type
}

func kind(v interface{}) string {
	return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
}

func (checker *SchemaChecker) makeInstance(name string) interface{} {
	v := reflect.New(checker.Registry[name]).Elem()
	// Maybe fill in fields here if necessary
	return v.Interface()
}

// https://stackoverflow.com/a/23031445

func New(objs []interface{}) *SchemaChecker {
	reg := map[string]reflect.Type{}
	for _, v := range objs {
		reg[fmt.Sprintf("%T", v)] = reflect.TypeOf(v)
	}
	return &SchemaChecker{
		Registry: reg,
	}
}

func (checker *SchemaChecker) TestChart(t *testing.T, chartName string) {
	schemaKind := flect.Titleize(chartName) + "Spec"
	valuesfile := filepath.Join(checker.RootDir, "charts", chartName, "values.yaml")
	checker.TestDefaultValues(t, schemaKind, valuesfile)
}

func (checker *SchemaChecker) TestKind(t *testing.T, kind string) {
	schemaKind := flect.Titleize(kind + "Spec")
	valuesfile := filepath.Join(checker.RootDir, "charts", flect.Dasherize(kind), "values.yaml")
	checker.TestDefaultValues(t, schemaKind, valuesfile)
}

func (checker *SchemaChecker) TestDefaultValues(t *testing.T, schemaKind string, valuesfile string) {
	diffstring, err := checker.compareDefaultValues(schemaKind, valuesfile)
	if err != nil {
		t.Error(err)
	}
	if diffstring != "" {
		t.Errorf("values file does not match, diff: %s", diffstring)
	}
}

func (checker *SchemaChecker) compareDefaultValues(schemaKind string, valuesfile string) (string, error) {
	data, err := ioutil.ReadFile(valuesfile)
	if err != nil {
		return "", err
	}

	var original map[string]interface{}
	err = yaml.Unmarshal(data, &original)
	if err != nil {
		return "", err
	}
	sorted, err := json.Marshal(&original)
	if err != nil {
		return "", err
	}

	spec := checker.makeInstance(schemaKind)
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		return "", err
	}
	parsed, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	// Then, compare them
	differ := diff.New()
	d, err := differ.Compare(sorted, parsed)
	if err != nil {
		fmt.Printf("Failed to unmarshal file: %s\n", err.Error())
		os.Exit(3)
	}

	if d.Modified() {
		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		f := formatter.NewAsciiFormatter(original, config)
		diffString, err := f.Format(d)
		if err != nil {
			return "", err
		}
		return diffString, nil
	}

	return "", nil
}
