package domain

const (
	StatusInSync       = "in_sync"
	StatusDrifted      = "drifted"
	StatusMissingRemote = "missing_remote"
	StatusMissingDesired = "missing_desired"
	StatusError        = "error"
)

const (
	ActionResultSuccess = "success"
	ActionResultFailed  = "failed"
	ActionResultPartial = "partial"
)
