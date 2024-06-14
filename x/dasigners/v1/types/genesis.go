package types

import "fmt"

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(params Params, epoch uint64, signers []*Signer, quorumsByEpoch []*Quorums) *GenesisState {
	return &GenesisState{
		Params:         params,
		EpochNumber:    epoch,
		Signers:        signers,
		QuorumsByEpoch: quorumsByEpoch,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(Params{
		TokensPerVote:     10,
		MaxVotesPerSigner: 1024,
		MaxQuorums:        10,
		EpochBlocks:       5760,
		EncodedSlices:     3072,
	}, 0, make([]*Signer, 0), []*Quorums{{
		Quorums: make([]*Quorum, 0),
	}})
}

// Validate performs basic validation of genesis data.
func (gs GenesisState) Validate() error {
	registered := make(map[string]struct{})
	for _, signer := range gs.Signers {
		if err := signer.Validate(); err != nil {
			return err
		}
		registered[signer.Account] = struct{}{}
	}
	if len(gs.QuorumsByEpoch) != int(gs.EpochNumber)+1 {
		return fmt.Errorf("epoch history missing")
	}
	for _, quorums := range gs.QuorumsByEpoch {
		for _, quorum := range quorums.Quorums {
			for _, signer := range quorum.Signers {
				if err := ValidateHexAddress(signer); err != nil {
					return err
				}
				if _, ok := registered[signer]; !ok {
					return fmt.Errorf("historical signer detail missing")
				}
			}
		}
	}
	return nil
}
