package cache

import (
	"os"
)

type FileCache struct {
	basePath string
}

func NewFileCache(basePath string) *FileCache {
	return &FileCache{
		basePath: basePath,
	}
}

func (f *FileCache) Get(key string) ([]byte, error) {
	file, err := os.Open(f.basePath + key)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, stat.Size())
	_, err = file.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (f *FileCache) Set(key string, value []byte) error {
	err := os.MkdirAll(f.basePath, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(f.basePath + key)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(value)
	if err != nil {
		return err
	}

	return nil
}
