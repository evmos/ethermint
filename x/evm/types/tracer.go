package types

import "github.com/ethereum/go-ethereum/core/vm"

const (
	TracerAccessList = "access_list"
	TracerJSON       = "json"
	TracerStruct     = "struct"
	TracerMarkdown   = "markdown"
	TracerMd         = "md"
)

// NewTracer
func NewTracer(tracer string) vm.Tracer {
	switch tracer {
	case TracerAccessList:
		return &vm.AccessListTracer{}
	case TracerJSON:
		return &vm.JSONLogger{}
	case TracerMarkdown, TracerMd:
		return vm.NewMarkdownLogger(nil, nil)
	case TracerStruct:
		return vm.NewStructLogger(nil)
	default:
		return nil
	}
}
