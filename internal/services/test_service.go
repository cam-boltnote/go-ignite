package services

// TestService provides a simple test endpoint
type TestService struct{}

// NewTestService creates a new test service instance
func NewTestService() *TestService {
	return &TestService{}
}

// GetTestMessage returns a test message
func (s *TestService) GetTestMessage() string {
	return "the work is mysterious and important"
}
