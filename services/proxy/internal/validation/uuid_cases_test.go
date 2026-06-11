package validation

type uuidFieldCase struct {
	name      string
	value     string
	field     string
	wantNil   bool
	wantCode  string
	wantField string
}

func uuidFieldCases() []uuidFieldCase {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	return []uuidFieldCase{
		{name: "valid", value: validUUID, wantNil: true},
		{name: "empty", value: "", wantCode: fieldCodeRequired, wantField: "agent_id"},
		{name: "whitespace only", value: "   ", wantCode: fieldCodeRequired, wantField: "agent_id"},
		{name: "malformed", value: "not-a-uuid", wantCode: fieldCodeInvalidFormat, wantField: "agent_id"},
		{name: "too short", value: "550e8400-e29b-41d4-a716", wantCode: fieldCodeInvalidFormat, wantField: "org_id"},
		{name: "non hex", value: "gggggggg-gggg-gggg-gggg-gggggggggggg", wantCode: fieldCodeInvalidFormat, wantField: "org_id"},
		{name: "valid trimmed", value: "  " + validUUID + "  ", wantNil: true},
	}
}
