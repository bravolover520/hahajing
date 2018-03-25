package kad

import (
	"crypto/md5"
)

// Md5Sum x
type Md5Sum struct {
	rawHash []byte
}

func (m *Md5Sum) calculate(data []byte) {
	ctx := md5.New()
	ctx.Write(data)
	m.rawHash = ctx.Sum(nil)
}

func (m *Md5Sum) getRawHash() []byte {
	return m.rawHash
}
