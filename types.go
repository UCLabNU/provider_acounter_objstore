package main

import (
	pcounter "github.com/synerex/proto_pcounter"
)

// ACBlock : type definition of ACBlock
type ACBlock struct {
	BaseDate  int64
	PrevLen   uint32
	ACounters []*pcounter.ACounters
}

// IACBlock : type definition of IACBlock
type IACBlock struct {
	BaseDate int64
	PrevLen  uint32
	ACounter []*pcounter.ACounter
}
