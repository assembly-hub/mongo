// Package mongo
package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BulkWriteModel 批量写入模型
type BulkWriteModel struct {
	models  []mongo.WriteModel
	ordered *bool
}

// NewBulkWriteModel 创建批量写入模型
func NewBulkWriteModel() *BulkWriteModel {
	bulkWriteModel := new(BulkWriteModel)
	return bulkWriteModel
}

func (bwm *BulkWriteModel) AddUpdateOneModel(filter *Query, upDoc map[string]interface{}, upsert bool) *BulkWriteModel {
	updateOperation := bson.M{"$set": upDoc}
	upModel := mongo.NewUpdateOneModel().SetFilter(filter.Cond()).SetUpdate(updateOperation).SetUpsert(upsert)
	bwm.models = append(bwm.models, upModel)
	return bwm
}

func (bwm *BulkWriteModel) AddUpdateManyModel(filter *Query, upDoc map[string]interface{}, upsert bool) *BulkWriteModel {
	updateOperation := bson.M{"$set": upDoc}
	upModel := mongo.NewUpdateManyModel().SetFilter(filter.Cond()).SetUpdate(updateOperation).SetUpsert(upsert)
	bwm.models = append(bwm.models, upModel)
	return bwm
}

// AddUpdateOneCustomModel 自定义类型包含：
func (bwm *BulkWriteModel) AddUpdateOneCustomModel(filter *Query, update updateType, upDoc map[string]interface{}, upsert bool) *BulkWriteModel {
	updateOperation := bson.M{update.String(): upDoc}
	upModel := mongo.NewUpdateOneModel().SetFilter(filter.Cond()).SetUpdate(updateOperation).SetUpsert(upsert)
	bwm.models = append(bwm.models, upModel)
	return bwm
}

func (bwm *BulkWriteModel) AddUpdateManyCustomModel(filter *Query, update updateType, upDoc map[string]interface{}, upsert bool) *BulkWriteModel {
	updateOperation := bson.M{update.String(): upDoc}
	upModel := mongo.NewUpdateManyModel().SetFilter(filter.Cond()).SetUpdate(updateOperation).SetUpsert(upsert)
	bwm.models = append(bwm.models, upModel)
	return bwm
}

func (bwm *BulkWriteModel) AddDeleteOneModel(filter *Query) *BulkWriteModel {
	delModel := mongo.NewDeleteOneModel().SetFilter(filter.Cond())
	bwm.models = append(bwm.models, delModel)
	return bwm
}

func (bwm *BulkWriteModel) AddDeleteManyModel(filter *Query) *BulkWriteModel {
	delModel := mongo.NewDeleteManyModel().SetFilter(filter.Cond())
	bwm.models = append(bwm.models, delModel)
	return bwm
}

func (bwm *BulkWriteModel) AddInsertOneModel(doc interface{}) *BulkWriteModel {
	var m map[string]interface{}
	switch doc := doc.(type) {
	case map[string]interface{}:
		m = doc
	default:
		m = Struct2Map(doc)
	}

	if id, ok := m["_id"]; ok {
		switch id := id.(type) {
		case string:
			if id == "" {
				delete(m, "_id")
			} else {
				m["_id"] = TryString2ObjectID(id)
			}
		}
	}

	delModel := mongo.NewInsertOneModel().SetDocument(m)
	bwm.models = append(bwm.models, delModel)
	return bwm
}

func (bwm *BulkWriteModel) AddNewReplaceOneModel(filter *Query, replaceDoc map[string]interface{}, upsert bool) *BulkWriteModel {
	delModel := mongo.NewReplaceOneModel().SetFilter(filter.Cond()).SetReplacement(replaceDoc).SetUpsert(upsert)
	bwm.models = append(bwm.models, delModel)
	return bwm
}

func (bwm *BulkWriteModel) SetOrdered(ordered bool) *BulkWriteModel {
	bwm.ordered = &ordered
	return bwm
}

func (bwm *BulkWriteModel) Empty() bool {
	return len(bwm.models) <= 0
}
