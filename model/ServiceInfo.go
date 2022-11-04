package model

import "mini-dfs/contents"

type ServiceInfo struct {
	ServiceName    string                 `json:"service_name"`    //服务名
	ServiceHost    string                 `json:"service_host"`    //服务地址
	ServicePort    int32                  `json:"service_port"`    //服务端口
	ServiceProtoc  contents.ServiceProtoc `json:"service_protoc"`  //服务协议
	AdditionalMeta map[string]any         `json:"additional_meta"` //服务附加元信息
}
