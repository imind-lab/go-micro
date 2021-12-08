/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package template

import (
	"os"
	"text/template"
)

// 生成Client
func CreateGateway(data *Data) error {
	// 生成client.go
	var tpl = `/**
 *  MindLab
 *
 *  Create by songli on {{.Date}}
 *  Copyright © {{.Year}} imind.tech All rights reserved.
 */
package gateway

import (
	"{{.Domain}}/{{.Project}}/{{.Service}}/application/proto/{{.Service}}"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/server"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"strconv"
)

type gateway struct {
	s    server.Server
	svc  {{.Service}}.{{.Svc}}Service
	json jsoniter.API
}

func NewGateway(service {{.Project}}.Service) *gateway {
	svc := {{.Service}}.New{{.Svc}}Service("{{.Service}}", service.Client())
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return &gateway{s: service.Server(), svc: svc, json: json}
}

func (m *gateway) InitRouter(router *gin.Engine) {
	router.PUT("/v1/{{.Service}}/create", m.Create{{.Svc}})
	router.GET("/v1/{{.Service}}/one/:id", m.Get{{.Svc}}ById)
	router.GET("/v1/{{.Service}}/list/:status", m.Get{{.Svc}}List)
	router.POST("/v1/{{.Service}}/status", m.Update{{.Svc}}Status)
	router.POST("/v1/{{.Service}}/count", m.Update{{.Svc}}Count)
	router.DELETE("/v1/{{.Service}}/del", m.Delete{{.Svc}}ById)
}

func (m *gateway) Create{{.Svc}}(c *gin.Context) {
	var req {{.Service}}.Create{{.Svc}}Request

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSONP(http.StatusNoContent, "错误请求")
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	resp, err := m.svc.Create{{.Svc}}(c.Request.Context(), &req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)
}

func (m *gateway) Get{{.Svc}}ById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	req := &{{.Service}}.Get{{.Svc}}ByIdRequest{Id: int32(id)}

	resp, err := m.svc.Get{{.Svc}}ById(c.Request.Context(), req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)
}

func (m *gateway) Get{{.Svc}}List(c *gin.Context) {
	status, err := strconv.Atoi(c.Param("status"))
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	req := &{{.Service}}.Get{{.Svc}}ListRequest{Status: int32(status)}

	resp, err := m.svc.Get{{.Svc}}List(c.Request.Context(), req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)
}

func (m *gateway) Update{{.Svc}}Status(c *gin.Context) {
	var req {{.Service}}.Update{{.Svc}}StatusRequest

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSONP(http.StatusNoContent, "错误请求")
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	resp, err := m.svc.Update{{.Svc}}Status(c.Request.Context(), &req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)

}

func (m *gateway) Update{{.Svc}}Count(c *gin.Context) {
	var req {{.Service}}.Update{{.Svc}}CountRequest

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSONP(http.StatusNoContent, "错误请求")
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	resp, err := m.svc.Update{{.Svc}}Count(c.Request.Context(), &req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)

}

func (m *gateway) Delete{{.Svc}}ById(c *gin.Context) {
	var req {{.Service}}.Delete{{.Svc}}ByIdRequest

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSONP(http.StatusNoContent, "错误请求")
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		c.JSONP(http.StatusBadRequest, "错误请求")
		return
	}

	resp, err := m.svc.Delete{{.Svc}}ById(c.Request.Context(), &req, client.WithAddress(m.s.Options().Address))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, "错误请求")
		return
	}
	c.JSONP(http.StatusOK, resp)

}
`

	t, err := template.New("gateway.go").Parse(tpl)
	if err != nil {
		return err
	}

	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/gateway/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + data.Service + ".go"

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
