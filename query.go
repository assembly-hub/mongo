// Package mongo
package mongo

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/assembly-hub/basics/util"
)

type queryNode struct {
	Key   string
	Value interface{}
}

// Query
// = : "key": "val" or "key__eq": "val"
// < : "key__lt": 1
// <= : "key__lte": 1
// > : "key__gt": 1
// >= : "key__gte": 1
// != : "key__ne": 1
// in : "key__in": [1]
// all : "key__all": [1, 2, 3]
// not in : "key__nin": [1]
// size : "arr__size": 1
// exists : "key__exists": true
// mod : "key__mod": [10, 1], 基数，余数
// elemMatch : "key__match": MongoQuery
// like :
//
//	"key__istartswith": "123"
//	"key__startswith": "123"
//	"key__iendswith": "123"
//	"key__endswith": "123"
//	"key__icontains": "123"
//	"key__contains": "123"
//
// geo within query 2dsphere
//
//	key__geo_within_polygon: [][]float64 or [][][]float64 闭合的多边形
//	key__geo_within_multi_polygon: [][][]float64 or [][][][]float64 多个闭合的多边形
//	key__geo_within_center_sphere: []float64, 长度必须是3，分辨为 lng, lat, radius(米)
//
// geo intersects query 2dsphere
//
//	key__geo_intersects_polygon: [][]float64 or [][][]float64 闭合的多边形
//
// geo near query 2dsphere
//
//			key__near: []float64 坐标点 or map[string]interface{}{
//				"point": []float64 坐标点,
//	            "min": 最小距离，单位：米，
//	            "max": 最大距离，单位：米
//			}
//			key__near_sphere: []float64 坐标点 or map[string]interface{}{
//				"point": []float64 坐标点,
//	            "min": 最小距离，单位：米，
//	            "max": 最大距离，单位：米
//			}
//
// geo within query 2d
//
//	key__geo_within_2d_box: [][]float64, bottom left, top right
//	key__geo_within_2d_polygon: [][]float64 非闭合的多边形
//	key__geo_within_2d_center: []float64, 长度必须是3，分辨为 lng, lat, radius(米)
type Query struct {
	nodes   []queryNode
	rawCond map[string]interface{}
}

func NewQuery() *Query {
	q := new(Query)
	q.nodes = []queryNode{}
	q.rawCond = map[string]interface{}{}
	return q
}

func NotQ(key string, value interface{}) *Query {
	rawQ := Q(key, value)
	if len(rawQ.nodes) <= 0 {
		panic(fmt.Sprintf("MongoNotQ key:%s, value:%v", key, value))
	}

	q := NewQuery()
	if len(rawQ.nodes) == 1 {
		q.nodes = []queryNode{{
			Key:   "$not",
			Value: rawQ.nodes[0],
		}}
	} else {
		var arrQ []queryNode
		for _, subQ := range rawQ.nodes {
			arrQ = append(arrQ, queryNode{
				Key:   "$not",
				Value: subQ,
			})
		}
		q.nodes = []queryNode{{
			Key:   "$or",
			Value: arrQ,
		}}
	}

	return q
}

func idFormat(value interface{}) interface{} {
	switch val := value.(type) {
	case string:
		value = TryString2ObjectID(val)
	case []string:
		newVal := make([]interface{}, len(val))
		for i, v := range val {
			newVal[i] = TryString2ObjectID(v)
		}
		value = newVal
	case [][]byte:
		newVal := make([]interface{}, len(val))
		for i, v := range val {
			newVal[i] = TryString2ObjectID(string(v))
		}
		value = newVal
	case []interface{}:
		newVal := make([]interface{}, len(val))
		for i, v := range val {
			switch v := v.(type) {
			case string:
				newVal[i] = TryString2ObjectID(v)
			case []byte:
				newVal[i] = TryString2ObjectID(string(v))
			default:
				newVal[i] = v
			}
		}
		value = newVal
	}

	return value
}

func generateForeignKeyCondition(key string, value interface{}) *Query {
	if value != nil {
		var q *refQ
		switch value := value.(type) {
		case refQ:
			q = &value
		case *refQ:
			q = value
		}
		if q != nil {
			if strings.Contains(key, "__") {
				panic("refQ cols name can not contains '__'")
			}

			val, tp := q.getData(key)
			mq := NewQuery()
			if tp == mongoRefMatch {
				mq.nodes = []queryNode{{
					Key:   key + ".$id__all",
					Value: val,
				}, {
					Key:   key + "__size",
					Value: len(val),
				}}
			} else if tp == mongoRefAll {
				mq.nodes = []queryNode{{
					Key:   key + ".$id__all",
					Value: val,
				}}
			} else if tp == mongoRefDefault {
				mq.nodes = []queryNode{{
					Key:   key + ".$id__in",
					Value: val,
				}}
			} else {
				panic("mongotool ref type error")
			}

			return mq
		}
	}

	return nil
}

func geoGeoComplexCondition(key string, keys []string, value interface{}) (string, interface{}, bool) {
	has := true
	switch keys[1] {
	case "geo_within_polygon":
		var coordinates [][][]float64
		switch value := value.(type) {
		case [][]float64:
			coordinates = [][][]float64{value}
		case [][][]float64:
			coordinates = value
		default:
			panic("type of geo_within_polygon's value must be [][]float64 or [][][]float64")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$geometry": map[string]interface{}{
				"type":        "Polygon",
				"coordinates": coordinates,
			},
		}
	case "geo_within_multi_polygon":
		var coordinates [][][][]float64
		switch value := value.(type) {
		case [][][]float64:
			coordinates = [][][][]float64{value}
		case [][][][]float64:
			coordinates = value
		default:
			panic("type of geo_within_multi_polygon's value must be [][][]float64 or [][][][]float64")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$geometry": map[string]interface{}{
				"type":        "MultiPolygon",
				"coordinates": coordinates,
			},
		}
	case "geo_within_center_sphere":
		var floatArr []float64
		switch value := value.(type) {
		case []float64:
			floatArr = value
		default:
			panic("type of geo_within_center_sphere's value must be []float64 and length need 3")
		}
		if len(floatArr) != 3 {
			panic("type of geo_within_center_sphere's value must be []float64 and length need 3")
		}
		if floatArr[2] <= 0 {
			panic("geo_within_center_sphere radius must be gt 0")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$centerSphere": []interface{}{[]float64{floatArr[0], floatArr[1]}, floatArr[2] / 1609.344 / 3963.2},
		}
	case "geo_intersects_polygon":
		var coordinates [][][]float64
		switch value := value.(type) {
		case [][]float64:
			coordinates = [][][]float64{value}
		case [][][]float64:
			coordinates = value
		default:
			panic("type of geo_within_polygon's value must be [][]float64 or [][][]float64")
		}
		key = keys[0] + "__geoIntersects"
		value = map[string]interface{}{
			"$geometry": map[string]interface{}{
				"type":        "Polygon",
				"coordinates": coordinates,
			},
		}
	case "geo_within_2d_box":
		var box [][]float64
		switch value := value.(type) {
		case [][]float64:
			if len(value) != 2 {
				panic("type of geo_within_2d_box's value length is 2")
			}
		default:
			panic("type of geo_within_2d_box's value must be [][]float64")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$box": box,
		}
	case "geo_within_2d_polygon":
		var polygon [][]float64
		switch value := value.(type) {
		case [][]float64:
			if len(value) < 3 {
				panic("type of geo_within_2d_polygon's value length gte 3")
			}
		default:
			panic("type of geo_within_2d_polygon's value must be [][]float64")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$polygon": polygon,
		}
	case "geo_within_2d_center":
		var floatArr []float64
		switch value := value.(type) {
		case []float64:
			floatArr = value
		default:
			panic("type of geo_within_2d_center's value must be []float64 and length need 3")
		}
		if len(floatArr) != 3 {
			panic("type of geo_within_2d_center's value must be []float64 and length need 3")
		}
		if floatArr[2] <= 0 {
			panic("geo_within_2d_center radius must be gt 0")
		}
		key = keys[0] + "__geoWithin"
		value = map[string]interface{}{
			"$center": []interface{}{[]float64{floatArr[0], floatArr[1]}, floatArr[2] / 1609.344 / 3963.2},
		}
	default:
		has = false
	}
	return key, value, has
}

func geoGeoCondition(key string, keys []string, value interface{}) (string, interface{}, bool) {
	key, value, has := geoGeoComplexCondition(key, keys, value)
	if has {
		return key, value, has
	}

	has = true

	switch keys[1] {
	case "near":
		var point []float64
		disMap := map[string]interface{}{}
		switch value := value.(type) {
		case []float64:
			if len(value) != 2 {
				panic("type of near's point length need 2")
			}
		case map[string]interface{}:
			if min, ok := value["min"]; ok {
				disMap["$minDistance"] = min
			}
			if max, ok := value["max"]; ok {
				disMap["maxDistance"] = max
			}
			point = value["point"].([]float64)
			if len(point) != 2 {
				panic("type of near's point length need 2")
			}
		default:
			panic("type of near's value must be []float64 or map[string]interface{}")
		}
		key = keys[0] + "__near"
		temp := map[string]interface{}{
			"$geometry": map[string]interface{}{
				"type":        "Point",
				"coordinates": point,
			},
		}
		for disKey, disVal := range disMap {
			temp[disKey] = disVal
		}
		value = temp
	case "near_sphere":
		var point []float64
		disMap := map[string]interface{}{}
		switch value := value.(type) {
		case []float64:
			if len(value) != 2 {
				panic("type of near's point length need 2")
			}
		case map[string]interface{}:
			if min, ok := value["min"]; ok {
				disMap["$minDistance"] = min
			}
			if max, ok := value["max"]; ok {
				disMap["maxDistance"] = max
			}
			point = value["point"].([]float64)
			if len(point) != 2 {
				panic("type of near's point length need 2")
			}
		default:
			panic("type of near's value must be []float64 or map[string]interface{}")
		}
		key = keys[0] + "__nearSphere"
		temp := map[string]interface{}{
			"$geometry": map[string]interface{}{
				"type":        "Point",
				"coordinates": point,
			},
		}
		for disKey, disVal := range disMap {
			temp[disKey] = disVal
		}
		value = temp
	default:
		has = false
	}
	return key, value, has
}

func geoCondition(key string, keys []string, value interface{}) (string, interface{}) {
	key, value, has := geoGeoCondition(key, keys, value)
	if has {
		return key, value
	}

	switch keys[1] {
	case "istartswith":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex":   "^" + regexp.QuoteMeta(value.(string)) + ".*",
			"$options": "i",
		}
	case "startswith":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex": "^" + regexp.QuoteMeta(value.(string)) + ".*",
		}
	case "iendswith":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex":   ".*" + regexp.QuoteMeta(value.(string)) + "$",
			"$options": "i",
		}
	case "endswith":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex": ".*" + regexp.QuoteMeta(value.(string)) + "$",
		}
	case "icontains":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex":   ".*" + regexp.QuoteMeta(value.(string)) + ".*",
			"$options": "i",
		}
	case "contains":
		key = keys[0] + "__regex"
		value = map[string]interface{}{
			"$regex": ".*" + regexp.QuoteMeta(value.(string)) + ".*",
		}
	case "match":
		if _, ok := value.(*Query); !ok {
			panic("operator[match]'s value type must be MongoQuery' pointer")
		}
		key = keys[0] + "__elemMatch"
		c := value.(*Query).Cond()
		if len(c) <= 0 {
			panic("match where cond is nil")
		}
		value = c
	}

	return key, value
}

func Q(key string, value interface{}) *Query {
	q := generateForeignKeyCondition(key, value)
	if q != nil {
		return q
	}

	if strings.Contains(key, "__") {
		keys := strings.Split(key, "__")
		if keys[0] == "_id" || util.EndWith(keys[0], ".$id", false) {
			value = idFormat(value)
		}

		key, value = geoCondition(key, keys, value)
	} else {
		if key == "_id" || util.EndWith(key, ".$id", false) {
			value = idFormat(value)
		}
	}

	q = NewQuery()
	q.nodes = []queryNode{{
		Key:   key,
		Value: value,
	}}
	return q
}

// MixQ
// ~: 结果取反，其中：$and、$or、$nor不适应
// $and、$or、$nor: values type must be map[string]interface{} or []*MongoQuery or []interface or []map[string]interface{}
func MixQ(cond map[string]interface{}) *Query {
	if len(cond) <= 0 {
		return NewQuery()
	}

	q := NewQuery()
	for k, v := range cond {
		if k == "" {
			continue
		}

		if k[0] == '~' {
			if util.ElemIn(k[1:], []string{"$and", "$or", "$nor"}) {
				panic("$and、$or、$nor not supported '~'")
			}
			q.And(NotQ(k[1:], v))
		} else if k == "$and" || k == "$or" || k == "$nor" {
			var arrQuery []interface{}
			switch v := v.(type) {
			case []interface{}:
				arrQuery = append(arrQuery, v...)
			case []map[string]interface{}:
				for _, v := range v {
					arrQuery = append(arrQuery, v)
				}
			case []*Query:
				for _, v := range v {
					arrQuery = append(arrQuery, v)
				}
			case map[string]interface{}:
				if k == "$and" {
					arrQuery = append(arrQuery, v)
				} else {
					for subKey, subVal := range v {
						arrQuery = append(arrQuery, map[string]interface{}{
							subKey: subVal,
						})
					}
				}
			default:
				panic("$and、$or、$nor: values type must be []map[string]interface{} or []*MongoQuery or mix type")
			}
			if len(arrQuery) <= 0 {
				continue
			} else {
				var arr []*Query
				for _, vv := range arrQuery {
					switch vv := vv.(type) {
					case map[string]interface{}:
						subQ := MixQ(vv)
						if !subQ.Empty() {
							arr = append(arr, subQ)
						}
					case *Query:
						if !vv.Empty() {
							arr = append(arr, vv)
						}
					default:
						panic("$and、$or、$nor value must be map[string]interface{} or *MongoQuery")
					}
				}

				if len(arr) <= 0 {
					continue
				}

				if k == "$and" {
					q.And(arr...)
				} else if k == "$or" {
					q.Or(arr...)
				} else if k == "$nor" {
					q.Nor(arr...)
				}
			}
		} else {
			q.And(Q(k, v))
		}
	}
	return q
}

// SetRawCond cond default nil means is not HQL and cond is not nil means use MongoQuery
// cond nil close raw query
func (q *Query) SetRawCond(cond map[string]interface{}) *Query {
	q.rawCond = cond
	return q
}

func (q *Query) And(filter ...*Query) *Query {
	for _, f := range filter {
		q.nodes = append(q.nodes, f.nodes...)
	}
	return q
}

func NewAnd(filter ...*Query) *Query {
	q := NewQuery()
	q.And(filter...)
	return q
}

func (q *Query) Empty() bool {
	return len(q.rawCond) <= 0 && len(q.nodes) <= 0
}

func (q *Query) orNor(filter ...*Query) []queryNode {
	var ns []queryNode
	for _, f := range filter {
		m := f.Cond()
		if len(m) <= 0 {
			continue
		}

		var v []map[string]interface{}
		if _, ok := m["$and"]; ok && len(m) == 1 {
			v = m["$and"].([]map[string]interface{})
		} else {
			v = []map[string]interface{}{
				m,
			}
		}
		ns = append(ns, queryNode{
			Key:   "$and",
			Value: v,
		})
	}
	return ns
}

func (q *Query) Or(filter ...*Query) *Query {
	var ns = q.orNor(filter...)
	if len(ns) <= 0 {
		return q
	}

	node := queryNode{
		Key:   "$or",
		Value: ns,
	}
	q.nodes = append(q.nodes, node)
	return q
}

func NewOr(filter ...*Query) *Query {
	q := NewQuery()
	q.Or(filter...)
	return q
}

func (q *Query) Nor(filter ...*Query) *Query {
	var ns = q.orNor(filter...)
	if len(ns) <= 0 {
		return q
	}

	node := queryNode{
		Key:   "$nor",
		Value: ns,
	}
	q.nodes = append(q.nodes, node)
	return q
}

func NewNor(filter ...*Query) *Query {
	q := NewQuery()
	q.Nor(filter...)
	return q
}

func (q *Query) innerOrNorFilter(nodeList []queryNode) []map[string]interface{} {
	var filterList []map[string]interface{}
	for _, node := range nodeList {
		m := q.innerFilter([]queryNode{node})
		filterList = append(filterList, m)
	}
	return filterList
}

func (q *Query) innerNodeFilter(node queryNode) (k string, v interface{}) {
	if node.Key == "$or" || node.Key == "$nor" {
		orNor := q.innerOrNorFilter(node.Value.([]queryNode))
		return node.Key, orNor
	} else if node.Key == "$not" {
		key, val := q.innerNodeFilter(node.Value.(queryNode))
		return key, map[string]interface{}{
			"$not": val,
		}
	} else if node.Key == "$and" {
		return node.Key, node.Value
	} else if strings.Contains(node.Key, "__") {
		keys := strings.Split(node.Key, "__")
		if keys[1] == "regex" {
			return keys[0], node.Value
		}
		return keys[0], map[string]interface{}{
			"$" + keys[1]: node.Value,
		}
	}
	return node.Key, map[string]interface{}{
		"$eq": node.Value,
	}
}

func (q *Query) innerFilter(nodes []queryNode) map[string]interface{} {
	filter := map[string]interface{}{}
	for _, node := range nodes {
		k, v := q.innerNodeFilter(node)
		if k == "$or" || k == "$nor" {
			if andQ, ok := filter["$and"]; ok {
				andQ = append(andQ.([]map[string]interface{}), map[string]interface{}{
					k: v,
				})
				filter["$and"] = andQ
			} else {
				filter["$and"] = []map[string]interface{}{{
					k: v,
				}}
			}
		} else {
			if mQ, ok := filter[k]; ok {
				for c, d := range v.(map[string]interface{}) {
					if c == "$not" {
						old := mQ.(map[string]interface{})[c]
						if old != nil {
							for kk, vv := range old.(map[string]interface{}) {
								d.(map[string]interface{})[kk] = vv
							}
						}
					}
					mQ.(map[string]interface{})[c] = d
				}
				filter[k] = mQ
			} else {
				filter[k] = v
			}
		}
	}
	return filter
}

func (q *Query) Cond() map[string]interface{} {
	if len(q.rawCond) > 0 {
		return q.rawCond
	}

	filter := q.innerFilter(q.nodes)
	return filter
}

func (q *Query) JSON() string {
	marshal, err := json.Marshal(q.Cond())
	if err != nil {
		return ""
	}
	return string(marshal)
}
