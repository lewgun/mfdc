package uuid

import (
	"encoding/hex"
	"errors"
	"time"
)

const (
	workerIDBits   = uint64(10)
	sequenceBits   = uint64(12)
	workerIDShift  = sequenceBits
	timestampShift = sequenceBits + workerIDBits
	sequenceMask   = int64(-1) ^ (int64(-1) << sequenceBits)

	// Tue, 21 Mar 2006 20:50:14.000 GMT
	twepoch = int64(1288834974657)

	snowFlakeLength = 16
)

var (
	ErrTimeBackwards   = errors.New("time has gone backwards")
	ErrSequenceExpired = errors.New("sequence expired")
	factory            = &guidFactory{}
)

type guid int64
type guidFactory struct {
	sequence      int64
	lastTimestamp int64
}

func (f *guidFactory) NewGUID(workerID int64) (guid, error) {
	ts := time.Now().UnixNano() / 1e6

	if ts < f.lastTimestamp {
		return 0, ErrTimeBackwards
	}

	if f.lastTimestamp == ts {
		f.sequence = (f.sequence + 1) & sequenceMask
		if f.sequence == 0 {
			return 0, ErrSequenceExpired
		}
	} else {
		f.sequence = 0
	}

	f.lastTimestamp = ts

	id := ((ts - twepoch) << timestampShift) |
		(workerID << workerIDShift) |
		f.sequence

	return guid(id), nil
}

func (g guid) Hex() []byte {
	var h [snowFlakeLength]byte
	var b [8]byte

	b[0] = byte(g >> 56)
	b[1] = byte(g >> 48)
	b[2] = byte(g >> 40)
	b[3] = byte(g >> 32)
	b[4] = byte(g >> 24)
	b[5] = byte(g >> 16)
	b[6] = byte(g >> 8)
	b[7] = byte(g)

	hex.Encode(h[:], b[:])
	return h[:]
}

func (g guid) HexString() string {
	return string(g.Hex())
}

func New() string {

	g, _ := factory.NewGUID(0)
	return g.HexString()

}
