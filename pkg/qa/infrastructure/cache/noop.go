package cache

import "github.com/openark-net/qa/pkg/qa/domain"

type NoOp struct{}

func (NoOp) Hit(domain.Command) bool           { return false }
func (NoOp) RecordResult(domain.Command, bool) {}
func (NoOp) Flush() error                      { return nil }
