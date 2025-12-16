package audit

type Logger struct {
	Enabled bool
	Secret  []byte
}
