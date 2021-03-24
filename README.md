# Azure Docker Volume plugin

Current implementation is based on following specifications and guideline:
- [Docker volume plugins](https://docs.docker.com/engine/extend/plugins_volume/)
- [Official azfile Go package to interact with Azure File Share API](https://pkg.go.dev/github.com/Azure/azure-storage-file-go/azfile#example-package)
- [Use Azure Files with Linux](https://docs.microsoft.com/en-us/azure/storage/files/storage-how-to-use-files-linux)