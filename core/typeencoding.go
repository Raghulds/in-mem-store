package core

import "errors"

func getType(value uint8) uint8 {
	return (value >> 4) << 4
}

func getEncoding(value uint8) uint8 {
	return value & 0b00001111
}

func assertType(value uint8, expected uint8) error {
	if getType(value) != expected {
		return errors.New("type mismatch")
	}
	return nil
}

func assertEncoding(value uint8, expected uint8) error {
	if getEncoding(value) != expected {
		return errors.New("encoding mismatch")
	}
	return nil
}
