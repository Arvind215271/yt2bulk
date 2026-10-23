package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPlanner(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{"Valid fields", []string{FieldID, FieldTitle}, false},
		{"Invalid field", []string{"invalid_field"}, true},
		{"Mixed valid and invalid fields", []string{FieldID, "invalid"}, true},
		{"Empty fields", []string{}, false},
		{"Duplicate fields", []string{FieldID, FieldID}, false},
		{"All valid fields", []string{
			FieldID, FieldTitle, FieldChannelID, FieldChannelTitle,
			FieldThumbnails, FieldPublishedAt, FieldDuration, FieldDurationSeconds,
			FieldDurationMinutes, FieldDurationTimestamp, FieldViewCount, FieldLikeCount,
			FieldCommentCount, FieldTags, FieldCategoryID, FieldCaptionAvailable,
			FieldPrivacyStatus, FieldLiveBroadcastStatus, FieldDescriptionChapters,
			FieldDescriptionLinks, FieldDescriptionEmails, FieldDescriptionCleaned,
			FieldTranscriptText, FieldTranscriptSignal,
		}, false},
		{"Nil input", nil, false},
		{"Single valid field", []string{FieldTitle}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planner, err := NewPlanner(tt.fields, nil)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, planner)
			}
		})
	}
}

func TestPlanner_Has(t *testing.T) {
	planner, _ := NewPlanner([]string{FieldID, FieldTitle}, nil)

	assert.True(t, planner.Has(FieldID))
	assert.True(t, planner.Has(FieldTitle))
	assert.False(t, planner.Has("non_existent"))
	assert.False(t, planner.Has(""))

	plannerAll, _ := NewPlanner([]string{}, nil)
	assert.True(t, plannerAll.Has(FieldID))
	// Now with default fields, Has("anything") will be false unless it's in default fields
	assert.False(t, plannerAll.Has("anything"))
}

func TestPlanner_ExportsEverything(t *testing.T) {
	// With empty/nil, it now uses DefaultFields(), which has content, so ExportsEverything is false
	plannerEmpty, _ := NewPlanner([]string{}, nil)
	assert.False(t, plannerEmpty.ExportsEverything())

	plannerNil, _ := NewPlanner(nil, nil)
	assert.False(t, plannerNil.ExportsEverything())

	plannerWithFields, _ := NewPlanner([]string{FieldID}, nil)
	assert.False(t, plannerWithFields.ExportsEverything())
}

func TestPlanner_HasAny(t *testing.T) {
	planner, _ := NewPlanner([]string{FieldID, FieldTitle}, nil)

	assert.True(t, planner.HasAny([]string{FieldID, FieldTags}))
	assert.True(t, planner.HasAny([]string{FieldTitle}))
	assert.True(t, planner.HasAny([]string{FieldID, FieldTitle}))
	assert.False(t, planner.HasAny([]string{FieldTags}))
	assert.False(t, planner.HasAny([]string{}))
	assert.False(t, planner.HasAny([]string{"unknown"}))

	plannerAll, _ := NewPlanner([]string{}, nil)
	assert.True(t, plannerAll.HasAny([]string{FieldID}))
	assert.False(t, plannerAll.HasAny([]string{"anything"}))
}

func TestPlanner_ExportFields(t *testing.T) {
	fields := []string{FieldID, FieldTitle}
	planner, _ := NewPlanner(fields, nil)
	exported := planner.ExportFields()

	assert.ElementsMatch(t, fields, exported)

	exported[0] = "changed"
	assert.NotEqual(t, exported[0], planner.ExportFields()[0])

	plannerAll, _ := NewPlanner([]string{}, nil)
	assert.NotNil(t, plannerAll.ExportFields())
}

func TestValidateFields(t *testing.T) {
	assert.Nil(t, ValidateFields([]string{FieldID, FieldTitle}))
	assert.Nil(t, ValidateFields(nil))
	assert.Nil(t, ValidateFields([]string{}))

	err := ValidateFields([]string{"bad_field", "other_bad"})
	assert.NotNil(t, err)

	err = ValidateFields([]string{"a", "b"})
	assert.NotNil(t, err)
}

func TestDefaultFields(t *testing.T) {
	defs := DefaultFields()
	assert.NotEmpty(t, defs)

	// Validate default fields
	err := ValidateFields(defs)
	assert.Nil(t, err, "Default fields should all be valid")
}
