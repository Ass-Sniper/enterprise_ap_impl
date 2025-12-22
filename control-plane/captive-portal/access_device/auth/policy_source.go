package auth

type PolicySource string

const (
	PolicyFromOverride PolicySource = "override"
	PolicyFromRadius   PolicySource = "radius"
	PolicyFromDefault  PolicySource = "default"
)
