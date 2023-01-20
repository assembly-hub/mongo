package mongo

import (
	"fmt"
	"testing"
)

func TestObjectIDIsZero(t *testing.T) {
	objID := ObjectID{}
	strID1 := ""
	strID2 := string(objID[:])

	fmt.Println(ObjectIDIsZero(objID))
	fmt.Println(ObjectIDIsZero(strID1))
	fmt.Println(ObjectIDIsZero(strID2))
}
