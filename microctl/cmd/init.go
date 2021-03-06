package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	tpl "github.com/imind-lab/micro/microctl/template"
)

var (
	domain  string
	project string
	service string
	gateway bool
)

var serverCmd = &cobra.Command{
	Use:   "init",
	Short: "Use microctl create new microservice",
	Run: func(cmd *cobra.Command, args []string) {
		date := time.Now().Format("2006/01/02")
		year := time.Now().Format("2006")

		data := &tpl.Data{
			Domain:  domain,
			Project: project,
			Service: service,
			Svc:     strings.Title(service),
			Date:    date,
			Year:    year,
		}

		err := tpl.CreateModel(data)
		if err == nil {
			fmt.Println("[INFO]生成Model成功")
		} else {
			fmt.Println("[ERROR]生成Model出错", err)
		}

		err = tpl.CreateRepository(data)
		if err == nil {
			fmt.Println("[INFO]生成Repository成功")
		} else {
			fmt.Println("[ERROR]生成Repository出错", err)
		}

		err = tpl.CreateProto(data)
		if err == nil {
			fmt.Println("[INFO]生成Proto成功")
		} else {
			fmt.Println("[ERROR]生成Proto出错", err)
		}

		err = tpl.CreateBuild(data)
		if err == nil {
			fmt.Println("[INFO]生成Build成功")
		} else {
			fmt.Println("[ERROR]生成Build出错", err)
		}

		err = tpl.CreateConf(data)
		if err == nil {
			fmt.Println("[INFO]生成Conf成功")
		} else {
			fmt.Println("[ERROR]生成Conf出错", err)
		}

		err = tpl.CreateDomain(data)
		if err == nil {
			fmt.Println("[INFO]生成Domain成功")
		} else {
			fmt.Println("[ERROR]生成Domain出错", err)
		}

		err = tpl.CreateService(data)
		if err == nil {
			fmt.Println("[INFO]生成Service成功")
		} else {
			fmt.Println("[ERROR]生成Service出错", err)
		}

		err = tpl.CreateGateway(data)
		if err == nil {
			fmt.Println("[INFO]生成Gateway成功")
		} else {
			fmt.Println("[ERROR]生成Gateway出错", err)
		}

		err = tpl.CreateMain(data)
		if err == nil {
			fmt.Println("[INFO]生成Main成功")
		} else {
			fmt.Println("[ERROR]生成Main出错", err)
		}

		err = tpl.CreatePkg(data)
		if err == nil {
			fmt.Println("[INFO]生成Pkg成功")
		} else {
			fmt.Println("[ERROR]生成Pkg出错", err)
		}

		err = tpl.CreateServer(data)
		if err == nil {
			fmt.Println("[INFO]生成Server成功")
		} else {
			fmt.Println("[ERROR]生成Server出错", err)
		}

		err = tpl.CreateDeploy(data)
		if err == nil {
			fmt.Println("[INFO]生成Deploy成功")
		} else {
			fmt.Println("[ERROR]生成Deploy出错", err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&domain, "domain", "d", "github.com", "company domain")
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "imind-lab", "project name")
	rootCmd.PersistentFlags().StringVarP(&service, "service", "s", "greeter", "service name")
	rootCmd.PersistentFlags().BoolVarP(&gateway, "api", "a", true, "generate api-gateway")
	rootCmd.AddCommand(serverCmd)
}
