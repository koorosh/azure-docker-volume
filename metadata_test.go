package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path"
	"testing"
	"time"
)

func TestMetadataDriver_Save(t *testing.T) {
	tmpMetaDir := fmt.Sprintf("%s/testvol_%d", t.TempDir(), rand.Int())
	volumeName := fmt.Sprintf("volname_%d", rand.Int())
	createdAt := time.Now()
	account := fmt.Sprintf("account%d", rand.Int())

	driver, err := newMetadataDriver(tmpMetaDir)
	if err != nil {
		t.Error(err)
	}
	err = driver.Save(volumeName, volumeMetadata{
		CreatedAt: createdAt,
		Account:   account,
		Options:   VolumeOptions{},
	})
	if err != nil {
		t.Error(err)
	}

	f, err := ioutil.ReadFile(path.Join(tmpMetaDir, volumeName))
	if err != nil {
		t.Error(err)
	}
	var meta volumeMetadata
	err = json.Unmarshal(f, &meta)
	if err != nil {
		t.Error(err)
	}
	if meta.Account != account || meta.CreatedAt.Equal(createdAt) == false {
		t.Errorf("saved metadata is corrupted. Expected: %v", meta)
	}
}

func TestMetadataDriver_Load(t *testing.T) {
	tmpMetaDir := fmt.Sprintf("%s/testvol_%d", t.TempDir(), rand.Int())
	volumeName := fmt.Sprintf("volname_%d", rand.Int())
	createdAt := time.Now()
	account := fmt.Sprintf("account%d", rand.Int())

	driver, err := newMetadataDriver(tmpMetaDir)
	if err != nil {
		t.Error(err)
	}
	err = driver.Save(volumeName, volumeMetadata{
		CreatedAt: createdAt,
		Account:   account,
		Options:   VolumeOptions{},
	})
	if err != nil {
		t.Errorf("Could not save test metadata file. %v", err)
	}
	// get a fresh instance of metadataDriver
	driver, err = newMetadataDriver(tmpMetaDir)
	if err != nil {
		t.Error(err)
	}
	meta, err := driver.Load(volumeName)
	if err != nil {
		t.Error(err)
	}
	if meta.Account != account || meta.CreatedAt.Equal(createdAt) == false {
		t.Errorf("loaded metadata is corrupted. Expected: %v", meta)
	}
}
