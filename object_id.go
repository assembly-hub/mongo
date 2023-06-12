// Package mongo
package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ObjectID2String object id 转 string
func ObjectID2String(v ObjectID) string {
	return v.Hex()
}

// String2ObjectID string 转 object id
func String2ObjectID(v string) (hex ObjectID, err error) {
	hex, err = primitive.ObjectIDFromHex(v)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return hex, nil
}

func TryString2ObjectID(v string) ObjectID {
	hex, err := primitive.ObjectIDFromHex(v)
	if err != nil {
		panic(err)
	}

	return hex
}

// NewObjectID 生成ObjectID
func NewObjectID() ObjectID {
	obj := primitive.NewObjectID()
	return obj
}
