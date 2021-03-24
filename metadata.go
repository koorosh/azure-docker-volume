package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type metadataDriver struct {
	metaDir string
}

type volumeMetadata struct {
	CreatedAt time.Time     `json:"created_at"`
	Account   string        `json:"account"`
	Options   VolumeOptions `json:"options"`
}

// TODO: driver options provided by user during volume creation
type VolumeOptions struct {
	Share string `json:"share"`
}

func newMetadataDriver(metaDir string) (*metadataDriver, error) {
	if err := os.MkdirAll(metaDir, 0700); err != nil {
		return nil, fmt.Errorf("error creating %s: %v", metaDir, err)
	}
	return &metadataDriver{metaDir}, nil
}

func (m *metadataDriver) Save(name string, metadata volumeMetadata) error {
	b, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("cannot serialize metadata: %v", err)
	}
	if err := ioutil.WriteFile(m.path(name), b, 0600); err != nil {
		return fmt.Errorf("cannot write metadata: %v", err)
	}
	return nil
}

func (m *metadataDriver) Load(name string) (volumeMetadata, error) {
	var v volumeMetadata
	f, err := ioutil.ReadFile(m.path(name))
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(f, &v); err != nil {
		return v, err
	}
	return v, nil
}

func (m *metadataDriver) Delete(name string) error {
	return os.Remove(m.path(name))
}

func (m *metadataDriver) List() ([]string, error) {
	var volumes []string

	// return all the file names under metadata directory
	if err := filepath.Walk(m.metaDir, func(path string, info os.FileInfo, inErr error) error {
		if inErr != nil {
			return inErr
		}
		if path == m.metaDir {
			// directory itself, skip
			return nil
		}

		if info.IsDir() { // a directory
			return filepath.SkipDir
		}

		// base file name indicates the volume name
		volumes = append(volumes, filepath.Base(path))
		return nil
	}); err != nil {
		return volumes, fmt.Errorf("cannot list directory: %v", err)
	}
	return volumes, nil
}

func (m *metadataDriver) Get(name string) (volumeMetadata, error) {
	var v volumeMetadata
	b, err := ioutil.ReadFile(m.path(name))
	if err != nil {
		return v, fmt.Errorf("cannot read metadata: %v", err)
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return v, fmt.Errorf("cannot deserialize metadata: %v", err)
	}
	return v, nil
}

func (m *metadataDriver) path(name string) string {
	return filepath.Join(m.metaDir, name)
}
