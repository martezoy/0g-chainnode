package types

import "fmt"

// NewGenesisState returns a new genesis state object for the module.
func NewGenesisState(params Params, epoch uint64, signers []*Signer, signersByEpoch []*EpochSignerSet) *GenesisState {
	return &GenesisState{
		Params:         params,
		EpochNumber:    epoch,
		Signers:        signers,
		SignersByEpoch: signersByEpoch,
	}
}

// DefaultGenesisState returns the default genesis state for the module.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(Params{
		QuorumSize:    1024,
		TokensPerVote: "100",
		MaxVotes:      100,
		EpochBlocks:   1000,
	}, 0, make([]*Signer, 0), make([]*EpochSignerSet, 0))
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
	if len(gs.SignersByEpoch) != int(gs.EpochNumber) {
		return fmt.Errorf("epoch history missing")
	}
	for _, signers := range gs.SignersByEpoch {
		for _, signer := range signers.Signers {
			if err := ValidateHexAddress(signer); err != nil {
				return err
			}
			if _, ok := registered[signer]; !ok {
				return fmt.Errorf("historical signer detail missing")
			}
		}
	}
	return nil
}
