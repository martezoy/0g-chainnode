package types

import (
	"testing"

	"github.com/0glabs/0g-chain/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRawPriceKey_Iteration(t *testing.T) {
	// An iterator key should only match price keys with the same market
	iteratorKey := RawPriceIteratorKey(chaincfg.GasDenom + ":usd")

	addr := sdk.AccAddress("test addr")

	testCases := []struct {
		name      string
		priceKey  []byte
		expectErr bool
	}{
		{
			name:      "equal marketID is included in iteration",
			priceKey:  RawPriceKey(chaincfg.GasDenom+":usd", addr),
			expectErr: false,
		},
		{
			name:      "prefix overlapping marketID excluded from iteration",
			priceKey:  RawPriceKey(chaincfg.GasDenom+":usd:30", addr),
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchedSubKey := tc.priceKey[:len(iteratorKey)]
			if tc.expectErr {
				require.NotEqual(t, iteratorKey, matchedSubKey)
			} else {
				require.Equal(t, iteratorKey, matchedSubKey)
			}
		})
	}
}
