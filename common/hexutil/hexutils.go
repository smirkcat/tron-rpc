package hexutil

import (
	"encoding/hex"
	"strings"
)

var (
	EmptyString = &hexError{"empty hex string"}
)

type hexError struct {
	msg string
}

func (h *hexError) Error() string {
	return h.msg
}

// Encode encodes bytes as a hex string.
func Encode(bytes []byte) string {
	encode := make([]byte, len(bytes)*2)
	hex.Encode(encode, bytes)
	return string(encode)
}

// Decode hex string as bytes
func Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, EmptyString
	}
	return hex.DecodeString(input[:])
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) ([]byte, error) {
	return hex.DecodeString(str)
}

// BytesToHexString encodes bytes as a hex string.
func BytesToHexString(bytes []byte) string {
	encode := make([]byte, len(bytes)*2)
	hex.Encode(encode, bytes)
	return "0x" + string(encode)
}

// HexStringToBytes hex string as bytes
func HexStringToBytes(input string) ([]byte, error) {
	if len(input) == 0 {
		return nil, EmptyString
	}

	return hex.DecodeString(strings.Replace(input, "0x", "", -1))
}
