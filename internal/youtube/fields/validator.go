package fields

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

// ValidateFields ensures all requested fields are valid using the registry.
func ValidateFields(fields []string) error {
	err := youtube.Err.Export.InvalidField()
	var foundInvalid bool
	for _, f := range fields {
		if _, ok := Registry[f]; !ok {
			err.AddField(f, "Invalid field")
			foundInvalid = true
		}
	}
	if foundInvalid {
		return err
	}
	return nil
}
