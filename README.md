# ORM使用说明

## 一、初始化

### 表定义

```go
// 根据数据库类型定义表关系对象
var ref = mongo.NewReference()

/**
  注意：
      bson：mongo数据标签必须存在，查询使用此标签
      json：序列化json使用
      ref：外键关联方式（查询出来的数据与外键数据比较方式）：
          def: 有交集即可
          all：所有的外键包含在查询出来的数据
          match：外键必须要与查询出来的数据完全匹配
*/

// 定义表结构
type Table1 struct {
// mongo 主键，bson 必须是 _id
ID   mongo.ObjectID            `bson:"_id" json:"key"`
Txt  string                    `bson:"txt" json:"txt"`
Ref  *mongo.Foreign[Table2]    `bson:"ref" json:"ref" ref:"def"`
Ref2 mongo.ForeignList[Table3] `bson:"ref2" json:"ref2" ref:"match"`
}
// 添加表定义（建议放在表声明的go文件的 init 方法中）
ref.AddTableDef("table1", dao.Table1{})

type Table2 struct {
ID   string                 `bson:"_id" json:"key"`
Name string                 `bson:"name" json:"name"`
Ref  *mongo.Foreign[Table3] `bson:"ref" json:"ref" ref:"def"`
}
ref.AddTableDef("table2", dao.Table2{})

type Table3 struct {
ID  string `bson:"_id" json:"key"`
Txt string `bson:"txt" json:"txt"`
}
ref.AddTableDef("table3", dao.Table3{})

// 编译整个数据库的表关系
ref.BuildRefs()
```

**注：以上仅仅在项目启动执行一次，切勿在业务代码中执行调用**

## 二、查询算子（每个算子前面需要用双下划线标注）

```go
// 获取链接配置
opts := mongo.OptionsFromURI("mongodb://localhost:27017")
// 创建client
client, err := mongo.NewClient(context.Background(), opts)
if err != nil {
fmt.Println("get mongo client error")
panic(err)
}
// 链接数据库
db := client.Database("example")
ctx := context.Background()

// 创建表对象
tb1 := mongo.NewORMByDB(ctx, db, "table1", mongoRef)
```

**查询有：Query Where Wheres方法，一下以 Where 为演示使用**

**注意：所有的 `key` 值，必须和 `bson` 一致**

### 1、eq 等于（唯一一个可以省略的算子）

> tb1.Where("key__eq", 1) or tb1.Where("key", 1)

### 2、ne 不等于

> tb1.Where("key__ne", 1)

### 3、lt 小于

> tb1.Where("key__lt", 1)

### 4、lte 小于等于

> tb1.Where("key__lte", 1)

### 5、gt 大于

> tb1.Where("key__gt", 1)

### 6、gte 大于等于

> tb1.Where("key__gte", 1)

### 7、in 包含

> tb1.Where("key__in", []int{1,2,3})

说明

```sql
key 可以是数组和值类型
key 与 []int{1,2,3} 存在交集即可
```

### 8、nin 不包含

> tb1.Where("key__nin", []int{1,2,3})

说明

```sql
key 可以是数组和值类型
key 与 []int{1,2,3} 不存在交集
```

### 9、all 全匹配

> tb1.Where("key__all", []string{"1","2","3"})

说明

```sql
key 是数组
条件当中的元素必须都在key中可以找到
```

### 10、size 数组长度判断

> tb1.Where("arr__size", 1)

说明

```sql
判断数组的长度
```

### 11、exists 字段存在判断

> tb1.Where("test__exists", true)

说明

```sql
判断指定的key是否存在
```

### 12、mod 取余

> tb1.Where("num__mod", []int{10, 1})

说明

```sql
判断余数是否满足要求
num % 10 == 1
```

### 13、match 文档子元素判断

> tb1.Where("arr__match", MixQ(map[string]interface{}{
"key": 1,
"name__in": []string{"test", "test2"},
> }))

说明

```sql
判断arr
元素字段
，需要满足
：
elem.id = 1 and elem.name in ["test", "test2"]
```

### 14、istartswith(startswith) 字符串开始匹配

> tb1.Where("txt__istartswith", "test")

说明

```sql
判断字符串是否以给定的条件开始
`i`为`ignore`的简写
，为忽略大小写
```

### 15、iendswith(endswith) 字符串结束匹配

> tb1.Where("txt__iendswith", "test")

说明

```sql
判断字符串是否以给定的条件结束
`i`为`ignore`的简写
，为忽略大小写
```

### 16、icontains(contains) 字符串包含匹配

> tb1.Where("txt__icontains", "test")

说明

```sql
判断字符串是否包含给定的条件
`i`为`ignore`的简写
，为忽略大小写
```

***注：$or $nor $and内部即可以包含基础，也可以嵌套 $or $nor $and，以下以 $or 举例***

### 20、$or 或，其内部为基础查询，参数可以是map、[]map、[]MixQ

```go
tb1.Where("$or", map[string]interface{}{
"key__gt": 0,
"name__startswith": "str",
})
```

说明

```sql
key大于0
或 name 以 "str" 开始
```

```go
tb1.Where("$or", []map[string]interface{}{
{
"key__gt": 0,
"name__startswith": "str",
},{
"key__lt": 0,
"name": "string",
}
})
```

说明

```sql
（key大于0
且 name 以 "str" 开始
）或者
（key小于0 且 name等于"string"
）
```

### 21、 ~ 取反操作符

> ~ 为条件取反，必须在最前面，可用在所有前面，如果与#连用，#应在~后面，如：~#test

> tb1.Where("~id__lt", 1)

说明

```sql
检索id不小于1的数据
```

### 22、 geo_within_polygon 闭合多边形检索

> tb1.Where("geo__geo_within_polygon", [][]float64{需要一个闭合的多边形坐标集合})

说明

```sql
检索在多边形内部的数据
```

### 23、 geo_within_multi_polygon 多个闭合多边形检索

> tb1.Where("geo__geo_within_multi_polygon", [][][]float64{需要多个闭合的多边形坐标集合})

说明

```sql
检索在多边形内部的数据
```

### 24、 geo_within_center_sphere 原型检索

> tb1.Where("geo__geo_within_center_sphere", []float64{})

> []float64, 长度必须是3，分辨为 lng, lat, radius(米)

说明

```sql
检索原型内部的数据
```

### 25、 geo_intersects_polygon 相交判断

> tb1.Where("geo__geo_intersects_polygon", [][]float64{需要一个闭合的多边形坐标集合})

说明

```sql
检索相交的数据
```

### 26、 其他geo算子

```
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
    key__geo_within_2d_center: []float64, 长度必须是3，分别为 lng, lat, radius(米)
```

## 三、select 查询

> 支持字段显示隐藏的控制

> tb1.Select("txt", "-name")

> -name: 隐藏name，txt或+txt: 显示txt

## 四、order by

> tb1.Order("txt", "-name")

说明

```sql
按照txt升序
，
name降序展示数据
```

## 五、ORM 函数介绍

### 1、Query Where Wheres

> 查询条件函数

### 2、OverLimit(over, size uint)

> 配置 limit 和 offset

### 3、Page(pageNo, pageSize uint)

> 传入页码和每页的大小，框架自定转化为说明数据库的数据查询条件

### 4、Limit(size uint)

> 设置limit

### 5、Distinct

> 设置是否去重

### 6、Select 配置字段查询

### 7、Order 配置排序字段

### 8、ToData(result interface{}) 万能数据接收接口

```go
其中 result 为数据指针，数据类型如下：
1、简单类型：int、string、uint等
2、简单切片：[]int []string []uint等
3、map类型：map[string]interface{}、map[string]int、map[string]string等
4、map切片：[]map[string]interface{}、[]map[string]int、[]map[string]string等
5、struct：Table1
6、struct切片：[]Table1
```

> demo

```go
var res []Table1
tb1.ToDate(&res)
```

### 9、PageData(result interface{}, pageNo, pageSize uint) (pg *Paging, err error) 获取某一页的数据

> 参数与ToData一致

```go
type Paging struct {
PageNo    int `json:"page_no"`   //当前页
PageSize  int `json:"page_size"` //每页条数
Total     int `json:"total"`      //总条数
PageTotal int `json:"page_total"` //总页数
}
```

### 10、Exist

> 检查是否有数据

### 11、Count 获取数据条数

### 12、数据插入

#### 1、InsertOne

> 数据类型可以是 map[string]interface{} 或 struct

#### 2、InsertMany

> 参数可以是 map 与 struct 混合的数组

### 13、数据更新或插入

#### 1、UpsertOne UpdateOneCustom

> 数据类型可以是 map[string]interface{} 或 struct

#### 2、UpsertMany UpdateManyCustom

> 参数可以是 map 与 struct 混合的数组

## 六、事务 orm.TransSession

```go
// 事务，要求mongo版本>4.0，需要mongo副本集群
// 返回nil事务执行成功
err = mongo.TransSession(client, func (sessionCtx mongo.SessionContext) error {
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
```

## 七、其他

### 1、mongo.Struct2Map

可以根据要求将struct转成map，过滤ref，格式化json自定义数据

## 八、结语

有问题随时留言，vx：lm2586127191