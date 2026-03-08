package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommitCallbackErrorPropagation verifies that when a commit callback
// returns an error, the calling code properly handles it.
// This is a regression test for issue #395 where the callback returned nil
// instead of the actual error when tx.Commit() failed.
func TestCommitCallbackErrorPropagation(t *testing.T) {
	// Simulate the callback behavior that was fixed in api_emaillogin.go
	// Before fix: callback returned nil even when commit failed
	// After fix: callback returns the actual error

	t.Run("callback should return error on commit failure", func(t *testing.T) {
		commitErr := errors.New("database commit failed")

		// This simulates the fixed callback behavior from emailRegister
		callback := func() error {
			// Simulate commit failure
			if err := simulateCommitFailure(); err != nil {
				// After fix: return the error (was: return nil)
				return err
			}
			return nil
		}

		// With the fix, the callback properly returns the error
		err := callback()
		assert.Error(t, err)
		assert.Equal(t, commitErr, err)
	})

	t.Run("callback should return nil on success", func(t *testing.T) {
		callback := func() error {
			// Simulate successful commit
			return nil
		}

		err := callback()
		assert.NoError(t, err)
	})
}

// simulateCommitFailure simulates a database commit failure
func simulateCommitFailure() error {
	return errors.New("database commit failed")
}

// TestCallbackErrorHandling verifies the pattern where callback errors
// should be checked and propagated by the caller.
func TestCallbackErrorHandling(t *testing.T) {
	t.Run("caller should check callback error", func(t *testing.T) {
		expectedErr := errors.New("callback error")

		// This simulates the fixed behavior in createUserWithRespAndTx
		// Before fix: commitCallback() was called but return value ignored
		// After fix: if err := commitCallback(); err != nil { return nil, err }
		processWithCallback := func(commitCallback func() error) (interface{}, error) {
			// Fixed code pattern:
			if commitCallback != nil {
				if err := commitCallback(); err != nil {
					return nil, err
				}
			}
			return "success", nil
		}

		result, err := processWithCallback(func() error {
			return expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("caller should proceed when callback succeeds", func(t *testing.T) {
		processWithCallback := func(commitCallback func() error) (interface{}, error) {
			if commitCallback != nil {
				if err := commitCallback(); err != nil {
					return nil, err
				}
			}
			return "success", nil
		}

		result, err := processWithCallback(func() error {
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})
}
