package main

import (
	"context"
	"fmt"

	"github.com/assembly-hub/mongo"
	"github.com/assembly-hub/mongo/example/dao"
)

var mongoRef = mongo.NewReference()

func initRef() {
	// 表定义：建议放在表结构文件的init函数中

	// 定义表结构
	mongoRef.AddTableDef("table1", dao.Table1{})
	mongoRef.AddTableDef("table2", dao.Table2{})
	mongoRef.AddTableDef("table3", dao.Table3{})

	// 编译库结构
	mongoRef.BuildRefs()
}

func getMongoConn() *mongo.Client {
	opts := mongo.OptionsFromURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongo client error")
		panic(err)
	}

	return client
}

func main() {
	// 初始化表定义
	initRef()

	// 初始化数据库链接
	client := getMongoConn()
	db := client.Database("example")
	ctx := context.Background()

	tb1 := mongo.NewORMByDB(ctx, db, "table1", mongoRef)

	exist, err := tb1.Exist()
	if err != nil {
		panic(err)
	}
	fmt.Println(exist, err)

	/**
	查询语法
	= : "key": "val" or "key__eq": "val"
	< : "key__lt": 1
	<= : "key__lte": 1
	> : "key__gt": 1
	>= : "key__gte": 1
	!= : "key__ne": 1
	in : "key__in": [1]
	all : "key__all": [1, 2, 3]
	not in : "key__nin": [1]
	size : "arr__size": 1
	exists : "key__exists": true
	mod : "key__mod": [10, 1], 基数，余数
	elemMatch : "key__match": MongoQuery
	like :
		"key__istartswith": "123"
		"key__startswith": "123"
		"key__iendswith": "123"
		"key__endswith": "123"
		"key__icontains": "123"
		"key__contains": "123"

	geo within query 2dsphere
		key__geo_within_polygon: [][]float64 or [][][]float64 闭合的多边形
		key__geo_within_multi_polygon: [][][]float64 or [][][][]float64 多个闭合的多边形
		key__geo_within_center_sphere: []float64, 长度必须是3，分辨为 lng, lat, radius(米)

	geo intersects query 2dsphere
		key__geo_intersects_polygon: [][]float64 or [][][]float64 闭合的多边形

	geo near query 2dsphere
				key__near: []float64 坐标点 or map[string]interface{}{
					"point": []float64 坐标点,
					"min": 最小距离，单位：米，
					"max": 最大距离，单位：米
				}
				key__near_sphere: []float64 坐标点 or map[string]interface{}{
					"point": []float64 坐标点,
					"min": 最小距离，单位：米，
					"max": 最大距离，单位：米
				}

	geo within query 2d
		key__geo_within_2d_box: [][]float64, bottom left, top right
		key__geo_within_2d_polygon: [][]float64 非闭合的多边形
		key__geo_within_2d_center: []float64, 长度必须是3，分辨为 lng, lat, radius(米)

	~: 结果取反，其中：$and、$or、$nor不适应
	$and、$or、$nor: values type must be map[string]interface{} or []*MongoQuery or []interface or []map[string]interface{}
	*/
	tb1.Query("txt", "1", "ref", mongo.RefWhere{
		"name": "test",
		"ref": mongo.RefWhere{ // 查询子表条件，可嵌套
			"txt": "1", // 子表的具体条件
		},
	}, "$or", mongo.Where{
		"txt": "1",
		"ref2": mongo.RefWhere{
			"txt": "1",
		},
	})

	//tb1.Select("txt", "-name")
	//tb1.Order("txt", "-name")

	var ret dao.Table1
	// 参数：接收数据的容器，可以是 struct 或 map 或 []map 或 []struct，参数非slice时，查询会只关心第一条
	err = tb1.ToData(&ret)
	if err != nil {
		panic(err)
	}

	// ret.ID.IsZero() 判断ID是否为空

	var dt map[string]interface{}
	// 外键的数据获取方式 1
	err = ret.Ref.ToData(ctx, db, &dt)
	if err != nil {
		panic(err)
	}
	fmt.Println(dt)

	// 外键的数据获取方式 2
	dt2, err := ret.Ref.GetData(context.Background(), db)
	if err != nil {
		panic(err)
	}
	fmt.Println(dt2.ID)

	dt3, err := ret.Ref2.GetData(context.Background(), db)
	if err != nil {
		panic(err)
	}
	fmt.Println("dt3: ", dt3[0].ID)

	// tb1.Page(no, size) 设置页参数
	// tb1.Limit(n) 限制数据条数
	// tb1.OverLimit(n, m) 跳过n 取m
	var data []dao.Table1
	// 按照页码取数据
	pageData, err := tb1.PageData(&data, 1, 10)
	if err != nil {
		panic(err)
	}
	fmt.Println(pageData)

	bwm := mongo.NewBulkWriteModel()
	// bwm.AddDeleteManyModel()
	// bwm.AddInsertOneModel()
	write, err := tb1.BulkWrite(bwm)
	if err != nil {
		return
	}
	fmt.Println(write)

	// 事务，要求mongo版本>4.0，需要mongo副本集群
	// 返回nil事务执行成功
	err = mongo.TransSession(client, func(sessionCtx mongo.SessionContext) error {
		sessTb1 := mongo.NewORMByClient(sessionCtx, client, "example", "table1", mongoRef)
		// 操作
		//sessTb1.InsertOne()
		//sessTb1.DeleteOne()
		_, err = sessTb1.UpdateOne(map[string]interface{}{
			"name": "test",
		}, true)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}
