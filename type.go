// package mongo
package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

type ObjectID = primitive.ObjectID

var emptyObjID = [12]byte{}

var emptyObject = string(emptyObjID[:])

type objectIDType interface {
	string | ObjectID
}

func ObjectIDIsZero[T objectIDType](id T) bool {
	var i interface{} = id
	switch i := i.(type) {
	case string:
		return i == "" || i == emptyObject
	case ObjectID:
		return i == emptyObjID
	default:
		panic("id type must be string or ObjectID")
	}
}
