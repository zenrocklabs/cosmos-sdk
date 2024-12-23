package v5fix1

import (
	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	// ParamsKey is the key of x/gov params
	ParamsKey = []byte{0x30}
	// ConstitutionKey is the key of x/gov constitution
	ConstitutionKey = collections.NewPrefix(49)
)

// MigrateStore performs in-place store migrations from v4 (v0.47) to v5 (v0.50). The
// migration includes:
//
// Addition of the new proposal expedited parameters that are set to 0 by default.
// Set of default chain constitution.
func migrateProposals(store storetypes.KVStore, cdc codec.BinaryCodec) error {
	propStore := prefix.NewStore(store, v1.ProposalsKeyPrefix)

	iter := propStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var proposal govv1beta1.Proposal
		err := cdc.Unmarshal(iter.Value(), &proposal)
		if err != nil {
			return err
		}

		if proposal.Content.TypeUrl == "/cosmos.mint.v1beta1.MsgUpdateParams" {
			proposal.Content.TypeUrl = "/zrchain.mint.v1beta1.MsgUpdateParams"
		}
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeService corestoretypes.KVStoreService, cdc codec.BinaryCodec) error {
	store := storeService.OpenKVStore(ctx)

	return migrateProposals(runtime.KVStoreAdapter(store), cdc)
}
