package main

import (
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

func init() {
	godotenv.Load()
}

func getTestCredentials() (string, string) {
	return os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"), os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
}

func TestCreateShare(t *testing.T) {
	// TODO: remove volume as a cleanup step (with defer)
	// TODO: define separate test env keys for testing purposes only!!!
	accountName, accountKey := getTestCredentials()
	driver, err := newVolumeDriver(accountName, accountKey)
	if err != nil {
		t.Error(err)
	}
	err = driver.Create(&volume.CreateRequest{Name: "testvolume23432r34r3"})
	if err != nil {
		t.Error(err)
	}
	// TODO: validate that volume is created (with not yet implemented API)
}
