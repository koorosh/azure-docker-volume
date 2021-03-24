package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-file-go/azfile"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/moby/sys/mountinfo"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

	// TODO: consume share name from driver options
	shareName := "bbbbb"
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

func (v *volumeDriver) List() (resp *volume.ListResponse, err error) {
	v.m.Lock()
	defer v.m.Unlock()
	vols, err := v.meta.List()
	if err != nil {
		return nil, err
	}
	for _, vol := range vols {
		resp.Volumes = append(resp.Volumes, v.volumeEntry(vol))
	}
	return
}

func (v *volumeDriver) Get(req *volume.GetRequest) (resp *volume.GetResponse, err error) {
	v.m.Lock()
	defer v.m.Unlock()
	_, err = v.meta.Get(req.Name)
	if err != nil {
		return
	}
	resp.Volume = v.volumeEntry(req.Name)
	return
}

func (v *volumeDriver) Path(req *volume.PathRequest) (resp *volume.PathResponse, err error) {
	v.m.Lock()
	defer v.m.Unlock()
	resp.Mountpoint = v.pathForVolume(req.Name)
	return
}

func (v *volumeDriver) Mount(req *volume.MountRequest) (resp *volume.MountResponse, err error) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "mount",
		"name":      req.Name,
	})

	path := v.pathForVolume(req.Name)
	if err = os.MkdirAll(path, 0700); err != nil {
		logctx.Error(err)
		return
	}

	meta, err := v.meta.Get(req.Name)
	if err != nil {
		logctx.Error(err)
		return
	}

	if meta.Account != v.accountName {
		err = fmt.Errorf("volume hosted on a different account ('%s') cannot mount", meta.Account)
		logctx.Error(err)
		return
	}

	if err = mount(v.accountName, v.accountKey, path, meta.Options); err != nil {
		logctx.Error(err)
		return
	}
	resp.Mountpoint = path
	return
}

func (v *volumeDriver) Unmount(req *volume.UnmountRequest) error {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "unmount",
		"name":      req.Name,
	})

	path := v.pathForVolume(req.Name)

	if err := unmount(path); err != nil {
		logctx.Error(err)
		return err
	}

	isMounted, err := mountinfo.Mounted(path)
	if err != nil {
		logctx.Error(err)
		return err
	}
	if isMounted {
		logctx.Debug("mountpoint still has active mounts, not removing")
	} else {
		logctx.Debug("mountpoint has no further mounts, removing")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			err = fmt.Errorf("error removing mountpoint: %v", err)
			logctx.Error(err)
			return err
		}
	}
	return nil
}

func (v *volumeDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local",
		},
	}
}

func (v *volumeDriver) volumeEntry(name string) *volume.Volume {
	return &volume.Volume{Name: name,
		Mountpoint: v.pathForVolume(name)}
}

func (v *volumeDriver) pathForVolume(name string) string {
	return filepath.Join(v.mountPoint, name)
}

func mount(accountName, accountKey, mountPath string, options VolumeOptions) error {
	mountURI := fmt.Sprintf("//%s.file.core.windows.net/%s", accountName, options.Share)
	cmd := exec.Command("mount", "-t", "cifs", mountURI, mountPath, "-o", fmt.Sprintf("vers=3.0,username=%s,password=%s,serverino", accountName, accountKey), "--verbose")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %v\noutput=%q", err, out)
	}
	return nil
}

func unmount(mountpoint string) error {
	cmd := exec.Command("umount", mountpoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %v\noutput=%q", err, out)
	}
	return nil
}
