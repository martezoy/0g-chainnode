package types

const (
	DefaultNextRequestID = 0
)

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(nextRequestID uint64, requests []DASRequest, responses []DASResponse) *GenesisState {
	return &GenesisState{
		NextRequestID: nextRequestID,
		Requests:      requests,
		Responses:     responses,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		DefaultNextRequestID,
		[]DASRequest{},
		[]DASResponse{},
	)
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	return nil
}
