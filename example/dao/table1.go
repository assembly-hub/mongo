package dao

import "github.com/assembly-hub/mongo"

/**
bson：mongo数据标签必须存在，查询使用此标签
json：序列化json使用
ref：外键关联方式（查询出来的数据与外键数据比较方式）：
	def: 有交集即可
	all：所有的外键包含在查询出来的数据
	match：外键必须要与查询出来的数据完全匹配
*/

type Table1 struct {
	// mongo 主键，bson 必须是 _id
	ID   mongo.ObjectID            `bson:"_id" json:"id"`
	Txt  string                    `bson:"txt" json:"txt"`
	Ref  *mongo.Foreign[Table2]    `bson:"ref" json:"ref" ref:"def"`
	Ref2 mongo.ForeignList[Table3] `bson:"ref2" json:"ref2" ref:"match"`
}
