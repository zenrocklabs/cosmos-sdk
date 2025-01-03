package v5fix1_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v1gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	v5gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v5fix1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

var voter = sdk.MustAccAddressFromBech32("zen13y3tm68gmu9kntcxwvmue82p6akacnpt2v7nty")

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	propTime := time.Unix(1e9, 0)

	msgUpdateParams := &minttypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: minttypes.Params{
			MintDenom:           "stake",
			InflationRateChange: math.LegacyMustNewDecFromStr("0.1"),
			InflationMax:        math.LegacyMustNewDecFromStr("0.2"),
			InflationMin:        math.LegacyMustNewDecFromStr("0.05"),
			GoalBonded:          math.LegacyMustNewDecFromStr("0.67"),
			BlocksPerYear:       6311520,
		},
	} // fill params as needed
	contentAny, err := types.NewAnyWithValue(msgUpdateParams)
	require.NoError(t, err)

	prop2 := govv1.Proposal{
		Id:       2,
		Messages: []*types.Any{contentAny},
		Status:   govv1.StatusPassed,
		FinalTallyResult: &govv1.TallyResult{
			YesCount:        "1000000000000000000",
			NoCount:         "0",
			AbstainCount:    "0",
			NoWithVetoCount: "0",
		},
		SubmitTime:      &propTime,
		DepositEndTime:  &propTime,
		VotingStartTime: &propTime,
		VotingEndTime:   &propTime,
	}

	// fmt.Println("prop1 typeURL: ")
	// fmt.Println(prop1.Messages[0].TypeUrl)

	msgUpdateGovParams := &govv1.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params: govv1.Params{
			MinDeposit:       []sdk.Coin{sdk.NewCoin("zen", math.NewInt(1000000000000000000))},
			MaxDepositPeriod: new(time.Duration),
			VotingPeriod:     new(time.Duration),
			Quorum:           "0.1",
			Threshold:        "0.2",
			VetoThreshold:    "0.5",
		},
	}

	contentAny2, err := types.NewAnyWithValue(msgUpdateGovParams)
	require.NoError(t, err)

	prop1 := govv1.Proposal{
		Id:       1,
		Messages: []*types.Any{contentAny2},
		Status:   govv1.StatusPassed,
		FinalTallyResult: &govv1.TallyResult{
			YesCount:        "1000000000000000000",
			NoCount:         "0",
			AbstainCount:    "0",
			NoWithVetoCount: "0",
		},
		SubmitTime:      &propTime,
		DepositEndTime:  &propTime,
		VotingStartTime: &propTime,
		VotingEndTime:   &propTime,
	}

	prop1Bz, err := cdc.Marshal(&prop1)
	require.NoError(t, err)
	prop2Bz, err := cdc.Marshal(&prop2)
	require.NoError(t, err)

	store.Set(v1gov.ProposalKey(prop1.Id), prop1Bz)
	store.Set(v1gov.ProposalKey(prop2.Id), prop2Bz)

	// Vote on prop 1
	options := []*govv1.WeightedVoteOption{
		{
			Option: govv1.OptionNo,
			Weight: "0.3",
		},
		{
			Option: govv1.OptionYes,
			Weight: "0.7",
		},
	}
	vote1 := govv1.NewVote(1, voter, options, "test")
	vote1Bz := cdc.MustMarshal(&vote1)
	store.Set(v1gov.VoteKey(1, voter), vote1Bz)

	// Run migrations.
	storeService := runtime.NewKVStoreService(govKey)
	err = v5gov.MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	var newProp1 v1.Proposal
	err = cdc.Unmarshal(store.Get(v1gov.ProposalKey(prop1.Id)), &newProp1)
	require.NoError(t, err)
	compareProps(t, prop1, newProp1)

	var newVote1 v1.Vote
	err = cdc.Unmarshal(store.Get(v1gov.VoteKey(prop1.Id, voter)), &newVote1)
	require.NoError(t, err)
	require.Equal(t, "0.3", newVote1.Options[0].Weight)
	require.Equal(t, "0.7", newVote1.Options[1].Weight)
}

func compareProps(t *testing.T, oldProp v1.Proposal, newProp v1.Proposal) {
	require.Equal(t, oldProp.Id, newProp.Id)
	require.Equal(t, oldProp.TotalDeposit, newProp.TotalDeposit)
	require.Equal(t, oldProp.Status.String(), newProp.Status.String())
	require.Equal(t, oldProp.FinalTallyResult.YesCount, newProp.FinalTallyResult.YesCount)
	require.Equal(t, oldProp.FinalTallyResult.NoCount, newProp.FinalTallyResult.NoCount)
	require.Equal(t, oldProp.FinalTallyResult.NoWithVetoCount, newProp.FinalTallyResult.NoWithVetoCount)
	require.Equal(t, oldProp.FinalTallyResult.AbstainCount, newProp.FinalTallyResult.AbstainCount)
	require.Equal(t, oldProp.Messages[0].TypeUrl, newProp.Messages[0].TypeUrl)
	// Compare UNIX times, as a simple Equal gives difference between Local and
	// UTC times.
	// ref: https://github.com/golang/go/issues/19486#issuecomment-292968278
	require.Equal(t, oldProp.SubmitTime.Unix(), newProp.SubmitTime.Unix())
	require.Equal(t, oldProp.DepositEndTime.Unix(), newProp.DepositEndTime.Unix())
	require.Equal(t, oldProp.VotingStartTime.Unix(), newProp.VotingStartTime.Unix())
	require.Equal(t, oldProp.VotingEndTime.Unix(), newProp.VotingEndTime.Unix())
}
