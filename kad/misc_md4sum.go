package kad

import (
	"golang.org/x/crypto/md4"
)

// Md4Sum x
type Md4Sum struct {
	rawHash []byte
}

func (m *Md4Sum) calculate(data []byte) {
	ctx := md4.New()
	ctx.Write(data)
	m.rawHash = ctx.Sum(nil)
}

func (m *Md4Sum) getRawHash() []byte {
	return m.rawHash
}
