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
	"time"
)

type volumeDriver struct {
	m           sync.Mutex
	su          azfile.ServiceURL
	accountName string
	accountKey  string
	mountPoint  string
	ctx         context.Context
	meta        *metadataDriver
}

func newVolumeDriver(accountName, accountKey, mountPointDir, metaDir string) (*volumeDriver, error) {
	credential, err := azfile.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal(err)
	}
	p := azfile.NewPipeline(credential, azfile.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net", accountName))
	serviceURL := azfile.NewServiceURL(*u, p)
	ctx := context.Background()
	meta, err := newMetadataDriver(metaDir)
	if err != nil {
		return nil, err
	}
	return &volumeDriver{
		su:          serviceURL,
		accountName: accountName,
		accountKey:  accountKey,
		mountPoint:  mountPointDir,
		ctx:         ctx,
		meta:        meta,
	}, nil
}

func (v *volumeDriver) Create(req *volume.CreateRequest) error {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "create",
		"name":      req.Name,
		"options":   req.Options})

	shareName := strings.ToLower(req.Name)
	shareURL := v.su.NewShareURL(shareName)
	// TODO: provide quota limitation in req.Options
	// TODO: handle case when File Share is already created. Do nothing in this case.
	var quotaInGb int32 = 1
	_, err := shareURL.Create(v.ctx, azfile.Metadata{}, quotaInGb)
	if err != nil {
		logctx.Error(err)
		return err
	}
	err = v.meta.Save(shareName, volumeMetadata{
		CreatedAt: time.Now().UTC(),
		Account:   v.accountName,
		Options:   VolumeOptions{},
	})
	if err != nil {
		logctx.Error(err)
		return err
	}
	return nil
}

func (v *volumeDriver) Remove(req *volume.RemoveRequest) error {
	v.m.Lock()
	defer v.m.Unlock()
	shareURL := v.su.NewShareURL(req.Name)
	_, err := shareURL.Delete(v.ctx, azfile.DeleteSnapshotsOptionInclude)
	if err != nil {
		return err
	}
	if err = v.meta.Delete(req.Name); err != nil {
		return err
	}
	return nil
}
