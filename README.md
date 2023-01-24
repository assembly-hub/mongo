# mongo

### mongo 算子

```
=: "key": "val" or "key__eq": "val"
<: "key__lt": 1
<=: "key__lte": 1
>: "key__gt": 1
>=: "key__gte": 1
!=: "key__ne": 1
in: "key__in": [1]
all: "key__all": [1, 2, 3]
not in: "key__nin": [1]
size: "arr__size": 1
exists: "key__exists": true
mod: "key__mod": [10, 1], 基数，余数
elemMatch: "key__match": MongoQuery
$or: map[string]interface{} or []MixQ or []interface or []map[string]interface{}
$and: map[string]interface{} or []MixQ or []interface or []map[string]interface{}
like:
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
```

## 快速开始

go get github.com/assembly-hub/mongo