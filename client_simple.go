// package mongo
package mongo

import (
	"context"
	"fmt"
	"time"
)

func SimpleCreateIndex() {
	opt := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opt)
	if err != nil {
		fmt.Println("get mongotool client error")
		return
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")
	err = collectionObj.CreateOneIndex(client.ctx, "index_test_name_1", []Index{{Key: "test", Value: IndexTypeAscending}}, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}

func SimpleInsertDoc() {
	opt := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opt)
	if err != nil {
		fmt.Println("get mongotool client error")
		return
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")
	id, err := collectionObj.InsertDoc(client.ctx, map[string]interface{}{
		"key1": "123",
		"key2": 1,
		"key3": 0.123456,
		"key4": time.Now(),
		"key5": map[string]interface{}{
			"key1": "123",
			"key2": 1,
			"key3": 0.123456,
		},
	})
	if err != nil {
		_ = fmt.Errorf("insert doc err=%w", err)
	}
	fmt.Println(id)
}

func SimpleInsertDocs() {
	opt := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opt)
	if err != nil {
		fmt.Println("get mongotool client error")
		return
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")
	ids, err := collectionObj.InsertDocs(client.ctx, []interface{}{
		map[string]interface{}{
			"key1": "123",
			"key2": 1,
			"key3": 0.123456,
			"key4": time.Now(),
			"key5": map[string]interface{}{
				"key1": "123",
				"key2": 1,
				"key3": 0.123456,
			},
		},
	}, false)
	if err != nil {
		_ = fmt.Errorf("insert doc err=%w", err)
	}
	fmt.Println(ids)
}

func SimpleFindDocs() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("key1__icontains", "2"))

	opt := NewFindOptions()
	opt.Select([]string{"-key5.key1"}).Sort([]string{"-key4"})
	opt.Page(2, 1)

	var v []interface{}
	err = collectionObj.FindDocs(client.ctx, q, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	for _, i := range v {
		fmt.Println(i)
	}
	fmt.Println("ok")
}

func SimpleFindDocs2() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("key1__icontains", "2"))

	type Test struct {
		ID   string
		Key1 string
		Key2 int32
		Key3 float64
	}

	opt := NewFindOptions()
	opt.Select([]string{"-key5.key1"}).Sort([]string{"-key4"})

	var v []Test
	err = collectionObj.FindDocs(client.ctx, q, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	for _, i := range v {
		fmt.Println(i)
	}
	fmt.Println("ok")
}

func SimpleFindOne() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	id, err := String2ObjectID("62788a8e92961d9287c6b8d2")
	if err != nil {
		return
	}
	q.And(Q("_id", id))

	opt := NewFindOneOptions()
	opt.Select([]string{"-key5.key1"}).Sort([]string{"-key4"})

	var v map[string]interface{}
	err = collectionObj.FindOne(client.ctx, q, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(v)
}

func SimpleFindOneAndDelete() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("627872b84aa071cebcd9e55e")))

	opt := NewFindOneAndDelete()
	opt.Select([]string{"key5.key1"}).Sort([]string{"key4"})

	var v map[string]interface{}
	err = collectionObj.FindOneAndDelete(client.ctx, q, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(v)
}

func SimpleFindOneAndReplace() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("6278619394410770437640ee")))

	opt := NewFindOneAndReplace()
	opt.Select([]string{"key5.key1"}).Sort([]string{"key4"})
	opt.Upsert(true)

	var v map[string]interface{}
	n := map[string]interface{}{
		"qwe":       "123",
		"key5.key5": "1",
	}
	err = collectionObj.FindOneAndReplace(client.ctx, q, n, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(v)
}

func SimpleFindOneAndUpdate() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("627896f258ada2e96014e3c0")))

	opt := NewFindOneAndUpdate()
	opt.Select([]string{"key5.key1"}).Sort([]string{"key4"})
	opt.Upsert(true)

	var v map[string]interface{}
	n := map[string]interface{}{
		"qwe":       "123",
		"key5.key5": "1",
	}
	err = collectionObj.FindOneAndUpdate(client.ctx, q, n, &v, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(v)
}

func SimpleUpdateOne() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("63302f47767bc2755b78ca51")))

	opt := NewUpdate()
	opt.Upsert(true)

	id := NewObjectID()
	n := map[string]interface{}{
		"_id":       id,
		"qwe":       "123",
		"key5.key5": "111",
	}
	data, err := collectionObj.UpdateOne(client.ctx, q, n, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(data)
}

func SimpleUpdateMany() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("627896f258ada2e96014e3c0")))

	opt := NewUpdate()
	opt.Upsert(true)

	n := map[string]interface{}{
		"qwe":       "123",
		"key5.key5": "111",
	}
	data, err := collectionObj.UpdateMany(client.ctx, q, n, opt)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Println(data)
}

func SimpleBulkWrite() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	q := NewQuery()
	q.And(Q("_id", TryString2ObjectID("627896f258ada2e96014e3c0")))

	bwm := NewBulkWriteModel()
	bwm.SetOrdered(true)

	bwm.AddInsertOneModel(map[string]interface{}{
		"test": "123",
	})

	// bwm.Empty()
	write, err := collectionObj.BulkWrite(client.ctx, bwm)
	if err != nil {
		panic(err)
	}

	fmt.Println(write)
}

func SimpleDistinct() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	q := NewQuery()

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")
	distinct, err := collectionObj.Distinct(client.ctx, "key1", q, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(distinct)
}

func SimpleCount() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}

	q := NewQuery()

	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")
	count, err := collectionObj.Count(client.ctx, q, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}

func SimpleAggregate() {
	opts := OptionsFromURI("mongodb://localhost:27017")
	client, err := NewClient(context.Background(), opts)
	if err != nil {
		fmt.Println("get mongotool client error")
		panic(err)
	}
	dbObj := client.Database("test_db")
	collectionObj := dbObj.Collection("test_collection")

	var results []interface{}
	pipeLine := map[string]interface{}{
		"$group": map[string]interface{}{
			"_id": "$test",
			"test": map[string]interface{}{
				"$first": "$test",
			},
		},
	}
	err = collectionObj.Aggregate(client.ctx, &results, nil, pipeLine)
	if err != nil {
		panic(err)
	}

	for _, i := range results {
		fmt.Println(i)
	}
}
