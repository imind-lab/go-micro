package template

import (
	"os"
	"text/template"
)

// 生成server
func CreateServer(data *Data) error {
	var tpl = `/**
 *  MindLab
 *
 *  Create by songli on {{.Date}}
 *  Copyright © {{.Year}} imind.tech All rights reserved.
 */

package server

import (
	"fmt"

	httpx "github.com/asim/go-micro/plugins/server/http/v4"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"

	"{{.Domain}}/{{.Project}}/{{.Service}}/application/handler"
	"{{.Domain}}/{{.Project}}/{{.Service}}/application/proto/{{.Service}}"
	"{{.Domain}}/{{.Project}}/{{.Service}}/gateway"
)

var (
	service = "{{.Service}}"
	version = "latest"
)

func init() {
	cfg := "./infrastructure/conf/conf.yaml"
	viper.SetConfigFile(cfg)
	//初始化全部的配置
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func Serve() error {
	svc := {{.Project}}.NewService(
		{{.Project}}.Name(service),
		{{.Project}}.Version(version),
	)

	svc.Init()

	{{.Service}}.Register{{.Svc}}ServiceHandler(svc.Server(), handler.New{{.Svc}}Service(svc.Client()))

	address := fmt.Sprintf(":%d", viper.GetInt("service.port.http"))
	srv := httpx.NewServer(
		server.Name(service),
		server.Address(address),
	)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	gw := gateway.NewGateway(svc)
	gw.InitRouter(router)

	hd := srv.NewHandler(router)
	if err := srv.Handle(hd); err != nil {
		logger.Errorf("http handle error:%+v", err)
		return err
	}

	// 启动http监听
	if err := srv.Start(); err != nil {
		logger.Errorf("start http server error:%+v", err)
		return err
	}

	// 启动grpc监听
	if err := svc.Run(); err != nil {
		logger.Errorf("service grpc server error:%+v", err)
		return err
	}

	return nil
}
`

	t, err := template.New("repository").Parse(tpl)
	if err != nil {
		return err
	}

	t.Option()
	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/infrastructure/server/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + "server.go"

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
