package benchmark

import _ "embed"

//go:embed docs/benchmarking/templates/run/run.json
var runTemplate []byte

// RunTemplate returns a fresh copy of the run record template.
func RunTemplate() []byte { return append([]byte(nil), runTemplate...) }
