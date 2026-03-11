package auth_test

import (
	"testing"

	"github.com/javier/api-task-user/internal/adapter/outbound/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptAdapter_HashAndCompare_Success(t *testing.T) {
	adapter := auth.NewBcryptAdapter(4) // low cost for fast tests

	hash, err := adapter.Hash("mypassword123")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "mypassword123", hash)

	err = adapter.Compare(hash, "mypassword123")
	assert.NoError(t, err)
}

func TestBcryptAdapter_Compare_WrongPassword(t *testing.T) {
	adapter := auth.NewBcryptAdapter(4)

	hash, err := adapter.Hash("correct-password")
	require.NoError(t, err)

	err = adapter.Compare(hash, "wrong-password")
	assert.Error(t, err)
}

func TestBcryptAdapter_Hash_DifferentOutputs(t *testing.T) {
	adapter := auth.NewBcryptAdapter(4)

	hash1, err := adapter.Hash("same-input")
	require.NoError(t, err)

	hash2, err := adapter.Hash("same-input")
	require.NoError(t, err)

	// bcrypt produces different hashes due to random salt
	assert.NotEqual(t, hash1, hash2)

	// Both should still validate correctly
	assert.NoError(t, adapter.Compare(hash1, "same-input"))
	assert.NoError(t, adapter.Compare(hash2, "same-input"))
}

func TestBcryptAdapter_InvalidCostDefaultsToDefault(t *testing.T) {
	// Cost 0 is below min, should default to bcrypt.DefaultCost
	adapter := auth.NewBcryptAdapter(0)

	hash, err := adapter.Hash("test")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}
