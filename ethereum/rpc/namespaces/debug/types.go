package debug

// TxTraceTask represents a single transaction trace task when an entire block
// is being traced.
type TxTraceTask struct {
	Index   int            // Transaction offset in the block
}

// TxTraceResult is the result of a single transaction trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}
