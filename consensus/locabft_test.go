package consensus

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/libs/log"
	sm "github.com/cometbft/cometbft/state"
	smmocks "github.com/cometbft/cometbft/state/mocks"
	"github.com/cometbft/cometbft/types"
)

func TestLocabftNeedCommitBlock(t *testing.T) {
	const initialHeight int64 = 10

	tests := []struct {
		name          string
		height        int64
		initialHeight int64
		startup       bool
		meta          *types.BlockMeta
		want          bool
	}{
		{
			name:          "genesis height",
			height:        initialHeight,
			initialHeight: initialHeight,
			want:          true,
		},
		{
			name:          "startup warm-up",
			height:        initialHeight + 1,
			initialHeight: initialHeight,
			want:          true,
		},
		{
			name:          "missing previous block meta",
			height:        initialHeight + 1,
			initialHeight: initialHeight,
			startup:       true,
			want:          true,
		},
		{
			name:          "previous block had transactions",
			height:        initialHeight + 1,
			initialHeight: initialHeight,
			startup:       true,
			meta:          &types.BlockMeta{NumTxs: 1},
			want:          true,
		},
		{
			name:          "previous block was empty",
			height:        initialHeight + 1,
			initialHeight: initialHeight,
			startup:       true,
			meta:          &types.BlockMeta{NumTxs: 0},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockStore := &smmocks.BlockStore{}
			if tt.startup && tt.height != tt.initialHeight {
				if tt.meta == nil {
					blockStore.On("LoadBlockMeta", tt.height-1).Return(nil).Once()
				} else {
					blockStore.On("LoadBlockMeta", tt.height-1).Return(tt.meta).Once()
				}
			}

			cs := &State{
				state:                 sm.State{InitialHeight: tt.initialHeight},
				blockStore:            blockStore,
				startupBlockCommitted: tt.startup,
			}
			cs.Logger = log.NewNopLogger()

			require.Equal(t, tt.want, cs.needCommitBlock(tt.height))
			blockStore.AssertExpectations(t)
		})
	}
}
