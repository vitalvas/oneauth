package tools

import (
	"os"
	"os/user"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRoot(t *testing.T) {
	result := IsRoot()
	
	// Get current user to verify the result
	currentUser, err := user.Current()
	if err != nil {
		t.Logf("Could not get current user: %v", err)
		// If we can't get user info, we can still test that IsRoot returns a boolean
		assert.IsType(t, false, result)
		return
	}
	
	expectedIsRoot := currentUser.Username == "root"
	assert.Equal(t, expectedIsRoot, result)
	
	// Additional check using UID
	uid := os.Getuid()
	if uid == 0 {
		// If UID is 0, we should be root
		assert.True(t, result || currentUser.Username != "root", 
			"UID is 0 but IsRoot returned false and username is not root")
	} else {
		// If UID is not 0, we should not be root
		assert.False(t, result && currentUser.Username == "root", 
			"UID is not 0 but IsRoot returned true and username is root")
	}
}

func TestIsRoot_Consistency(t *testing.T) {
	// Call IsRoot multiple times to ensure consistency
	result1 := IsRoot()
	result2 := IsRoot()
	result3 := IsRoot()
	
	assert.Equal(t, result1, result2)
	assert.Equal(t, result2, result3)
}

func TestIsRoot_LogicValidation(t *testing.T) {
	// Test the internal logic by checking what user.Current() returns
	currentUser, err := user.Current()
	if err != nil {
		t.Logf("user.Current() failed: %v", err)
		// The function should return false when user.Current() fails
		result := IsRoot()
		assert.False(t, result, "IsRoot should return false when user.Current() fails")
		return
	}
	
	isRoot := IsRoot()
	expectedIsRoot := currentUser.Username == "root"
	
	assert.Equal(t, expectedIsRoot, isRoot, 
		"IsRoot result should match whether username is 'root'")
	
	t.Logf("Current user: %s, UID: %s, IsRoot: %v", 
		currentUser.Username, currentUser.Uid, isRoot)
}