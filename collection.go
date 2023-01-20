// package mongo
package mongo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Collection struct {
	*Database

	collectionName string
	collection     *mongo.Collection
}

// CreateOneIndex 创建索引
func (c *Collection) CreateOneIndex(ctx context.Context, indexName string, keys []Index, indexUnique bool) error {
	indexView := c.collection.Indexes()

	if len(keys) <= 0 {
		return errors.New("index keys have nothing")
	}

	indexKeys := bson.D{}
	for _, key := range keys {
		indexKeys = append(indexKeys, bson.E{Key: key.Key, Value: key.Value})
	}

	indexModel := mongo.IndexModel{
		Keys: indexKeys,
		Options: &options.IndexOptions{
			Name:   &indexName,
			Unique: &indexUnique,
		},
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	index, err := indexView.CreateOne(ctxObj, indexModel)
	if err != nil {
		return err
	}
	log.Println("index: ", index)
	return nil
}

// CreateManyIndex 创建索引
func (c *Collection) CreateManyIndex(ctx context.Context, indexList []ManyIndex) error {
	indexView := c.collection.Indexes()

	if len(indexList) <= 0 {
		return errors.New("index keys have nothing")
	}

	var indexModelList []mongo.IndexModel
	for _, index := range indexList {
		indexKeys := bson.D{}
		for _, key := range index.IndexData {
			indexKeys = append(indexKeys, bson.E{Key: key.Key, Value: key.Value})
		}

		indexModelList = append(indexModelList, mongo.IndexModel{
			Keys: indexKeys,
			Options: &options.IndexOptions{
				Name:   &index.IndexName,
				Unique: &index.Unique,
			},
		})
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	ret, err := indexView.CreateMany(ctxObj, indexModelList)
	if err != nil {
		return err
	}
	log.Println("index list: ", ret)
	return nil
}

// InsertDoc
// 添加一个文档
func (c *Collection) InsertDoc(ctx context.Context, doc map[string]interface{}) (string, error) {
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	insertOneResult, err := c.collection.InsertOne(ctxObj, doc)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return insertOneResult.InsertedID.(primitive.ObjectID).Hex(), nil
}

// InsertDocs
// 添加数据
func (c *Collection) InsertDocs(ctx context.Context, docs []interface{}, ordered bool) ([]string, error) {
	insertManyOpts := options.InsertMany().SetOrdered(ordered)
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	insertManyResult, err := c.collection.InsertMany(ctxObj, docs, insertManyOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rets := make([]string, len(insertManyResult.InsertedIDs))
	for i, v := range insertManyResult.InsertedIDs {
		rets[i] = v.(primitive.ObjectID).Hex()
	}
	return rets, nil
}

// FindDocs
// filter 查询对象
// results slice 结果返回，可以是map or struct
// opts 查询参数选择
func (c *Collection) FindDocs(ctx context.Context,
	filter *Query, results interface{}, opts *FindOptions) error {
	if filter == nil {
		filter = NewQuery()
	}

	mongoOpts := options.Find()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}

		if opts.skip != nil {
			mongoOpts.SetSkip(*opts.skip)
		}

		if opts.limit != nil {
			mongoOpts.SetLimit(*opts.limit)
		}

		if len(opts.projection) > 0 {
			mongoOpts.SetProjection(opts.projection)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	cur, err := c.collection.Find(ctxObj, filter.Cond(), mongoOpts)
	if err != nil {
		return err
	}

	if err = cur.All(ctxObj, results); err != nil {
		return err
	}

	return nil
}

// FindOne
// filter 查询对象
// results 结果返回，可以是map or struct
func (c *Collection) FindOne(ctx context.Context,
	filter *Query, result interface{}, opts *FindOneOptions) error {
	if filter == nil {
		filter = NewQuery()
	}

	mongoOpts := options.FindOne()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}

		if opts.skip != nil {
			mongoOpts.SetSkip(*opts.skip)
		}

		if len(opts.projection) > 0 {
			mongoOpts.SetProjection(opts.projection)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	singleResult := c.collection.FindOne(ctxObj, filter.Cond(), mongoOpts)
	if err := singleResult.Decode(result); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	return nil
}

func (c *Collection) FindOneAndDelete(ctx context.Context,
	filter *Query, delDoc interface{}, opts *FindOneAndDelete) error {
	mongoOpts := options.FindOneAndDelete()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	singleResult := c.collection.FindOneAndDelete(ctxObj, filter.Cond(), mongoOpts)
	if delDoc == nil {
		return nil
	}

	if err := singleResult.Decode(delDoc); err != nil {
		log.Println(err)
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	return nil
}

func (c *Collection) FindOneAndReplace(ctx context.Context,
	filter *Query, newDoc map[string]interface{}, oldDoc interface{}, opts *FindOneAndReplace) error {
	if newDoc == nil {
		return fmt.Errorf("newDoc is not nil")
	}

	mongoOpts := options.FindOneAndReplace()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}

		if opts.upsert != nil {
			mongoOpts.SetUpsert(*opts.upsert)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	singleResult := c.collection.FindOneAndReplace(ctxObj, filter.Cond(), newDoc, mongoOpts)
	if oldDoc == nil {
		return nil
	}

	if err := singleResult.Decode(oldDoc); err != nil {
		log.Print(err)
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	return nil
}

func (c *Collection) FindOneAndUpdate(ctx context.Context,
	filter *Query, upDoc map[string]interface{}, oldDoc interface{}, opts *FindOneAndUpdate) error {
	if upDoc == nil {
		return fmt.Errorf("upDoc is nil")
	}

	mongoOpts := options.FindOneAndUpdate()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}

		if opts.upsert != nil {
			mongoOpts.SetUpsert(*opts.upsert)
		}
	}

	delete(upDoc, "_id")

	if len(upDoc) <= 0 {
		return fmt.Errorf("upDoc is empty")
	}

	upObj := map[string]interface{}{
		"$set": upDoc,
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	singleResult := c.collection.FindOneAndUpdate(ctxObj, filter.Cond(), upObj, mongoOpts)
	if oldDoc == nil {
		return nil
	}

	if err := singleResult.Decode(oldDoc); err != nil {
		log.Println(err)
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	return nil
}

func (c *Collection) FindOneAndUpdateCustom(ctx context.Context,
	filter *Query, customDoc map[string]interface{}, oldDoc interface{}, opts *FindOneAndUpdate) error {
	if customDoc == nil {
		return fmt.Errorf("customDoc is nil")
	}

	mongoOpts := options.FindOneAndUpdate()
	if opts != nil {
		if opts.field != nil {
			mongoOpts.SetProjection(opts.field)
		}

		if opts.sort != nil {
			mongoOpts.SetSort(opts.sort)
		}

		if opts.upsert != nil {
			mongoOpts.SetUpsert(*opts.upsert)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	singleResult := c.collection.FindOneAndUpdate(ctxObj, filter.Cond(), customDoc, mongoOpts)
	if oldDoc == nil {
		return nil
	}

	if err := singleResult.Decode(oldDoc); err != nil {
		log.Println(err)
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	return nil
}

func (c *Collection) UpdateOne(ctx context.Context,
	filter *Query, upDoc map[string]interface{}, opts *Update) (*UpdateResult, error) {
	if upDoc == nil {
		return nil, fmt.Errorf("upDoc is not nil")
	}

	updateOneOpts := options.Update()
	if opts != nil {
		if opts.upsert != nil {
			updateOneOpts.SetUpsert(*opts.upsert)
		}
	}

	delete(upDoc, "_id")

	if len(upDoc) <= 0 {
		return nil, fmt.Errorf("upDoc is empty")
	}

	updateOneSet := bson.M{"$set": upDoc}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	updateResult, err := c.collection.UpdateOne(ctxObj, filter.Cond(), updateOneSet, updateOneOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  updateResult.MatchedCount,
		ModifiedCount: updateResult.ModifiedCount,
		UpsertedCount: updateResult.UpsertedCount,
		UpsertedID:    updateResult.UpsertedID,
	}, nil
}

func (c *Collection) UpdateMany(ctx context.Context,
	filter *Query, upDoc map[string]interface{}, opts *Update) (*UpdateResult, error) {
	if upDoc == nil {
		return nil, fmt.Errorf("upDoc is not nil")
	}

	updateManyOpts := options.Update()
	if opts != nil {
		if opts.upsert != nil {
			updateManyOpts.SetUpsert(*opts.upsert)
		}
	}

	delete(upDoc, "_id")

	if len(upDoc) <= 0 {
		return nil, fmt.Errorf("upDoc is empty")
	}

	updateManeySet := bson.M{"$set": upDoc}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	updateResult, err := c.collection.UpdateMany(ctxObj, filter.Cond(), updateManeySet, updateManyOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  updateResult.MatchedCount,
		ModifiedCount: updateResult.ModifiedCount,
		UpsertedCount: updateResult.UpsertedCount,
		UpsertedID:    updateResult.UpsertedID,
	}, nil
}

func (c *Collection) UpdateOneCustom(ctx context.Context,
	filter *Query, update updateType, updateDoc map[string]interface{}, opts *Update) (*UpdateResult, error) {
	if updateDoc == nil {
		return nil, fmt.Errorf("updateDoc is not nil")
	}

	updateSet := bson.M{update.String(): updateDoc}

	updateOneOpts := options.Update()
	if opts != nil {
		if opts.upsert != nil {
			updateOneOpts.SetUpsert(*opts.upsert)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	updateResult, err := c.collection.UpdateOne(ctxObj, filter.Cond(), updateSet, updateOneOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  updateResult.MatchedCount,
		ModifiedCount: updateResult.ModifiedCount,
		UpsertedCount: updateResult.UpsertedCount,
		UpsertedID:    updateResult.UpsertedID,
	}, nil
}

func (c *Collection) UpdateManyCustom(ctx context.Context,
	filter *Query, update updateType, updateDoc map[string]interface{}, opts *Update) (*UpdateResult, error) {
	if updateDoc == nil {
		return nil, fmt.Errorf("updateDoc is not nil")
	}

	if updateDoc == nil {
		return nil, fmt.Errorf("updateDoc is not nil")
	}

	updateSet := bson.M{update.String(): updateDoc}

	updateManyOpts := options.Update()
	if opts != nil {
		if opts.upsert != nil {
			updateManyOpts.SetUpsert(*opts.upsert)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	updateResult, err := c.collection.UpdateMany(ctxObj, filter.Cond(), updateSet, updateManyOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  updateResult.MatchedCount,
		ModifiedCount: updateResult.ModifiedCount,
		UpsertedCount: updateResult.UpsertedCount,
		UpsertedID:    updateResult.UpsertedID,
	}, nil
}

func (c *Collection) ReplaceOne(ctx context.Context,
	filter *Query, replaceDoc map[string]interface{}, opts *Replace) (*UpdateResult, error) {
	if replaceDoc == nil {
		return nil, fmt.Errorf("upDoc is not nil")
	}

	replaceOpts := options.Replace()
	if opts != nil {
		if opts.upsert != nil {
			replaceOpts.SetUpsert(*opts.upsert)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	updateResult, err := c.collection.ReplaceOne(ctxObj, filter.Cond(), replaceDoc, replaceOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &UpdateResult{
		MatchedCount:  updateResult.MatchedCount,
		ModifiedCount: updateResult.ModifiedCount,
		UpsertedCount: updateResult.UpsertedCount,
		UpsertedID:    updateResult.UpsertedID,
	}, nil
}

func (c *Collection) DeleteOne(ctx context.Context, filter *Query) (*DeleteResult, error) {
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	delResult, err := c.collection.DeleteOne(ctxObj, filter.Cond())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteResult{
		DeletedCount: delResult.DeletedCount,
	}, nil
}

func (c *Collection) DeleteMany(ctx context.Context, filter *Query) (*DeleteResult, error) {
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	delResult, err := c.collection.DeleteMany(ctxObj, filter.Cond())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &DeleteResult{
		DeletedCount: delResult.DeletedCount,
	}, nil
}

func (c *Collection) BulkWrite(ctx context.Context, bwm *BulkWriteModel) (*BulkWriteResult, error) {
	if len(bwm.models) <= 0 {
		return nil, fmt.Errorf("BulkWriteModel models's length is 0")
	}

	bulkWriteOpts := options.BulkWrite()
	if bwm.ordered != nil {
		bulkWriteOpts.SetOrdered(*bwm.ordered)
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	bulkWriteResults, err := c.collection.BulkWrite(ctxObj, bwm.models, bulkWriteOpts)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &BulkWriteResult{
		InsertedCount: bulkWriteResults.InsertedCount,
		MatchedCount:  bulkWriteResults.MatchedCount,
		ModifiedCount: bulkWriteResults.ModifiedCount,
		DeletedCount:  bulkWriteResults.DeletedCount,
		UpsertedCount: bulkWriteResults.UpsertedCount,
		UpsertedIDs:   bulkWriteResults.UpsertedIDs,
	}, nil
}

func (c *Collection) Distinct(ctx context.Context, fieldName string,
	filter *Query, serverMaxTime *time.Duration) ([]interface{}, error) {
	distinctOpts := options.Distinct()
	if serverMaxTime == nil {
		distinctOpts.SetMaxTime(5 * time.Second)
	} else {
		distinctOpts.SetMaxTime(*serverMaxTime)
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	distinctValues, err := c.collection.Distinct(ctxObj, fieldName, filter.Cond(), distinctOpts)
	if err != nil {
		return nil, err
	}
	return distinctValues, nil
}

func (c *Collection) Count(ctx context.Context, filter *Query, opts *CountOptions) (int64, error) {
	countOpts := options.Count()
	if opts != nil {
		if opts.limit != nil {
			countOpts.SetLimit(*opts.limit)
		}
		if opts.skip != nil {
			countOpts.SetSkip(*opts.skip)
		}
	}

	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}
	count, err := c.collection.CountDocuments(ctxObj, filter.Cond(), countOpts)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Collection) Aggregate(ctx context.Context, results interface{}, serverMaxTime *time.Duration, pipeline ...interface{}) error {
	ctxObj := c.ctx
	if ctx != nil {
		ctxObj = ctx
	}

	opts := options.Aggregate()
	if serverMaxTime == nil {
		opts.SetMaxTime(3 * time.Second)
	} else {
		opts.SetMaxTime(*serverMaxTime)
	}

	aggCursor, err := c.collection.Aggregate(ctxObj, pipeline, opts)
	if err != nil {
		log.Println(err)
		return err
	}

	if err = aggCursor.All(ctx, results); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
