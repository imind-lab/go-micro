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

// 生成main
func CreateMain(data *Data) error {
	var tpl = `/**
 *  MindLab
 *
 *  Create by songli on {{.Date}}
 *  Copyright © {{.Year}} imind.tech All rights reserved.
 */

package main

import (
	"{{.Domain}}/{{.Project}}/{{.Service}}/infrastructure/server"
)

func main() {
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
`

	t, err := template.New("main").Parse(tpl)
	if err != nil {
		return err
	}

	t.Option()
	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + "main.go"

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	// 生成go.mod
	tpl = `module {{.Domain}}/{{.Project}}/{{.Service}}

go 1.16

require (
	github.com/imind-lab/micro v0.1.0
	go-micro.dev/v4 v4.4.0
)
`

	t, err = template.New("go.mod").Parse(tpl)
	if err != nil {
		return err
	}

	fileName = dir + "go.mod"

	f, err = os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	// 生成go.mod
	tpl = `package main

import (
	_ "github.com/asim/go-micro/plugins/transport/grpc/v4"
	_ "github.com/asim/go-micro/plugins/broker/kafka/v4"
)
`

	t, err = template.New("plugins.go").Parse(tpl)
	if err != nil {
		return err
	}

	fileName = dir + "plugins.go"

	f, err = os.Create(fileName)
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
