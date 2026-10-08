package export

import (
	"reflect"
	"strings"

	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/fields"
)

func getJSONKey(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	key := parts[0]
	if key == "" {
		key = f.Name
	}
	return key, true
}

// exportStruct converts any data passed to it to exported Data Type.
//
// It would deal with Playlist and Videos only for current purpose.
// we convert this data to JSON. So we can simply use this data to convert to other format we might need in future.
func exportStruct(v any, planner *fields.Planner) (ExportData, error) {
	//  this creates a map of string key and value as any
	out := make(ExportData)

	// Guard against nil planner - requiring a non-nil planner is expected
	if planner == nil {
		return nil, youtube.Err.Export.MarshalFailed()
	}

	// reflect value

	// this gives us the actual value stored in the binary. Because to create our struct, GO allocate Bytes in sizes. And we use that to store that. So this is where we can check what is stored in each field of the struct. (actual data)
	val := reflect.ValueOf(v)
	// here, gives us the struct definitoin. Like which field exist, what is the type of that field, etc. So we can simply use that to extract the field that we are required or given in the map to this function and filter those out... based on this typ.
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return out, nil
		}
		val = val.Elem()
	}

	typ := val.Type()

	// check if export everything
	exportAll := planner.ExportsEverything()

	// for each field present in struct... do a for loop .
	for i := 0; i < typ.NumField(); i++ {
		// get the field from the struct definition
		structField := typ.Field(i)
		if structField.PkgPath != "" {
			continue
		}

		// use json tag as export key (this is same as we write when defining structs)
		key, ok := getJSONKey(structField)
		if !ok {
			continue
		}

		// check if present or all keys have to be returned
		if exportAll || planner.Has(key) {
			// firstly, val store the actual data. Then we need to know which field to get the data from. Interface convert it to any data type. That is all
			out[key] = val.Field(i).Interface()
		}
	}

	return out, nil
}
