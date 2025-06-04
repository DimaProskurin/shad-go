//go:build !solution

package otp

import (
	"errors"
	"fmt"
	"io"
)

type CipherDecoder struct {
	encodedStream io.Reader
	keyStream     io.Reader
}

func (c *CipherDecoder) Read(p []byte) (int, error) {
	readN, err := c.encodedStream.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		return readN, err
	}

	keyBlock := make([]byte, readN)
	curReadKeyN := 0
	for curReadKeyN < readN {
		readKeyN, err := c.keyStream.Read(keyBlock[curReadKeyN:])
		if err != nil {
			return readN, err
		}
		curReadKeyN += readKeyN
	}

	decoded, errXor := xorBlocks(p[:readN], keyBlock)
	if errXor != nil {
		return readN, errXor
	}
	copy(p[:readN], decoded)
	return readN, err
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &CipherDecoder{
		encodedStream: r,
		keyStream:     prng,
	}
}

type CipherEncoder struct {
	encodedStream io.Writer
	keyStream     io.Reader
}

func (c *CipherEncoder) Write(p []byte) (int, error) {
	keyBlock := make([]byte, len(p))
	curReadKeyN := 0
	for curReadKeyN < len(p) {
		readKeyN, err := c.keyStream.Read(keyBlock)
		if err != nil {
			return 0, err
		}
		curReadKeyN += readKeyN
	}

	encoded, errXor := xorBlocks(p, keyBlock)
	if errXor != nil {
		return 0, errXor
	}

	return c.encodedStream.Write(encoded)
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &CipherEncoder{
		encodedStream: w,
		keyStream:     prng,
	}
}

func xorBlocks(a []byte, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("xorBlocks: a and b different lens")
	}
	res := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		res[i] = a[i] ^ b[i]
	}
	return res, nil
}
