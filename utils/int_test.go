package utils

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalAndUnmarshalInt(t *testing.T) {
	i := big.NewInt(3)
	m, err := MarshalBigInt(i)
	require.NoError(t, err)

	i2, err := UnmarshalBigInt(m)
	require.NoError(t, err)
	require.Equal(t, i, i2)
}
