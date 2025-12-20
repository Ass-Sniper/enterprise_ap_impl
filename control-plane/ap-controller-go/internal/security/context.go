package security

// ctxKey is an unexported type to prevent collisions
// with context keys from other packages.
type ctxKey string

// CtxKeyClientMAC is the context key used to store
// authenticated client MAC address.
const CtxKeyClientMAC ctxKey = "portal_client_mac"
