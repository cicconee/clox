package cloudstore

// idFunc is the function that gets the ID that a new file or directory
// should be written under. The function is passed a fallback ID (r) incase
// the parent directory cannot be determined. For most cases, the fallback
// ID (r) will be the users root directory ID.
type idFunc func(r string) (string, error)
