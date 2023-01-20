// package mongo
package mongo

import "go.mongodb.org/mongo-driver/mongo/options"

type Conf struct {
	HostMaster               string
	HostSlave                string
	ReplicaSet               string
	User                     string
	Pass                     string
	DB                       string
	AuthDB                   string
	AuthMechanism            string
	ServerSelectionTimeoutMS int
	ConnectTimeoutMS         int
	Connect                  bool
}

// ClientOptions 客户端配置
type ClientOptions struct {
	*options.ClientOptions
}

// OptionsFromURI 根据uri创建配置
func OptionsFromURI(uri string) *ClientOptions {
	return &ClientOptions{
		options.Client().ApplyURI(uri),
	}
}

// CreateEmptyOptions 创建空的配置
func CreateEmptyOptions() *ClientOptions {
	return &ClientOptions{
		options.Client(),
	}
}
