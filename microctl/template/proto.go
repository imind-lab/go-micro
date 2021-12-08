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

// 生成proto
func CreateProto(data *Data) error {
	var tpl = `syntax = "proto3";

package {{.Service}};

option go_package = "{{.Domain}}/{{.Project}}/{{.Service}}/application/proto/{{.Service}}";

service {{.Svc}}Service {
    rpc Create{{.Svc}} (Create{{.Svc}}Request) returns (Create{{.Svc}}Response) {}
    rpc Get{{.Svc}}ById (Get{{.Svc}}ByIdRequest) returns (Get{{.Svc}}ByIdResponse) {}
    rpc Get{{.Svc}}List (Get{{.Svc}}ListRequest) returns (Get{{.Svc}}ListResponse) {}
    rpc Update{{.Svc}}Status (Update{{.Svc}}StatusRequest) returns (Update{{.Svc}}StatusResponse) {}
    rpc Update{{.Svc}}Count (Update{{.Svc}}CountRequest) returns (Update{{.Svc}}CountResponse) {}
    rpc Delete{{.Svc}}ById (Delete{{.Svc}}ByIdRequest) returns (Delete{{.Svc}}ByIdResponse) {}

    rpc Get{{.Svc}}ListByStream (stream Get{{.Svc}}ListByStreamRequest) returns (stream Get{{.Svc}}ListByStreamResponse);
}

message Create{{.Svc}}Request {
    // @inject_tag: validate:"required"
    {{.Svc}} dto = 1;
}

message Create{{.Svc}}Response {
    Error error = 1;
}

message Get{{.Svc}}ByIdRequest {
    int32 id = 1;
}

message Get{{.Svc}}ByIdResponse {
    Error error = 1;
    {{.Svc}} dto = 2;
}

message Get{{.Svc}}ListRequest {
    // @inject_tag: validate:"gte=0,lte=3"
    int32 status = 1;
    int32 lastid = 2;
    // @inject_tag: validate:"gte=5,lte=20"
    int32 pagesize = 3;
    int32 page = 4;
}

message Get{{.Svc}}ListResponse {
    Error error = 1;
    {{.Svc}}List data = 2;
}

message Update{{.Svc}}StatusRequest {
    int32 id = 1;
    int32 status = 2;
}

message Update{{.Svc}}StatusResponse {
    Error error = 1;
}

message Update{{.Svc}}CountRequest {
    int32 id = 1;
    int32 num = 2;
    string column = 3;
}

message Update{{.Svc}}CountResponse {
    Error error = 1;
}

message Delete{{.Svc}}ByIdRequest {
    int32 id = 1;
}

message Delete{{.Svc}}ByIdResponse {
    Error error = 1;
}

message Get{{.Svc}}ListByStreamRequest {
    int32 index = 1;
    int32 id = 2;
}

message Get{{.Svc}}ListByStreamResponse {
    int32 index = 1;
    {{.Svc}} result = 2;
}

message {{.Svc}} {
    int32 id = 1;
    // @inject_tag: validate:"required,email"
    string name = 2;
    int32 view_num = 3;
    // @inject_tag: validate:"gte=0,lte=3"
    int32 status = 4;
    int64 create_time = 5;
    string create_datetime = 6;
    string update_datetime = 7;
}

message {{.Svc}}List {
    int32 total = 1;
    int32 total_page = 2;
    int32 cur_page = 3;
    repeated {{.Svc}} datalist = 4;
}

message Error {
    int32 code = 1;
    string message = 2;
}
`

	t, err := template.New("proto").Parse(tpl)
	if err != nil {
		return err
	}

	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/application/proto/" + data.Service + "/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + data.Service + ".proto"

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
