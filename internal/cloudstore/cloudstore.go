package cloudstore

import "context"

// idFunc is the function that gets the ID that a new file or directory
// should be written under. The function is passed a fallback ID (r) incase
// the parent directory cannot be determined. For most cases, the fallback
// ID (r) will be the users root directory ID.
type idFunc func(r string) (string, error)

// UserValidatorFunc validates that a users directory structure exists
// on the server. It should take a user ID and verify the storage
// mechanism is set up for the user. If it is not, it should set up the
// users storage.
type UserValidatorFunc func(ctx context.Context, userID string) (Dir, error)
