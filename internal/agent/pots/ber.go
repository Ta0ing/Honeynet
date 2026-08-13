package pots

import (
	"errors"
	"io"
)

const maxBERPayload = 1 << 20

func readBERPacket(reader io.Reader) (byte, []byte, error) {
	var first [2]byte
	if _, err := io.ReadFull(reader, first[:]); err != nil {
		return 0, nil, err
	}
	length := 0
	if first[1]&0x80 == 0 {
		length = int(first[1])
	} else {
		lengthBytes := int(first[1] & 0x7f)
		if lengthBytes < 1 || lengthBytes > 4 {
			return 0, nil, errors.New("invalid BER length")
		}
		raw := make([]byte, lengthBytes)
		if _, err := io.ReadFull(reader, raw); err != nil {
			return 0, nil, err
		}
		for _, value := range raw {
			length = length<<8 | int(value)
		}
	}
	if length < 0 || length > maxBERPayload {
		return 0, nil, errors.New("BER payload is too large")
	}
	payload := make([]byte, length)
	_, err := io.ReadFull(reader, payload)
	return first[0], payload, err
}

func readBERElement(raw []byte) (byte, []byte, []byte, error) {
	if len(raw) < 2 {
		return 0, nil, nil, io.ErrUnexpectedEOF
	}
	length, used, err := parseBERLength(raw[1:])
	if err != nil {
		return 0, nil, nil, err
	}
	offset := 1 + used
	if length > len(raw)-offset {
		return 0, nil, nil, io.ErrUnexpectedEOF
	}
	return raw[0], raw[offset : offset+length], raw[offset+length:], nil
}

func parseBERLength(raw []byte) (int, int, error) {
	if len(raw) < 1 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	if raw[0]&0x80 == 0 {
		return int(raw[0]), 1, nil
	}
	count := int(raw[0] & 0x7f)
	if count < 1 || count > 4 || len(raw) < count+1 {
		return 0, 0, errors.New("invalid BER length")
	}
	length := 0
	for _, value := range raw[1 : count+1] {
		length = length<<8 | int(value)
	}
	if length > maxBERPayload {
		return 0, 0, errors.New("BER payload is too large")
	}
	return length, count + 1, nil
}

func berElement(tag byte, value []byte) []byte {
	result := []byte{tag}
	result = append(result, berLength(len(value))...)
	return append(result, value...)
}

func berLength(length int) []byte {
	if length < 0x80 {
		return []byte{byte(length)}
	}
	if length <= 0xff {
		return []byte{0x81, byte(length)}
	}
	if length <= 0xffff {
		return []byte{0x82, byte(length >> 8), byte(length)}
	}
	return []byte{0x83, byte(length >> 16), byte(length >> 8), byte(length)}
}

func berIntegerValue(value int64) []byte {
	if value == 0 {
		return []byte{0}
	}
	raw := make([]byte, 8)
	for index := 7; index >= 0; index-- {
		raw[index] = byte(value)
		value >>= 8
	}
	start := 0
	for start < len(raw)-1 && raw[start] == 0 && raw[start+1]&0x80 == 0 {
		start++
	}
	return raw[start:]
}

func parseBERInteger(raw []byte) (int64, error) {
	if len(raw) < 1 || len(raw) > 8 {
		return 0, errors.New("invalid BER integer")
	}
	value := int64(0)
	for _, item := range raw {
		value = value<<8 | int64(item)
	}
	return value, nil
}
