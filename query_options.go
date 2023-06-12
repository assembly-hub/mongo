// Package mongo
package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
)

// 所有索引类型
const (
	IndexTypeAscending   = 1
	IndexTypeDescending  = -1
	IndexTypeHashed      = "hashed"
	IndexTypeText        = "text"
	IndexTypeGeo2dSphere = "2dsphere"
	IndexTypeGeo2d       = "2d"
	IndexTypeGeoHaystack = "geoHaystack"
)

type updateType string

func (u updateType) String() string {
	return string(u)
}

// 更新类型名称
const (
	UpdateSet   = updateType("$set")
	UpdateUnset = updateType("$unset")
	UpdateInc   = updateType("$inc")
	UpdatePush  = updateType("$push")
	UpdatePull  = updateType("$pull")
	UpdatePop   = updateType("$pop")
)

type BasicFindOptions struct {
	field      bson.D
	sort       bson.D
	projection map[string]interface{}
}

// Select 设置需要展示或隐藏的字段
// 如：["key1", "-key2"]，显示key1，隐藏key2
func (op *BasicFindOptions) Select(cols []string) *BasicFindOptions {
	for _, col := range cols {
		v := 1
		if col[0] == '-' {
			v = 0
			col = col[1:]
		} else if col[0] == '+' {
			col = col[1:]
		}
		op.field = append(op.field, bson.E{Key: col, Value: v})
	}
	return op
}

func (op *BasicFindOptions) Projection(projectionMap map[string]interface{}) *BasicFindOptions {
	if op.projection == nil {
		op.projection = map[string]interface{}{}
	}
	for k, v := range projectionMap {
		op.projection[k] = v
	}
	return op
}

// Sort 设置排序字段
// 如：["key1", "-key2"]，key1 升序，key2 降序
func (op *BasicFindOptions) Sort(cols []string) *BasicFindOptions {
	for _, col := range cols {
		v := 1
		if col[0] == '-' {
			v = -1
			col = col[1:]
		} else if col[0] == '+' {
			col = col[1:]
		}
		op.sort = append(op.sort, bson.E{Key: col, Value: v})
	}
	return op
}

type BasicUpdateOptions struct {
	upsert *bool
}

func (op *BasicUpdateOptions) Upsert(b bool) *BasicUpdateOptions {
	op.upsert = &b
	return op
}

type Update struct {
	BasicUpdateOptions
}

type Replace struct {
	BasicUpdateOptions
}

type FindOneAndDelete struct {
	BasicFindOptions
}

type FindOneAndReplace struct {
	BasicFindOptions
	BasicUpdateOptions
}

type FindOneAndUpdate struct {
	BasicFindOptions
	BasicUpdateOptions
}

type FindOneOptions struct {
	BasicFindOptions

	skip *int64
}

// FindOptions 查询文档限制条件
type FindOptions struct {
	FindOneOptions

	limit *int64
}

type CountOptions struct {
	skip  *int64
	limit *int64
}

func NewCount() *CountOptions {
	op := new(CountOptions)
	return op
}

// Skip 跳过的文档数
func (op *CountOptions) Skip(i int64) *CountOptions {
	if i < 0 {
		panic("Skip param must be gte 0")
	}
	op.skip = &i
	return op
}

// Limit 限制数据数量
func (op *CountOptions) Limit(i int64) *CountOptions {
	if i < 1 {
		panic("limit param must be gte 1")
	}
	op.limit = &i
	return op
}

func NewUpdate() *Update {
	op := new(Update)
	return op
}

func NewReplace() *Replace {
	op := new(Replace)
	return op
}

func NewFindOneAndDelete() *FindOneAndDelete {
	op := new(FindOneAndDelete)
	return op
}

func NewFindOneAndReplace() *FindOneAndReplace {
	op := new(FindOneAndReplace)
	return op
}

func NewFindOneAndUpdate() *FindOneAndUpdate {
	op := new(FindOneAndUpdate)
	return op
}

// NewFindOneOptions 创建新的查询选项对象
func NewFindOneOptions() *FindOneOptions {
	op := new(FindOneOptions)
	return op
}

// Skip 跳过的文档数
func (op *FindOneOptions) Skip(i int64) *FindOneOptions {
	if i < 0 {
		panic("Skip param must be gte 0")
	}
	op.skip = &i
	return op
}

// NewFindOptions 创建新的查询选项对象
func NewFindOptions() *FindOptions {
	op := new(FindOptions)
	return op
}

// Limit 限制数据数量
func (op *FindOptions) Limit(i int64) *FindOptions {
	if i < 1 {
		panic("limit param must be gte 1")
	}
	op.limit = &i
	return op
}

// Page 设置分页数据
func (op *FindOptions) Page(no int64, size int64) *FindOptions {
	if no < 1 {
		panic("page no must be gte 1")
	}
	if size < 1 {
		panic("page size must be gte 1")
	}

	skip := (no - 1) * size
	op.skip = &skip
	op.limit = &size
	return op
}

// Index 索引数据结构
type Index struct {
	Key   string
	Value interface{}
}

// ManyIndex 索引数据结构
type ManyIndex struct {
	IndexName string
	IndexData []Index
	Unique    bool
}
