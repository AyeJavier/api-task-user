// Package model contains the core domain entities and value objects.
package model

// Profile represents the role assigned to a user in the system.
type Profile string

const (
	// ProfileAdmin can manage users and tasks.
	ProfileAdmin Profile = "ADMIN"
	// ProfileExecutor can view and update their assigned tasks.
	ProfileExecutor Profile = "EXECUTOR"
	// ProfileAuditor has read-only access to all tasks.
	ProfileAuditor Profile = "AUDITOR"
)

// IsValid reports whether the profile is a recognized value.
func (p Profile) IsValid() bool {
	switch p {
	case ProfileAdmin, ProfileExecutor, ProfileAuditor:
		return true
	}
	return false
}

// CanBeCreatedBy reports whether an Admin is allowed to create a user with this profile.
// Admins cannot create other Admins.
func (p Profile) CanBeCreatedBy(creator Profile) bool {
	if creator != ProfileAdmin {
		return false
	}
	return p == ProfileExecutor || p == ProfileAuditor
}
