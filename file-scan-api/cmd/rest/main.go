package main

import (
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"file-scan-api/internal/appcontroller"
	"file-scan-api/internal/clamav"
	"file-scan-api/internal/config"
	"file-scan-api/internal/http/router"
)

const readHeaderTimeout = time.Second * 2

func main() {
	cfg, err := config.New()
	if err != nil {
		stdlog.Fatalf("unable to load config: %v", err)
	}

	clamService := clamav.NewService(cfg.ClamAV)

	if err = clamService.Wait(); err != nil {
		panic("ClamAV is not ready")
	}

	serviceContainer := appcontroller.ServiceContainer{
		ClamService: clamService,
		Cfg:         cfg,
	}

	newRouter := router.NewRouter(serviceContainer)
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           newRouter,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	if err = httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err.Error())
	}
}
