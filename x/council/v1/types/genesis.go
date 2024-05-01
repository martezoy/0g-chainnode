package types

const (
	DefaultVotingStartHeight = 1
	DefaultVotingPeriod      = 200
)

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(params Params, votingStartHeight uint64, votingPeriod uint64, currentCouncilID uint64, councils Councils) *GenesisState {
	return &GenesisState{
		Params:            params,
		VotingStartHeight: votingStartHeight,
		VotingPeriod:      votingPeriod,
		CurrentCouncilID:  currentCouncilID,
		Councils:          councils,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(
		Params{
			CouncilSize: 1,
		},
		DefaultVotingStartHeight,
		DefaultVotingPeriod,
		1,
		[]Council{
			{
				ID:                1,
				VotingStartHeight: DefaultVotingStartHeight,
				StartHeight:       DefaultVotingStartHeight + DefaultVotingPeriod,
				EndHeight:         DefaultVotingStartHeight + DefaultVotingPeriod*2,
				Votes:             Votes{},
			}},
	)
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	return nil
}
