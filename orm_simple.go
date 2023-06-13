// Package mongo
package mongo

import (
	"context"
	"fmt"
)

type tb1 struct {
	ID   ObjectID         `bson:"_id" json:"id"`
	Txt  string           `bson:"txt" json:"txt"`
	Ref  *Foreign[tb2]    `bson:"ref" json:"ref" ref:"def"`
	Ref2 ForeignList[tb3] `bson:"ref2" json:"ref2" ref:"match"`
}

type tb2 struct {
	ID   string        `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
	Ref  *Foreign[tb3] `bson:"ref" json:"ref" ref:"def"`
}

type tb3 struct {
	ID  string `bson:"_id" json:"id"`
	Txt string `bson:"txt" json:"txt"`
}

func simpleOrm() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	ref := NewReference()
	ref.AddTableDef("test1", tb1{})
	ref.AddTableDef("test2", tb2{})
	ref.AddTableDef("test3", tb3{})
	ref.BuildRefs()

	dbObj := client.Database("test_db")

	q := NewORMByDB(context.Background(), dbObj, "test1", ref)
	exist, err := q.Exist()
	if err != nil {
		panic(err)
	}
	fmt.Println(exist, err.Error())

	q.Query("txt", "1", "ref", RefWhere{
		"name": "test",
		"ref": RefWhere{
			"txt": "1",
		},
	}, "$or", Where{
		"txt": "1",
		"ref2": RefWhere{
			"txt": "1",
		},
	})
	//q.Wheres(map[string]interface{}{
	//	"txt": "123",
	//})

	var ret tb1

	// var ret tb1
	err = q.ToData(&ret)
	if err != nil {
		panic(err)
	}

	fmt.Println("ret: ", ret)

	if ret.ID.IsZero() {
		return
	}

	var dt map[string]interface{}
	err = ret.Ref.ToData(client.ctx, dbObj, &dt)
	if err != nil {
		panic(err)
	}
	fmt.Println(dt)

	dt2, err := ret.Ref.GetData(context.Background(), q.Database())
	if err != nil {
		panic(err)
	}

	fmt.Println(dt2)

	dt3, err := ret.Ref2.GetData(context.Background(), q.Database())
	if err != nil {
		panic(err)
	}

	fmt.Println("dt3: ", dt3)
}

func simpleRef() {
	ref := NewReference()
	ref.AddTableDef("test1", tb1{})
	ref.AddTableDef("test2", tb2{})
	ref.AddTableDef("test3", tb3{})
	ref.BuildRefs()

	ret := tb1{}
	ret.ID = NewObjectID()

	data, err := ToRefData(ref, &ret)
	fmt.Println(data, err)

	p := []*tb1{&ret}
	r, err := ToRefListData[tb1](ref, p)
	fmt.Println(r, err)
}

func simpleMongoOrmQueryInsertOne() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	ref := NewReference()
	ref.AddTableDef("test1", tb1{})
	ref.AddTableDef("test2", tb2{})
	ref.AddTableDef("test3", tb3{})
	ref.BuildRefs()

	dbObj := client.Database("test_db")
	q := &ORM{
		db:        dbObj,
		tableName: "test1",
	}

	id := ObjectID2String(NewObjectID())
	fmt.Println(id)

	one, err := q.InsertOne(map[string]interface{}{
		"_id":  id,
		"test": 123,
	})
	if err != nil {
		_ = fmt.Errorf("insert err:%w", err)
	}

	fmt.Println(one)
}

func simpleProjection() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")

	q := NewORMByDB(context.Background(), dbObj, "test_arr", NewReference())
	q.Query("arr_list.k1", "k").Select("arr_list.k2").Projection("arr_list.$", true)

	// var s string
	var s []tb1
	err = q.ToData(&s)
	if err != nil {
		panic(err)
	}

	fmt.Println("ret: ", s)
}
