package http

type parseAgentIDCase struct {
	name    string
	header  string
	wantNil bool
	wantID  string
}

func parseAgentIDCases() []parseAgentIDCase {
	return []parseAgentIDCase{
		{name: "empty", header: "", wantNil: true},
		{name: "whitespace", header: "  ", wantNil: true},
		{name: "invalid", header: "not-a-uuid", wantNil: true},
		{name: "valid", header: agentTestAgentID(), wantID: agentTestAgentID()},
	}
}
