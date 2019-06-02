package adapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const perm = 0644

type activeChannelStorage struct {
	filename string
}

func newActiveChannelStorage(path string) *activeChannelStorage {
	return &activeChannelStorage{filepath.Join(path, "active")}
}

func (s *activeChannelStorage) store(ch string) error {
	err := ioutil.WriteFile(s.filename, []byte(ch), perm)
	if err != nil {
		return fmt.Errorf("could not write channel file: %v", err)
	}
	return nil
}

func (s *activeChannelStorage) load() (string, error) {
	data, err := ioutil.ReadFile(s.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("could not read channel file: %v", err)
	}

	return string(data), nil
}

func (s *activeChannelStorage) remove() error {
	err := os.Remove(s.filename)
	if err != nil {
		return fmt.Errorf("could not remove channel file: %v", err)
	}
	return nil
}
