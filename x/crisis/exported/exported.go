package exported

import (
	paramtypes "cosmossdk.io/x/params/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	ParamSet = paramtypes.ParamSet

	// Subspace defines an interface that implements the legacy x/params Subspace
	// type.
	//
	// NOTE: This is used solely for migration of x/params managed parameters.
	Subspace interface {
		Get(ctx sdk.Context, key []byte, ptr interface{})
	}
)
