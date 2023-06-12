// Package mongo
package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/assembly-hub/basics/util"
	"github.com/assembly-hub/task"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	//ErrTargetNotSettable means the second param of Bind is not settable
	ErrTargetNotSettable = errors.New("[scanner]: target is not settable! a pointer is required")
)

type Select = []string
type Order = []string
type Limit = []uint
type Where = map[string]interface{}
type RefWhere = map[string]interface{}
type Projection = map[string]interface{}

type tempRefQ struct {
	From  string
	Query Where
}

type mongoOrmQ struct {
	Distinct   bool
	Select     Select
	Order      Order
	Limit      Limit
	Where      Where
	Projection Projection
}

func newMongoOrmQ() *mongoOrmQ {
	q := new(mongoOrmQ)
	q.Distinct = false
	q.Select = Select{}
	q.Order = Order{}
	q.Limit = Limit{}
	q.Where = Where{}
	q.Projection = Projection{}
	return q
}

type ORM struct {
	ctx       context.Context
	refConf   *Reference
	db        *Database
	tableName string
	keepQuery bool
	Q         *mongoOrmQ
}

type Paging struct {
	PageNo    int `json:"page_no"`    //当前页
	PageSize  int `json:"page_size"`  //每页条数
	Total     int `json:"total"`      //总条数
	PageTotal int `json:"page_total"` //总页数
}

func NewORMByDB(ctx context.Context, db *Database, tbName string, ref *Reference) *ORM {
	ctxObj := db.ctx
	if ctx != nil {
		ctxObj = ctx
	}

	return &ORM{
		ctx:       ctxObj,
		refConf:   ref,
		keepQuery: true,
		db:        db,
		tableName: tbName,
		Q:         newMongoOrmQ(),
	}
}

func NewORMByClient(ctx context.Context, cli *Client, dbName, tbName string, ref *Reference) *ORM {
	ctxObj := cli.ctx
	if ctx != nil {
		ctxObj = ctx
	}

	return &ORM{
		ctx:       ctxObj,
		refConf:   ref,
		keepQuery: true,
		db:        cli.Database(dbName),
		tableName: tbName,
		Q:         newMongoOrmQ(),
	}
}

// Query 条件对
// "id__gt", 1, "name": "test"
func (orm *ORM) Query(pair ...interface{}) *ORM {
	if len(pair)%2 != 0 {
		panic("pair长度必须是2的整数倍")
	}
	if len(pair) <= 0 {
		return orm
	}

	for i, n := 0, len(pair)/2; i < n; i++ {
		orm.Q.Where[util.Any2String(pair[i*2])] = pair[i*2+1]
	}
	return orm
}

func (orm *ORM) OverLimit(over, size uint) *ORM {
	orm.Q.Limit = Limit{over, size}
	return orm
}

func (orm *ORM) Page(pageNo, pageSize uint) *ORM {
	if pageSize <= 0 {
		panic("page size must be gt 0")
	}
	if pageNo <= 0 {
		panic("page no must be gt 0")
	}

	orm.Q.Limit = Limit{pageSize * (pageNo - 1), pageSize}
	return orm
}

func (orm *ORM) Limit(size uint) *ORM {
	orm.Q.Limit = Limit{size}
	return orm
}

func (orm *ORM) Distinct(b bool) *ORM {
	orm.Q.Distinct = b
	return orm
}

func (orm *ORM) KeepQuery(b bool) *ORM {
	orm.keepQuery = b
	return orm
}

func (orm *ORM) Where(col string, value interface{}) *ORM {
	orm.Q.Where[col] = value
	return orm
}

func (orm *ORM) Wheres(where Where) *ORM {
	for k, v := range where {
		orm.Q.Where[k] = v
	}
	return orm
}

func (orm *ORM) Projection(col string, value interface{}) *ORM {
	orm.Q.Projection[col] = value
	return orm
}

func (orm *ORM) Projections(where Where) *ORM {
	for k, v := range where {
		orm.Q.Projection[k] = v
	}
	return orm
}

func (orm *ORM) Select(cols ...string) *ORM {
	orm.Q.Select = append(orm.Q.Select, cols...)
	return orm
}

func (orm *ORM) Order(cols ...string) *ORM {
	orm.Q.Order = append(orm.Q.Order, cols...)
	return orm
}

func (orm *ORM) ClearCache() *ORM {
	orm.Q = newMongoOrmQ()
	return orm
}

func (orm *ORM) Cond() *Query {
	return MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
}

func (orm *ORM) ToJSON() string {
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	return q.JSON()
}

// Exist 检查数据是否存在
func (orm *ORM) Exist() (bool, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opts := NewFindOneOptions()
	opts.Select(Select{"_id"})
	if len(orm.Q.Limit) == 1 {
		opts.Skip(int64(orm.Q.Limit[0]))
	}
	var target map[string]string
	err := table.FindOne(orm.ctx, q, &target, opts)
	if err != nil {
		return false, err
	}

	if target != nil && target["_id"] != "" {
		return true, nil
	}

	return false, nil
}

func (orm *ORM) PageData(target interface{}, pageNo, pageSize uint) (*Paging, error) {
	totalCount, err := orm.Count(false)
	if err != nil {
		return nil, err
	}

	if pageNo == 0 || pageSize == 0 {
		return nil, fmt.Errorf("page no page size need gt 0")
	}

	totalPage := totalCount / int64(pageSize)
	if totalCount%int64(pageSize) > 0 {
		totalPage++
	}

	if pageNo > uint(totalPage) {
		pageNo = uint(totalPage)
	}
	if pageNo < 1 {
		pageNo = 1
	}
	orm.Page(pageNo, pageSize)

	err = orm.ToData(target)
	if err != nil {
		return nil, err
	}

	p := &Paging{
		PageNo:    int(pageNo),
		PageSize:  int(pageSize),
		Total:     int(totalCount),
		PageTotal: int(totalPage),
	}
	return p, nil
}

func (orm *ORM) setDataFunc(dataVal reflect.Value, v interface{}) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	val := fmt.Sprintf("%v", v)
	switch v := v.(type) {
	case string:
		dataVal.SetString(v)
	case int, int8, int16, int32, int64:
		i64, err := util.Str2Int[int64](val)
		if err != nil {
			return err
		}
		dataVal.SetInt(i64)
	case uint, uint8, uint16, uint32, uint64:
		u64, err := util.Str2Uint[uint64](val)
		if err != nil {
			return err
		}
		dataVal.SetUint(u64)
	case float32, float64:
		f64, err := util.Str2Float[float64](val)
		if err != nil {
			return err
		}
		dataVal.SetFloat(f64)
	case []byte:
		dataVal.SetBytes(v)
	case primitive.ObjectID:
		dataVal.SetString(v.Hex())
	default:
		dataVal.Set(reflect.ValueOf(v))
	}

	return nil
}

func (orm *ORM) returnDataFormat(m map[string]interface{}) (ret map[string]interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	ret = map[string]interface{}{}
	for k, v := range m {
		switch v := v.(type) {
		case map[string]interface{}:
			ret[k], err = orm.returnDataFormat(v)
			if err != nil {
				return nil, err
			}
		case primitive.A:
			ret[k], err = orm.mongoDataFormat(v)
			if err != nil {
				return nil, err
			}
		default:
			ret[k] = v
		}
	}

	return ret, nil
}

func (orm *ORM) mongoDataFormat(arr primitive.A) (ret interface{}, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	for _, v := range arr {
		switch v := v.(type) {
		case map[string]interface{}:
			ret, err = orm.returnDataFormat(v)
			if err != nil {
				return nil, err
			}
		case primitive.A:
			ret, err = orm.mongoDataFormat(v)
			if err != nil {
				return nil, err
			}
		default:
			ret = v
		}
		break
	}

	return ret, nil
}

func (orm *ORM) toListData(target interface{}, dataValue *reflect.Value, table *Collection) (err error) {
	elemType := dataValue.Type().Elem()
	if elemType.Kind() == reflect.Struct || elemType.Kind() == reflect.Map ||
		(elemType.Kind() == reflect.Ptr && (elemType.Elem().Kind() == reflect.Struct || elemType.Elem().Kind() == reflect.Map)) {
		if orm.Q.Distinct {
			return fmt.Errorf("distinct only support simple data array, such as []number, []string")
		}

		q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
		opts := NewFindOptions()
		opts.Select(orm.Q.Select)
		if len(orm.Q.Limit) == 1 {
			opts.Limit(int64(orm.Q.Limit[0]))
		} else if len(orm.Q.Limit) == 2 {
			opts.Skip(int64(orm.Q.Limit[0]))
			opts.Limit(int64(orm.Q.Limit[1]))
		}
		opts.Projection(orm.Q.Projection)
		opts.Sort(orm.Q.Order)
		err = table.FindDocs(orm.ctx, q, target, opts)
		if err != nil {
			return err
		}
	} else {
		if len(orm.Q.Select) != 1 {
			return fmt.Errorf("must be select one field data")
		}

		if orm.Q.Distinct {
			q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
			result, err := table.Distinct(orm.ctx, orm.Q.Select[0], q, nil)
			if err != nil {
				return err
			}

			if len(result) <= 0 {
				return nil
			}

			elemList := reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(result))
			for _, d := range result {
				newData := reflect.New(elemType)
				newData = newData.Elem()

				err = orm.setDataFunc(newData, d)
				if err != nil {
					return err
				}
				elemList = reflect.Append(elemList, newData)
			}

			dataValue.Set(elemList)
			return nil
		}

		q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
		opts := NewFindOptions()
		opts.Select(orm.Q.Select)
		if len(orm.Q.Limit) == 1 {
			opts.Limit(int64(orm.Q.Limit[0]))
		} else if len(orm.Q.Limit) == 2 {
			opts.Skip(int64(orm.Q.Limit[0]))
			opts.Limit(int64(orm.Q.Limit[1]))
		}
		opts.Projection(orm.Q.Projection)
		opts.Sort(orm.Q.Order)
		var ret []map[string]interface{}
		err = table.FindDocs(orm.ctx, q, &ret, opts)
		if err != nil {
			return err
		}

		if len(ret) <= 0 {
			return nil
		}

		subKey := strings.Split(orm.Q.Select[0], ".")
		elemList := reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(ret))
		for _, m := range ret {
			newData := reflect.New(elemType)
			newData = newData.Elem()

			m, err = orm.returnDataFormat(m)
			if err != nil {
				return err
			}

			var temp = m
			for _, k := range subKey[:len(subKey)-1] {
				temp = temp[k].(map[string]interface{})
			}
			err = orm.setDataFunc(newData, temp[subKey[len(subKey)-1]])
			if err != nil {
				return err
			}
			elemList = reflect.Append(elemList, newData)
		}

		dataValue.Set(elemList)
	}

	return nil
}

func (orm *ORM) ToData(target interface{}) (err error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	dataValue := reflect.ValueOf(target)
	if nil == target || dataValue.IsNil() || dataValue.Type().Kind() != reflect.Ptr {
		return ErrTargetNotSettable
	}

	table := orm.db.Collection(orm.tableName)

	dataValue = dataValue.Elem()
	if dataValue.Type().Kind() == reflect.Slice {
		return orm.toListData(target, &dataValue, table)
	} else if dataValue.Type().Kind() == reflect.Map || dataValue.Type().Kind() == reflect.Struct {
		if orm.Q.Distinct {
			return fmt.Errorf("distinct only support simple data array, such as []number, []string")
		}

		q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
		opts := NewFindOneOptions()
		opts.Select(orm.Q.Select)
		if len(orm.Q.Limit) == 1 {
			opts.Skip(int64(orm.Q.Limit[0]))
		}
		opts.Projection(orm.Q.Projection)
		opts.Sort(orm.Q.Order)
		err = table.FindOne(orm.ctx, q, target, opts)
		if err != nil {
			return err
		}
	} else {
		if orm.Q.Distinct {
			return fmt.Errorf("distinct only support simple data array, such as []number, []string")
		}

		if len(orm.Q.Select) != 1 {
			return fmt.Errorf("must be select one field data")
		}
		q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
		opts := NewFindOneOptions()
		opts.Select(orm.Q.Select)
		if len(orm.Q.Limit) == 1 {
			opts.Skip(int64(orm.Q.Limit[0]))
		}
		opts.Projection(orm.Q.Projection)
		opts.Sort(orm.Q.Order)

		var ret map[string]interface{}
		err = table.FindOne(orm.ctx, q, &ret, opts)
		if err != nil {
			return err
		}

		if len(ret) <= 0 {
			return nil
		}

		ret, err = orm.returnDataFormat(ret)
		if err != nil {
			return err
		}

		subKey := strings.Split(orm.Q.Select[0], ".")
		var temp = ret
		for _, k := range subKey[:len(subKey)-1] {
			temp = temp[k].(map[string]interface{})
		}
		err = orm.setDataFunc(dataValue, temp[subKey[len(subKey)-1]])
		if err != nil {
			return err
		}
	}

	return err
}

func (orm *ORM) formatWhereArr(tbName string, where interface{}) interface{} {
	var newWhere []interface{}
	if arrInterfaceWhere, ok := where.([]interface{}); ok {
		for _, vv := range arrInterfaceWhere {
			switch vv := vv.(type) {
			case map[string]interface{}:
				subQ := orm.formatWhere(tbName, vv)
				newWhere = append(newWhere, subQ)
			default:
				newWhere = append(newWhere, vv)
			}
		}
	} else if arrMapWhere, ok := where.([]map[string]interface{}); ok {
		for _, vv := range arrMapWhere {
			subQ := orm.formatWhere(tbName, vv)
			newWhere = append(newWhere, subQ)
		}
	} else if mapWhere, ok := where.(map[string]interface{}); ok {
		return orm.formatWhere(tbName, mapWhere)
	}
	if len(newWhere) <= 0 {
		return where
	}
	return newWhere
}

func (orm *ORM) formatWhere(tbName string, raw Where) Where {
	var needHandleKeyList []string
	for k, v := range raw {
		if util.ElemIn(k, []string{"$and", "$or", "$nor"}) {
			raw[k] = orm.formatWhereArr(tbName, v)
		} else {
			realKey := k
			if k[0] == '~' {
				realKey = k[1:]
			}

			r := orm.refConf.getRef(tbName, realKey)
			if r != nil {
				if _, ok := v.(map[string]interface{}); !ok {
					if _, ok := v.(*refQ); !ok {
						panic(fmt.Sprintf("ref key[%s] condition type must be map[string]interface{}", k))
					} else {
						continue
					}
				}
				cond := orm.formatWhere(r.To, v.(map[string]interface{}))
				raw[k] = &tempRefQ{
					From:  tbName,
					Query: cond,
				}
				needHandleKeyList = append(needHandleKeyList, k)
			}
		}
	}
	if len(needHandleKeyList) == 1 {
		k := needHandleKeyList[0]
		if data, ok := raw[k].(*tempRefQ); ok {
			raw[k] = &refQ{
				Ref:   orm.refConf,
				From:  data.From,
				DB:    orm.db,
				Query: MixQ(data.Query),
			}
		} else {
			panic("ref query type error")
		}
	} else if len(needHandleKeyList) > 1 {
		condExecutor := task.NewTaskExecutor(" mongotool query")
		for _, k := range needHandleKeyList {
			condExecutor.AddFixed(func(param ...interface{}) (interface{}, error) {
				cond := param[0].(Where)
				key := param[1].(string)
				if data, ok := cond[key].(*tempRefQ); ok {
					cond[key] = &refQ{
						Ref:   orm.refConf,
						From:  data.From,
						DB:    orm.db,
						Query: MixQ(data.Query),
					}
				} else {
					return nil, fmt.Errorf("ref query type error")
				}
				return nil, nil
			}, raw, k)
		}
		_, err := condExecutor.Execute(orm.ctx)
		if err != nil {
			panic(err)
		}
	}
	return raw
}

func (orm *ORM) Count(clearCache bool) (int64, error) {
	if clearCache {
		defer func() {
			orm.ClearCache()
		}()
	}
	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))

	count, err := table.Count(orm.ctx, q, nil)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (orm *ORM) InsertOne(data interface{}) (string, error) {
	table := orm.db.Collection(orm.tableName)
	var m map[string]interface{}
	switch data := data.(type) {
	case map[string]interface{}:
		m = map[string]interface{}{}
		for k, v := range data {
			m[k] = v
		}
	default:
		m = Struct2Map(data)
	}

	if id, ok := m["_id"]; ok {
		switch id := id.(type) {
		case string:
			if id == "" {
				delete(m, "_id")
			} else {
				m["_id"] = TryString2ObjectID(id)
			}
		case ObjectID:
			if id.IsZero() {
				delete(m, "_id")
			}
		case [12]byte:
			var objID ObjectID
			copy(objID[:], id[:])
			if objID.IsZero() {
				delete(m, "_id")
			}
		}
	}
	return table.InsertDoc(orm.ctx, m)
}

func (orm *ORM) InsertMany(data []interface{}, ordered bool) ([]string, error) {
	table := orm.db.Collection(orm.tableName)

	var insertDataList []interface{}
	for _, v := range data {
		var m map[string]interface{}
		switch v := v.(type) {
		case map[string]interface{}:
			m = map[string]interface{}{}
			for k, v := range v {
				m[k] = v
			}
		default:
			m = Struct2Map(v)
		}

		if id, ok := m["_id"]; ok {
			switch id := id.(type) {
			case string:
				if id == "" {
					delete(m, "_id")
				} else {
					m["_id"] = TryString2ObjectID(id)
				}
			case ObjectID:
				if id.IsZero() {
					delete(m, "_id")
				}
			case [12]byte:
				var objID ObjectID
				copy(objID[:], id[:])
				if objID.IsZero() {
					delete(m, "_id")
				}
			}
		}
		insertDataList = append(insertDataList, m)
	}

	return table.InsertDocs(orm.ctx, insertDataList, ordered)
}

func (orm *ORM) UpdateOne(data map[string]interface{}, upsert bool) (*UpdateResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opt := NewUpdate()
	opt.Upsert(upsert)
	return table.UpdateOne(orm.ctx, q, data, opt)
}

func (orm *ORM) UpdateMany(data map[string]interface{}, upsert bool) (*UpdateResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opt := NewUpdate()
	opt.Upsert(upsert)
	return table.UpdateMany(orm.ctx, q, data, opt)
}

func (orm *ORM) UpdateOneCustom(update updateType, data map[string]interface{}, upsert bool) (*UpdateResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opt := NewUpdate()
	opt.Upsert(upsert)
	return table.UpdateOneCustom(orm.ctx, q, update, data, opt)
}

func (orm *ORM) UpdateManyCustom(update updateType, data map[string]interface{}, upsert bool) (*UpdateResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opt := NewUpdate()
	opt.Upsert(upsert)
	return table.UpdateManyCustom(orm.ctx, q, update, data, opt)
}

func (orm *ORM) DeleteOne() (*DeleteResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	return table.DeleteOne(orm.ctx, q)
}

func (orm *ORM) DeleteMany() (*DeleteResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	return table.DeleteMany(orm.ctx, q)
}

func (orm *ORM) BulkWrite(bwm *BulkWriteModel) (*BulkWriteResult, error) {
	table := orm.db.Collection(orm.tableName)
	return table.BulkWrite(orm.ctx, bwm)
}

func (orm *ORM) ReplaceOne(data map[string]interface{}, upsert bool) (*UpdateResult, error) {
	if !orm.keepQuery {
		defer func() {
			orm.ClearCache()
		}()
	}

	table := orm.db.Collection(orm.tableName)
	q := MixQ(orm.formatWhere(orm.tableName, orm.Q.Where))
	opt := NewReplace()
	opt.Upsert(upsert)
	return table.ReplaceOne(orm.ctx, q, data, opt)
}

func (orm *ORM) Client() *Client {
	return orm.db.Client
}

func (orm *ORM) Collection() *Collection {
	table := orm.db.Collection(orm.tableName)
	return table
}

func (orm *ORM) Database() *Database {
	return orm.db
}
