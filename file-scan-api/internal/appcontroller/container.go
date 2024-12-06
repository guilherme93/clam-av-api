package appcontroller

import (
	"file-scan-api/internal/clamav"
	"file-scan-api/internal/config"
)

type ServiceContainer struct {
	ClamService clamav.Service
	Cfg         *config.Config
}
