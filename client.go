// package mongo
package mongo

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/assembly-hub/basics/util"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonoptions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

type SessionContext = mongo.SessionContext

type Client struct {
	clientOptions []*options.ClientOptions
	ctx           context.Context
	mongoClient   *mongo.Client
}

func Connection(ctx context.Context, appName string, mongoConf *Conf) *Client {
	if mongoConf.AuthDB == "" {
		mongoConf.AuthDB = mongoConf.DB
	}
	if mongoConf.HostMaster == "" {
		panic(fmt.Sprintf("mongotool[%s] HostMaster error", appName))
	}
	hosts := []string{mongoConf.HostMaster}

	if mongoConf.HostSlave != "" {
		hosts = append(hosts, mongoConf.HostSlave)
	}

	if mongoConf.DB == "" {
		panic(fmt.Sprintf("mongotool[%s] db error", appName))
	}

	auth := ""
	if mongoConf.User != "" {
		auth = mongoConf.User
		if mongoConf.Pass != "" {
			auth += ":" + mongoConf.Pass
		}

		auth += "@"
	}

	var params []string
	if mongoConf.ServerSelectionTimeoutMS <= 0 {
		mongoConf.ServerSelectionTimeoutMS = 5000
	}
	params = append(params, fmt.Sprintf("serverSelectionTimeoutMS=%d", mongoConf.ServerSelectionTimeoutMS))

	if mongoConf.ConnectTimeoutMS <= 0 {
		mongoConf.ConnectTimeoutMS = 10000
	}
	params = append(params, fmt.Sprintf("connectTimeoutMS=%d", mongoConf.ConnectTimeoutMS))

	if mongoConf.AuthMechanism == "" {
		mongoConf.AuthMechanism = "SCRAM-SHA-1"
	}
	params = append(params, fmt.Sprintf("authMechanism=%s", mongoConf.AuthMechanism))

	params = append(params, fmt.Sprintf("authSource=%s", mongoConf.AuthDB))

	if mongoConf.ReplicaSet != "" {
		params = append(params, fmt.Sprintf("replicaSet=%s", mongoConf.ReplicaSet))
	} else {
		params = append(params, "connect=direct")
	}

	uri := fmt.Sprintf("mongodb://%s%s/%s?%s",
		auth, util.JoinArr(hosts, ","), mongoConf.DB, util.JoinArr(params, "&"))

	opts := OptionsFromURI(uri)

	opts.AppName = &appName
	minPoolSize := uint64(5)
	maxPoolSize := uint64(runtime.GOMAXPROCS(0) * 5)

	opts.MinPoolSize = &minPoolSize
	opts.MaxPoolSize = &maxPoolSize

	client, err := NewClient(ctx, opts)
	if err != nil {
		panic(err)
	}

	if !mongoConf.Connect {
		log.Println("mongotool created")
		return client
	}

	err = client.Ping(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("mongotool connected")
	return client
}

func NewClient(ctx context.Context, opt *ClientOptions) (*Client, error) {
	if ctx == nil || opt == nil {
		return nil, fmt.Errorf("ctx or opt not be nil")
	}

	optList := []*options.ClientOptions{opt.ClientOptions, options.Client().SetRegistry(register())}
	client, err := mongo.Connect(ctx, optList...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	c := new(Client)
	c.clientOptions = optList
	c.mongoClient = client
	c.ctx = ctx
	return c, nil
}

func register() *bsoncodec.Registry {
	builder := bsoncodec.NewRegistryBuilder()

	// 注册默认的编码和解码器
	bsoncodec.DefaultValueEncoders{}.RegisterDefaultEncoders(builder)
	bsoncodec.DefaultValueDecoders{}.RegisterDefaultDecoders(builder)

	// 注册时间解码器
	tTime := reflect.TypeOf(time.Time{})
	tCodec := bsoncodec.NewTimeCodec(bsonoptions.TimeCodec().SetUseLocalTimeZone(true))
	registry := builder.RegisterTypeDecoder(tTime, tCodec).Build()

	return registry
}

func (c *Client) Ping(ctx context.Context) error {
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	err := c.mongoClient.Ping(ctxObj, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Ctx() context.Context {
	return c.ctx
}

// NewSession
// 要求mongo 版本 4.0起
// 需要mongo副本集群
func (c *Client) NewSession(fn func(sessionCtx SessionContext) error) error {
	// session
	sessionOpts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	session, err := c.mongoClient.StartSession(sessionOpts)
	if err != nil {
		log.Println(err)
		return err
	}
	defer session.EndSession(context.Background())

	// transaction
	err = mongo.WithSession(c.ctx, session, func(sessionCtx SessionContext) (err error) {
		defer func() {
			if p := recover(); p != nil {
				errTx := session.AbortTransaction(context.Background())
				err = fmt.Errorf("%v, Transaction: %w", p, errTx)
			}
		}()

		if err = session.StartTransaction(); err != nil {
			return err
		}

		err = fn(sessionCtx)
		if err != nil {
			log.Println(err)
			err2 := session.AbortTransaction(context.Background())
			if err2 != nil {
				return fmt.Errorf("%w, %v", err2, err)
			}
			return err
		}
		return session.CommitTransaction(context.Background())
	})

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c *Client) Database(dbName string) *Database {
	db := new(Database)
	db.Client = c
	db.dbName = dbName
	db.db = c.mongoClient.Database(dbName)
	return db
}

func (c *Client) TryDatabase(dbName string) (db *Database, exist bool, err error) {
	names, err := c.mongoClient.ListDatabaseNames(c.ctx, map[string]string{"name": dbName})
	if err != nil {
		return nil, false, err
	}

	exist = false
	if len(names) > 0 {
		exist = true
	}

	db = new(Database)
	db.Client = c
	db.dbName = dbName
	db.db = c.mongoClient.Database(dbName)
	return db, exist, nil
}
