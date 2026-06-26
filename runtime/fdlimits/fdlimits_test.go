package fdlimits_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/fdlimits"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	silaLimit "github.com/sila-chain/Sila/common/fdlimit"
)

func TestSetMaxFdLimits(t *testing.T) {
	assert.NoError(t, fdlimits.SetMaxFdLimits())

	curr, err := silaLimit.Current()
	assert.NoError(t, err)

	max, err := silaLimit.Maximum()
	assert.NoError(t, err)

	assert.Equal(t, max, curr, "current and maximum file descriptor limits do not match up.")

}
