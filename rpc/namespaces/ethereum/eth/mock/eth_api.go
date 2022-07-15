package mock

// MockPublicAPI will send all eth related request to the named API,
// so you can test API behavior from a client without needing
// an entire Ethermint node
type MockPublicAPI struct {
}

var (
// _ eth.PublicAPI = (*MockPublicAPI)(nil)

// _ client.ABCIClient = ABCIApp{}
// _ client.ABCIClient = ABCIMock{}
// _ client.ABCIClient = (*ABCIRecorder)(nil)
)
