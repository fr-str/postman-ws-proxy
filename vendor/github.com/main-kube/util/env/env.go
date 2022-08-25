package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
	"syslabit.com/git/syslabit/log"
)

type types interface {
	~bool | ~[]string | constraints.Ordered
}

// Get gets variable from env, if not found return default value
// If defaultValue is set and variable not found, then panics
func Get[T types](envName string, defaultValue ...T) T {
	value := os.Getenv(envName)

	var ret any = value
	var err error

	var def T
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}

	switch any(def).(type) {
	case string:
		ret = value

	case bool:
		ret, err = strconv.ParseBool(value)

	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint:
		ret, err = strconv.Atoi(value)
		ret = reflect.ValueOf(ret).Convert(reflect.TypeOf(def)).Interface()

	case float64:
		ret, err = strconv.ParseFloat(value, 64)

	case []string:
		if strings.Contains(value, ";") {
			ret = strings.Split(value, ";")
		} else {
			ret = strings.Split(value, ",")
		}
	}

	switch {
	case value == "" && len(defaultValue) == 0:
		log.Fatal("Required variable {{name}} is not set - type: {{type}}", log.Vars{
			"name": envName,
			"type": fmt.Sprintf("%T", def),
		})
	case value == "":
		ret = def
	case err != nil:
		log.Fatal("Variable {{name}} could not be parsed - type: {{type}}, value: {{value}}", log.Vars{
			"name":  envName,
			"type":  fmt.Sprintf("%T", def),
			"value": value,
		})
	}

	return ret.(T)
}
