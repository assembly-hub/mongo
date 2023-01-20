package dao

import "github.com/assembly-hub/mongo"

type Table2 struct {
	ID   string                 `bson:"_id" json:"id"`
	Name string                 `bson:"name" json:"name"`
	Ref  *mongo.Foreign[Table3] `bson:"ref" json:"ref" ref:"def"`
}
