/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/gin-middleware"
	"github.com/SENERGY-Platform/go-cc-job-handler/ccjh"
	"github.com/SENERGY-Platform/go-service-base/job-hdl"
	sb_logger "github.com/SENERGY-Platform/go-service-base/logger"
	"github.com/SENERGY-Platform/go-service-base/srv-info-hdl"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/go-service-base/watchdog"
	cew_client "github.com/SENERGY-Platform/mgw-container-engine-wrapper/client"
	"github.com/SENERGY-Platform/mgw-core-manager/api"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/cleanup_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/http_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/kratos_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/log_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/nginx_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/handler/service_hdl"
	"github.com/SENERGY-Platform/mgw-core-manager/lib/model"
	"github.com/SENERGY-Platform/mgw-core-manager/util"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

var version string

var endpointTemplates = map[int]string{
	nginx_hdl.StandardLocationTmpl:    "/endpoints/deployment/{ref}/{path}",
	nginx_hdl.StandardRewriteTmpl:     "/endpoints/deployment/{ref}/{path}(.*) /$1 break",
	nginx_hdl.StandardProxyPassTmpl:   "http://{var}{port}{path}$1$is_args$args",
	nginx_hdl.DefaultGuiLocationTmpl:  "/",
	nginx_hdl.DefaultGuiProxyPassTmpl: "http://{var}{port}{path}",
	nginx_hdl.AliasLocationTmpl:       "/endpoints/alias/{path}",
	nginx_hdl.AliasRewriteTmpl:        "/endpoints/alias/{path}(.*) /$1 break",
	nginx_hdl.AliasProxyPassTmpl:      "http://{var}{port}{path}$1$is_args$args",
}

func main() {
	srvInfoHdl := srv_info_hdl.New("core-manager", version)

	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	util.ParseFlags()

	config, err := util.NewConfig(util.Flags.ConfPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		ec = 1
		return
	}

	logFile, err := util.InitLogger(config.Logger)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		var logFileError *sb_logger.LogFileError
		if errors.As(err, &logFileError) {
			ec = 1
			return
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}

	util.Logger.Printf("%s %s", srvInfoHdl.GetName(), srvInfoHdl.GetVersion())

	util.Logger.Debugf("config: %s", sb_util.ToJsonStr(config))

	watchdog.Logger = util.Logger
	wtchdg := watchdog.New(syscall.SIGINT, syscall.SIGTERM)

	kratosCtx, kratosCf := context.WithCancel(context.Background())
	kratosHdl, err := kratos_hdl.New(kratosCtx, config.Kratos.Version, config.Kratos.ConfigPath, config.Kratos.SecretLength, time.Duration(config.Kratos.SecretMaxAge), time.Duration(config.Kratos.Interval))
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}
	if err = kratosHdl.Init(); err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", config.HttpClient.CewSocketPath)
			},
		},
	}

	cewClient := cew_client.New(httpClient, "http://unix")

	coreServiceHdl := service_hdl.New(cewClient, config.CoreID, time.Duration(config.HttpClient.Timeout))
	if err = coreServiceHdl.Init(config.ComposeFilePath); err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	gwCtrHdl, err := coreServiceHdl.GetCtrHandler(config.CoreService.GatewaySrvName)
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	gwEndpointHdl := nginx_hdl.New(gwCtrHdl, config.EndpointsConfPath, endpointTemplates)
	if err = gwEndpointHdl.Init(); err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	logConfig, err := log_hdl.ReadConfig(config.LogHandler.Path)
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}
	logHdl, err := log_hdl.New(logConfig, config.LogHandler.BufferSize)
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	cleanupHdl := cleanup_hdl.New(cewClient, time.Duration(config.HttpClient.Timeout))

	ccHandler := ccjh.New(config.Jobs.BufferSize)

	job_hdl.Logger = util.Logger
	job_hdl.ErrCodeMapper = util.GetErrCode
	job_hdl.NewNotFoundErr = model.NewNotFoundError
	job_hdl.NewInvalidInputError = model.NewInvalidInputError
	job_hdl.NewInternalErr = model.NewInternalError
	jobCtx, jobCF := context.WithCancel(context.Background())
	jobHandler := job_hdl.New(jobCtx, ccHandler)
	purgeJobsHdl := job_hdl.NewPurgeJobsHandler(jobHandler, time.Duration(config.Jobs.PJHInterval), time.Duration(config.Jobs.MaxAge))

	wtchdg.RegisterStopFunc(func() error {
		ccHandler.Stop()
		jobCF()
		if ccHandler.Active() > 0 {
			util.Logger.Info("waiting for active jobs to cancel ...")
			ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
			defer cf()
			for ccHandler.Active() != 0 {
				select {
				case <-ctx.Done():
					return fmt.Errorf("canceling jobs took too long")
				default:
					time.Sleep(50 * time.Millisecond)
				}
			}
			util.Logger.Info("jobs canceled")
		}
		return nil
	})

	gin.SetMode(gin.ReleaseMode)
	httpHandler := gin.New()
	staticHeader := map[string]string{
		model.HeaderApiVer:  srvInfoHdl.GetVersion(),
		model.HeaderSrvName: srvInfoHdl.GetName(),
	}
	httpHandler.Use(gin_mw.StaticHeaderHandler(staticHeader), requestid.New(requestid.WithCustomHeaderStrKey(model.HeaderRequestID)), gin_mw.LoggerHandler(util.Logger, nil, func(gc *gin.Context) string {
		return requestid.Get(gc)
	}), gin_mw.ErrorHandler(util.GetStatusCode, ", "), gin.Recovery())
	httpHandler.UseRawPath = true
	cmApi := api.New(coreServiceHdl, gwEndpointHdl, cleanupHdl, logHdl, jobHandler, srvInfoHdl)

	http_hdl.SetRoutes(httpHandler, cmApi)
	util.Logger.Debugf("routes: %s", sb_util.ToJsonStr(http_hdl.GetRoutes(httpHandler)))

	listener, err := sb_util.NewUnixListener(config.Socket.Path, os.Getuid(), config.Socket.GroupID, config.Socket.FileMode)
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}
	server := &http.Server{Handler: httpHandler}
	srvCtx, srvCF := context.WithCancel(context.Background())
	wtchdg.RegisterStopFunc(func() error {
		if srvCtx.Err() == nil {
			ctxWt, cf := context.WithTimeout(context.Background(), time.Second*5)
			defer cf()
			if err := server.Shutdown(ctxWt); err != nil {
				return err
			}
			util.Logger.Info("http server shutdown complete")
		}
		return nil
	})
	wtchdg.RegisterHealthFunc(func() bool {
		if srvCtx.Err() == nil {
			return true
		}
		util.Logger.Error("http server closed unexpectedly")
		return false
	})

	wtchdg.RegisterHealthFunc(kratosHdl.Running)
	wtchdg.RegisterStopFunc(func() error {
		kratosCf()
		kratosHdl.Wait()
		return nil
	})

	kratosHdl.Start()

	wtchdg.Start()

	err = ccHandler.RunAsync(config.Jobs.MaxNumber, time.Duration(config.Jobs.JHInterval*1000))
	if err != nil {
		util.Logger.Error(err)
		ec = 1
		return
	}

	purgeJobsHdl.Start(jobCtx)

	if err = cmApi.PurgeCoreImages(time.Duration(config.ImgPurgeDelay)); err != nil {
		util.Logger.Error(err)
	}

	go func() {
		defer srvCF()
		util.Logger.Info("starting http server ...")
		if err := server.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			util.Logger.Error(err)
			ec = 1
			return
		}
	}()

	ec = wtchdg.Join()
}
