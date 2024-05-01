package types_test

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/0glabs/0g-chain/app"
	types "github.com/0glabs/0g-chain/x/committee/types"
	pricefeedtypes "github.com/0glabs/0g-chain/x/pricefeed/types"
)

type ParamsChangeTestSuite struct {
	suite.Suite

	ctx sdk.Context
	pk  types.ParamKeeper

	cdpCollateralRequirements []types.SubparamRequirement
}

func (suite *ParamsChangeTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

	suite.ctx = ctx
	suite.pk = tApp.GetParamsKeeper()

	suite.cdpCollateralRequirements = []types.SubparamRequirement{
		{
			Key:                        "type",
			Val:                        "bnb-a",
			AllowedSubparamAttrChanges: []string{"conversion_factor", "liquidation_ratio", "spot_market_id"},
		},
		{
			Key:                        "type",
			Val:                        "btc-a",
			AllowedSubparamAttrChanges: []string{"stability_fee", "debt_limit", "auction_size", "keeper_reward_percentage"},
		},
	}
}

// Test subparam value with slice data unchanged comparision
func (s *ParamsChangeTestSuite) TestParamsChangePermission_SliceSubparamComparision() {
	permission := types.ParamsChangePermission{
		AllowedParamsChanges: types.AllowedParamsChanges{{
			Subspace: pricefeedtypes.ModuleName,
			Key:      string(pricefeedtypes.KeyMarkets),
			MultiSubparamsRequirements: []types.SubparamRequirement{
				{
					Key:                        "market_id",
					Val:                        "xrp:usd",
					AllowedSubparamAttrChanges: []string{"quote_asset", "oracles"},
				},
				{
					Key:                        "market_id",
					Val:                        "btc:usd",
					AllowedSubparamAttrChanges: []string{"active"},
				},
			},
		}},
	}
	_, oracles := app.GeneratePrivKeyAddressPairs(5)

	testcases := []struct {
		name     string
		expected bool
		value    string
	}{
		{
			name:     "success changing allowed attrs",
			expected: true,
			value: fmt.Sprintf(`[{
				"market_id": "xrp:usd",
				"base_asset": "xrp",
				"quote_asset": "usdx",
				"oracles": [],
				"active": true
			},
			{
				"market_id": "btc:usd",
				"base_asset": "btc",
				"quote_asset": "usd",
				"oracles": ["%s"],
				"active": false
			}]`, oracles[1].String()),
		},
		{
			name:     "fails when changing not allowed attr (oracles)",
			expected: false,
			value: fmt.Sprintf(`[{
				"market_id": "xrp:usd",
				"base_asset": "xrp",
				"quote_asset": "usdx",
				"oracles": ["%s"],
				"active": true
			},
			{
				"market_id": "btc:usd",
				"base_asset": "btc",
				"quote_asset": "usd",
				"oracles": ["%s"],
				"active": false
			}]`, oracles[0].String(), oracles[2].String()),
		},
	}
	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()

			subspace, found := s.pk.GetSubspace(pricefeedtypes.ModuleName)
			s.Require().True(found)
			currentMs := pricefeedtypes.Markets{
				{MarketID: "xrp:usd", BaseAsset: "xrp", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracles[0]}, Active: true},
				{MarketID: "btc:usd", BaseAsset: "btc", QuoteAsset: "usd", Oracles: []sdk.AccAddress{oracles[1]}, Active: true},
			}
			subspace.Set(s.ctx, pricefeedtypes.KeyMarkets, &currentMs)

			proposal := paramsproposal.NewParameterChangeProposal(
				"A Title",
				"A description of this proposal.",
				[]paramsproposal.ParamChange{{
					Subspace: pricefeedtypes.ModuleName,
					Key:      string(pricefeedtypes.KeyMarkets),
					Value:    tc.value,
				}},
			)
			s.Require().Equal(
				tc.expected,
				permission.Allows(s.ctx, s.pk, proposal),
			)
		})
	}
}

func TestParamsChangeTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsChangeTestSuite))
}

func TestAllowedParamsChanges_Get(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name  string
		apcs  types.AllowedParamsChanges
		args  args
		found bool
		out   types.AllowedParamsChange
	}{
		{
			name: "when element exists it is found",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceA",
				key:      "key2",
			},
			found: true,
			out:   exampleAPCs[1],
		},
		{
			name: "when element doesn't exist it isn't found",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceB",
				key:      "key1",
			},
			found: false,
		},
		{
			name: "when slice is nil, no elements are found",
			apcs: nil,
			args: args{
				subspace: "",
				key:      "",
			},
			found: false,
		},
		{
			name: "when slice is empty, no elements are found",
			apcs: types.AllowedParamsChanges{},
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			found: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, found := tc.apcs.Get(tc.args.subspace, tc.args.key)
			require.Equal(t, tc.found, found)
			require.Equal(t, tc.out, out)
		})
	}
}

func TestAllowedParamsChanges_Set(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name string
		apcs types.AllowedParamsChanges
		arg  types.AllowedParamsChange
		out  types.AllowedParamsChanges
	}{
		{
			name: "when element isn't present it is added",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceB",
				Key:                        "key1",
				SingleSubparamAllowedAttrs: []string{"attribute1"},
			},
			out: append(exampleAPCs, types.AllowedParamsChange{
				Subspace:                   "subspaceB",
				Key:                        "key1",
				SingleSubparamAllowedAttrs: []string{"attribute1"},
			}),
		},
		{
			name: "when element matches, it is overwritten",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when element matches, it is overwritten",
			apcs: exampleAPCs,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when slice is nil, elements are added",
			apcs: nil,
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
		{
			name: "when slice is empty, elements are added",
			apcs: types.AllowedParamsChanges{},
			arg: types.AllowedParamsChange{
				Subspace:                   "subspaceA",
				Key:                        "key2",
				SingleSubparamAllowedAttrs: []string{"attribute3"},
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key2",
					SingleSubparamAllowedAttrs: []string{"attribute3"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			(&tc.apcs).Set(tc.arg)
			require.Equal(t, tc.out, tc.apcs)
		})
	}
}

func TestAllowedParamsChanges_Delete(t *testing.T) {
	exampleAPCs := types.AllowedParamsChanges{
		{
			Subspace:                   "subspaceA",
			Key:                        "key1",
			SingleSubparamAllowedAttrs: []string{"attribute1"},
		},
		{
			Subspace:                   "subspaceA",
			Key:                        "key2",
			SingleSubparamAllowedAttrs: []string{"attribute2"},
		},
	}

	type args struct {
		subspace, key string
	}
	testCases := []struct {
		name string
		apcs types.AllowedParamsChanges
		args args
		out  types.AllowedParamsChanges
	}{
		{
			name: "when element exists it is removed",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceA",
				key:      "key2",
			},
			out: types.AllowedParamsChanges{
				{
					Subspace:                   "subspaceA",
					Key:                        "key1",
					SingleSubparamAllowedAttrs: []string{"attribute1"},
				},
			},
		},
		{
			name: "when element doesn't exist, none are removed",
			apcs: exampleAPCs,
			args: args{
				subspace: "subspaceB",
				key:      "key1",
			},
			out: exampleAPCs,
		},
		{
			name: "when slice is nil, nothing happens",
			apcs: nil,
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			out: nil,
		},
		{
			name: "when slice is empty, nothing happens",
			apcs: types.AllowedParamsChanges{},
			args: args{
				subspace: "subspaceA",
				key:      "key1",
			},
			out: types.AllowedParamsChanges{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			(&tc.apcs).Delete(tc.args.subspace, tc.args.key)
			require.Equal(t, tc.out, tc.apcs)
		})
	}
}
