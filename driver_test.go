package main

import (
	"encoding/json"
	"fmt"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/joho/godotenv"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
)

func init() {
	godotenv.Load()
}

func getTestCredentials() (string, string) {
	return os.Getenv("AZURE_TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_TEST_STORAGE_ACCOUNT_KEY")
}

func teardownAllFileShares(driver *volumeDriver, fileShareName string) func() {
	return func() {
		driver.Remove(&volume.RemoveRequest{Name: fileShareName})
	}
}

func TestVolumeDriver_Create(t *testing.T) {
	fileShareName := fmt.Sprintf("testvolume%d", rand.Intn(1000000))
	accountName, accountKey := getTestCredentials()
	mountPointDir := t.TempDir()
	metaDir := t.TempDir()
	driver, err := newVolumeDriver(accountName, accountKey, mountPointDir, metaDir)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(teardownAllFileShares(driver, fileShareName))
	var vol = volume.CreateRequest{Name: fileShareName}
	err = driver.Create(&vol)
	if err != nil {
		t.Error(err)
	}
	b, err := ioutil.ReadFile(path.Join(metaDir, fileShareName))
	if err != nil {
		t.Error(err)
	}
	var v volumeMetadata
	err = json.Unmarshal(b, &v)
	if err != nil {
		t.Error(err)
	}
	if v.Account != accountName {
		t.Errorf("Metadata contains incorrect account name. Got %s, expected: %s", v.Account, accountName)
	}
	// TODO: validate that volume is created (with not yet implemented API)
}

func TestVolumeDriver_Remove(t *testing.T) {
	fileShareName := fmt.Sprintf("testvolume%d", rand.Intn(1000000))
	accountName, accountKey := getTestCredentials()
	mountPointDir := t.TempDir()
	metaDir := t.TempDir()
	// prepare
	driver, err := newVolumeDriver(accountName, accountKey, mountPointDir, metaDir)
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(teardownAllFileShares(driver, fileShareName))
	var vol = volume.CreateRequest{Name: fileShareName}
	err = driver.Create(&vol)
	if err != nil {
		t.Error(err)
	}
	// test
	err = driver.Remove(&volume.RemoveRequest{Name: fileShareName})
	if err != nil {
		t.Error(err)
	}
}
