package auth

type grpcValidatorCase struct {
	name    string
	client  *mockAuthServiceClient
	want    *ValidateResult
	wantErr error
}

type agentVerifierCase struct {
	name    string
	client  *mockAuthServiceClient
	want    *AgentRecord
	wantErr error
}
