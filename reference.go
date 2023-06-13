// Package mongo
package mongo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const foreignKeyErrStr = "foreign keys must use mongo.Foreign[ForeignCollectionStruct] or mongo.Foreign[ForeignList]"

// refQ Mongo ref query
type refQ struct {
	From  string
	Ref   *Reference
	DB    *Database
	Query *Query
}

func (q *refQ) getData(colName string) ([]interface{}, int) {
	if q.DB == nil {
		panic("refQ collection can not nil")
	}

	if q.Ref == nil {
		panic("refQ ref can not nil")
	}

	if q.From == "" {
		panic("from can not empty")
	}

	c := q.Query.Cond()
	if len(c) <= 0 {
		panic("refQ ORM can not Empty")
	}

	ref := q.Ref.getRef(q.From, colName)
	if ref == nil {
		panic("refQ ref data can not nil")
	}

	collection := q.DB.Collection(ref.To)

	opt := NewFindOptions()
	opt.Select([]string{"_id"})

	var idData []map[string]interface{}
	err := collection.FindDocs(q.DB.ctx, q.Query, &idData, opt)
	if err != nil {
		panic(err)
	}
	arr := []interface{}{}
	for _, v := range idData {
		arr = append(arr, v["_id"])
	}
	return arr, ref.RefType
}

type dataType interface{}

// Foreign mongotool 外键
// tag: `bson "test" json:"test" ref:"def"` ref values [def, all, match]
// def: 有交集即可；all：所有的外键均存在；match：所有的外键均存在，并且与条件完全一致
type Foreign[T dataType] struct {
	Ref string   `bson:"$ref" json:"ref"`
	ID  ObjectID `bson:"$id" json:"id"`
}

// ForeignList mongotool 外键数组
// tag: `bson "test" json:"test" ref:"def"` ref values [def, all, match]
// def: 有交集即可；all：所有的外键均存在；match：所有的外键均存在，并且与条件完全一致
type ForeignList[T dataType] []*Foreign[T]

func (r *Foreign[T]) ToData(ctx context.Context, db *Database, data interface{}) error {
	if r.Ref == "" || ObjectIDIsZero(r.ID) {
		return nil
	}
	collection := db.Collection(r.Ref)

	opt := NewFindOneOptions()
	q := MixQ(map[string]interface{}{
		"_id": r.ID,
	})
	err := collection.FindOne(ctx, q, data, opt)
	if err != nil {
		return err
	}
	return nil
}

func (r *Foreign[T]) GetData(ctx context.Context, db *Database) (*T, error) {
	var d T
	err := r.ToData(ctx, db, &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r ForeignList[T]) ToData(ctx context.Context, db *Database, data interface{}) error {
	if len(r) <= 0 {
		return nil
	}

	collectionName := ""
	var ids []primitive.ObjectID
	for _, ref := range r {
		if ref.Ref != "" {
			collectionName = ref.Ref
		}
		if !ObjectIDIsZero(ref.ID) {
			ids = append(ids, ref.ID)
		}
	}

	if collectionName == "" || len(ids) <= 0 {
		return nil
	}

	collection := db.Collection(collectionName)
	opt := NewFindOptions()
	q := MixQ(map[string]interface{}{
		"_id__in": ids,
	})
	err := collection.FindDocs(ctx, q, data, opt)
	if err != nil {
		return err
	}
	return nil
}

func (r ForeignList[T]) GetData(ctx context.Context, db *Database) ([]*T, error) {
	var arr []*T
	err := r.ToData(ctx, db, &arr)
	if err != nil {
		return nil, err
	}
	return arr, nil
}

// Reference 表定义
type Reference struct {
	tableDef      map[string]reflect.Type
	tableRef      map[string]map[string]*refType
	structToTable map[string]string
}

const (
	mongoRefDefault = 1
	mongoRefAll     = 2
	mongoRefMatch   = 3
)

// 关联类型
type refType struct {
	To      string
	RefType int // 1: def; 2: all; 3: match
}

func NewReference() *Reference {
	ref := new(Reference)
	ref.tableDef = map[string]reflect.Type{}
	ref.tableRef = map[string]map[string]*refType{}
	ref.structToTable = map[string]string{}
	return ref
}

func (r *Reference) getTableName(structFullName string) string {
	return r.structToTable[structFullName]
}

func (r *Reference) AddTableDef(tbName string, def interface{}) {
	if _, ok := r.tableDef[tbName]; ok {
		panic(fmt.Sprintf("collection [%s] is already in def", tbName))
	}

	tp := reflect.TypeOf(def)
	if tp.Kind() != reflect.Struct {
		panic(fmt.Sprintf("collection [%s] type must be struct", tbName))
	}

	structFullName := fmt.Sprintf("%s.%s", tp.PkgPath(), tp.Name())
	if _, ok := r.structToTable[structFullName]; ok {
		panic(fmt.Sprintf("table [%s] struct [%s] is exist", tbName, structFullName))
	}
	r.structToTable[structFullName] = tbName

	r.tableDef[tbName] = tp

	for i := 0; i < tp.NumField(); i++ {
		colName := tp.Field(i).Tag.Get("bson")
		if colName == "" || !tp.Field(i).IsExported() {
			continue
		}

		ref := tp.Field(i).Tag.Get("ref")
		if ref != "" {
			t := mongoRefDefault
			if ref == "all" {
				t = mongoRefAll
			} else if ref == "match" {
				t = mongoRefMatch
			}

			colType := tp.Field(i).Type
			if colType.Kind() == reflect.Ptr {
				colType = colType.Elem()
			}

			foreignName := colType.String()
			first, second := strings.Index(foreignName, "["), strings.Index(foreignName, "]")
			if first < 0 || second < 0 || second <= first {
				panic(foreignKeyErrStr)
			}
			foreignName = foreignName[first+1 : second]
			if foreignName == "" {
				panic(foreignKeyErrStr)
			}
			if foreignName[0] == '*' {
				panic(foreignKeyErrStr)
			}

			if _, ok := r.tableRef[tbName]; ok {
				r.tableRef[tbName][colName] = &refType{
					To:      foreignName,
					RefType: t,
				}
			} else {
				r.tableRef[tbName] = map[string]*refType{
					colName: {
						To:      foreignName,
						RefType: t,
					},
				}
			}
		}
	}
}

func (r *Reference) BuildRefs() {
	for _, refMap := range r.tableRef {
		for _, ref := range refMap {
			errStr := fmt.Sprintf("table[%s] not be defined", ref.To)
			ref.To = r.structToTable[ref.To]
			if ref.To == "" {
				panic(errStr)
			}
		}
	}
}

func (r *Reference) getDef(tbName string) reflect.Type {
	if def, ok := r.tableDef[tbName]; ok {
		return def
	}
	return nil
}

func (r *Reference) getRef(tbName string, colName string) *refType {
	if q, ok := r.tableRef[tbName]; ok {
		if ref, has := q[colName]; has {
			return ref
		}
	}
	return nil
}

func ToRefData[T dataType](r *Reference, data *T) (*Foreign[T], error) {
	valValue := reflect.ValueOf(data)
	if data == nil || valValue.IsNil() {
		return nil, fmt.Errorf("data type must be *struct")
	}
	if valValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("data type must be *struct")
	}

	valValue = valValue.Elem()
	valType := valValue.Type()
	if valType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("data type must be *struct")
	}

	structFullName := fmt.Sprintf("%s.%s", valType.PkgPath(), valType.Name())
	tbName := r.getTableName(structFullName)
	if tbName == "" {
		return nil, fmt.Errorf("struct[%s] is undefined", structFullName)
	}

	var objID ObjectID
	var err error
	for i := 0; i < valValue.NumField(); i++ {
		if valValue.Type().Field(i).Tag.Get("bson") == "_id" {
			if valValue.Field(i).Kind() == reflect.String {
				refID := valValue.Field(i).String()
				objID, err = primitive.ObjectIDFromHex(refID)
				if err != nil {
					return nil, err
				}
			} else {
				b := valValue.Field(i).Bytes()
				copy(objID[:], b)
			}
			break
		}
	}

	if ObjectIDIsZero(objID) {
		return nil, fmt.Errorf("ObjectID is zero")
	}

	return &Foreign[T]{
		Ref: tbName,
		ID:  objID,
	}, nil
}

type refListType[T dataType] interface {
	[]*T | []T
}

func ToRefListData[T dataType, DT refListType[T]](r *Reference, data DT) (ForeignList[T], error) {
	valValue := reflect.ValueOf(data)
	if data == nil || valValue.IsNil() {
		return nil, fmt.Errorf("data type must be slice of [*]struct")
	}
	if valValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("data type must be slice of [*]struct")
	}

	if valValue.Len() <= 0 {
		return nil, nil
	}

	refList := ForeignList[T]{}
	if strings.Contains(reflect.TypeOf(refList).Name(), "*") {
		return nil, fmt.Errorf("generic type cannot be a pointer")
	}
	for i := 0; i < valValue.Len(); i++ {
		elem := valValue.Index(i)

		structName := ""
		if elem.Type().Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		if elem.Type().Kind() != reflect.Struct {
			return nil, fmt.Errorf("data type must be slice of [*]struct")
		}

		structName = fmt.Sprintf("%s.%s", elem.Type().PkgPath(), elem.Type().Name())

		tbName := r.getTableName(structName)
		if tbName == "" {
			return nil, fmt.Errorf("struct[%s] is undefined", structName)
		}

		var objID ObjectID
		var err error
		for j := 0; j < elem.NumField(); j++ {
			if elem.Type().Field(j).Tag.Get("bson") == "_id" {
				if elem.Field(j).Kind() == reflect.String {
					refID := elem.Field(j).String()
					objID, err = primitive.ObjectIDFromHex(refID)
					if err != nil {
						return nil, err
					}
				} else {
					b := elem.Field(j).Bytes()
					copy(objID[:], b)
				}
				break
			}
		}
		if ObjectIDIsZero(objID) {
			return nil, fmt.Errorf("ObjectID is zero")
		}

		refList = append(refList, &Foreign[T]{
			Ref: tbName,
			ID:  objID,
		})
	}

	return refList, nil
}
