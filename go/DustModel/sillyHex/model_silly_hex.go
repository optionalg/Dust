package sillyHex

import (
	"math/rand"
	"time"

	"github.com/blanu/Dust/go/Dust"
)

type sillyHexCodec struct {
	topNybble int8
	lenNext, lenMin, lenStep, lenMax uint16
	sleepNext, sleepMin, sleepStep, sleepMax time.Duration
}

var (
	hexAlphabet = []byte("0123456789abcdef")
)

type sillyHexModel struct {
}

func newSillyHexCodec() Dust.ShapingCodec {
	return &sillyHexCodec{
		topNybble: -1,
		lenNext: 200,
		lenMin: 200,
		lenStep: 200,
		lenMax: 1000,
		sleepNext: 20 * time.Millisecond,
		sleepMin: 20 * time.Millisecond,
		sleepStep: 5 * time.Millisecond,
		sleepMax: 50 * time.Millisecond,
	}
}

func (model *sillyHexModel) MakeClientPair() (Dust.ShapingEncoder, Dust.ShapingDecoder, error) {
	return newSillyHexCodec(), newSillyHexCodec(), nil
}

func (model *sillyHexModel) MakeServerPair() (Dust.ShapingEncoder, Dust.ShapingDecoder, error) {
	return newSillyHexCodec(), newSillyHexCodec(), nil
}

func makeSillyHexModel(params map[string]string) (Dust.ShapingModel, error) {
	for key, _ := range params {
		return nil, &Dust.ParameterError{Dust.ParameterUnexpected, "parameter", key}
	}
	
	return &sillyHexModel{}, nil
}

func init() {
	Dust.RegisterModel("sillyHex", makeSillyHexModel)
}

func (hex *sillyHexCodec) WholeStreamDuration() time.Duration {
	return 1000 * time.Hour
}

func (hex *sillyHexCodec) MaxPacketLength() uint16 {
	return hex.lenMax
}

func (hex *sillyHexCodec) NextPacketLength() uint16 {
	hex.lenNext += hex.lenStep
	if hex.lenNext > hex.lenMax {
		hex.lenNext = hex.lenMin
	}

	return hex.lenNext
}

func (hex *sillyHexCodec) NextPacketSleep() time.Duration {
	hex.sleepNext += hex.sleepStep
	if hex.sleepNext > hex.sleepMax {
		hex.sleepNext = hex.sleepMin
	}

	return hex.sleepNext
}

func (hex *sillyHexCodec) ShapeBytes(p []byte) []byte {
	out := make([]byte, len(p) * 4)
	j := 0
	pattern := uint64(rand.Int63())

	maybeIntersperse := func() {
		if pattern & 1 != 0 {
			out[j] = '.'
			j++
		}
		pattern = pattern >> 1 | pattern << 63
	}

	for _, byte := range p {
		out[j] = hexAlphabet[byte >> 4]
		j++
		maybeIntersperse()
		out[j] = hexAlphabet[byte & 0xf]
		j++
		maybeIntersperse()
	}

	return out[:j]
}

func (hex *sillyHexCodec) UnshapeBytes(p []byte) []byte {
	out := make([]byte, 1 + (len(p) / 2))
	j := 0

	for _, byte := range p {
		var nybble int8 = -1
		
		if '0' <= byte && byte <= '9' {
			nybble = int8(byte - '0')
		} else if 'a' <= byte && byte <= 'f' {
			nybble = int8(10 + (byte - 'a'))
		} else if 'A' <= byte && byte <= 'F' {
			nybble = int8(10 + (byte - 'A'))
		}

		if nybble < 0 {
			continue
		}
		
		if hex.topNybble < 0 {
			hex.topNybble = nybble
		} else {
			out[j] = uint8(hex.topNybble) << 4 | uint8(nybble)
			hex.topNybble = -1
			j++
		}
	}

	return out[:j]
}