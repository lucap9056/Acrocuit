package models

import (
	"strings"
	"sync"
)

var builderPool = sync.Pool{
	New: func() any {
		b := &strings.Builder{}
		b.Grow(128)
		return b
	},
}

type bitmask uint64

func (b *bitmask) add(bit uint8) bool {
	mask := bitmask(1) << bit
	if *b&mask == 0 {
		*b |= mask
		return true
	}
	return false
}

type columnCollector struct {
	builder *strings.Builder
	fields  []uint8
	seen    bitmask
}

func newColumnCollector(columnsStr string) *columnCollector {
	maxCols := strings.Count(columnsStr, ",") + 1

	builder := builderPool.Get().(*strings.Builder)
	builder.Grow(maxCols * 16)

	return &columnCollector{
		builder: builder,
		fields:  make([]uint8, 0, maxCols),
	}
}

func (c *columnCollector) Append(bit uint8, name string) {
	if !c.seen.add(bit) {
		return
	}
	if c.builder.Len() > 0 {
		c.builder.WriteByte(',')
	}
	c.builder.WriteString(name)
	c.fields = append(c.fields, bit)
}

func (c *columnCollector) Release() {
	if c.builder.Cap() <= 2048 {
		c.builder.Reset()
		builderPool.Put(c.builder)
	}
}

func (c *columnCollector) Result() (string, []uint8) {
	return c.builder.String(), c.fields
}
