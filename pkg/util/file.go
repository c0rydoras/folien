package util

import (
	"errors"
	"os"
)

func ReadFile(path string) (string, error) {
	s, err := os.Stat(path)
	if err != nil {
		return "", errors.New("could not read file")
	}
	if s.IsDir() {
		return "", errors.New("can not read directory")
	}

	m := s.Mode()
	if m&os.ModeDevice != 0 {
		if m&os.ModeCharDevice != 0 {
			return "", errors.New("can not read char device")
		}
		return "", errors.New("can not read block device")
	}
	if m&os.ModeNamedPipe != 0 {
		return "", errors.New("can not read pipe")
	}
	if m&os.ModeSocket != 0 {
		return "", errors.New("can not read socket")
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
