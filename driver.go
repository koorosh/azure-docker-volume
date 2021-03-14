package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-file-go/azfile"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
	"sync"
)

type volumeDriver struct {
	m           sync.Mutex
	su          azfile.ServiceURL
	accountName string
	accountKey  string
	ctx         context.Context
}

func newVolumeDriver(accountName, accountKey string) (*volumeDriver, error) {
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal(err)
	}
	p := azfile.NewPipeline(credential, azfile.PipelineOptions{})
	// TODO: consume URL format from configuration
	u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net", accountName))
	serviceURL := azfile.NewServiceURL(*u, p)
	ctx := context.Background()

	return &volumeDriver{
		su:          serviceURL,
		accountName: accountName,
		accountKey:  accountKey,
		ctx:         ctx,
	}, nil
}

func (v *volumeDriver) Create(req *volume.CreateRequest) error {
	shareName := strings.ToLower(req.Name)
	shareURL := v.su.NewShareURL(shareName)
	// TODO: provide quota limitation in req.Options
	var quotaInGb int32 = 0
	_, err := shareURL.Create(v.ctx, azfile.Metadata{}, quotaInGb)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
