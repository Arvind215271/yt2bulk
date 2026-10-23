package fields

import "github.com/Arvind215271/yt2bulk/internal/logger"

// NewPlanner creates a new Planner instance, validating the fields.
func NewPlanner(fields []string, logger *logger.Logger) (*Planner, error) {

	if len(fields) == 0 {
		fields = DefaultFields()
	}

	if err := ValidateFields(fields); err != nil {
		if logger != nil {
			logger.Error("failed to validate fields", "error", err)
		}
		return nil, err
	}

	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}
	// Always include errors
	fieldSet[FieldErrors] = true

	if logger != nil {
		logger.Debug("created fields planner", "count", len(fields))
	}

	return &Planner{
		fields:   fields,
		fieldSet: fieldSet,
		logger:   logger,
	}, nil
}

// Has checks if the requested field exists in the planner.
func (p *Planner) Has(field string) bool {
	if p.ExportsEverything() {
		return true
	}
	return p.fieldSet[field]
}

// HasAny checks if any of the provided fields are requested.
func (p *Planner) HasAny(fields []string) bool {
	if p.ExportsEverything() {
		return true
	}
	for _, f := range fields {
		if p.Has(f) {
			return true
		}
	}
	return false
}

// ExportFields returns the validated list of requested export fields.
func (p *Planner) ExportFields() []string {
	if p.ExportsEverything() {
		return nil
	}
	// Return a copy to ensure immutability
	out := make([]string, len(p.fields))
	copy(out, p.fields)
	return out
}

// ExportsEverything returns true if no specific fields were requested.
func (p *Planner) ExportsEverything() bool {
	return len(p.fields) == 0
}
