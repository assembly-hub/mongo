// package mongo
package mongo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SimpleNewMongoQuery() {
	q := NewQuery()
	q.And(Q("k1__gt", 1), Q("k2", 1))
	q.Or(Q("k__gt", 1), Q("k__ne", 2))
	q.Or(Q("k1", 1))

	q1 := NewQuery()
	q1.And(Q("k3__gt", 1), Q("k4", 1))

	q2 := NewQuery()
	q2.And(Q("k5__gt", 1), Q("k6", 1))

	q.And(q1).Or(q2)

	fmt.Println(q.Cond())
	fmt.Println(q.JSON())
}

func SimpleNewMongoQuery2() {
	q := NewQuery()
	q.And(Q("k1__gt", 1), Q("k2", 1)).Or(
		Q("k", 1), Q("K__size", 123)).And(
		Q("kkk__gt", 1))
	q.Or(Q("key111__istartswith", "test"), Q("key111__startswith", "test"))

	q.And(Q("key123__iendswith", "123")).And(Q("key222__icontains", "345"))

	fmt.Println(q.Cond())
	fmt.Println(q.JSON())
}

func SimpleNewMongoQuery3() {
	q := NewQuery().And(Q("k__eq", 1), Q("k__eq", 11))
	q.And(Q("k1__gt", 1), Q("k2", 1)).Or(
		Q("k", 1), Q("K__size", 123)).And(
		Q("kkk__gt", 1))
	q.Or(NotQ("key111__istartswith", "test"), Q("key111__startswith", "test"))

	q.And(Q("key123__iendswith", "123")).And(Q("key222__icontains", "345"))

	q.And(NotQ("arr__match", Q("age__gt", 1).And(Q("name__icontains", "test"))))

	q.And(NotQ("k", 2), NotQ("k__gt", 3))

	q.And(Q("test__exists", true))

	fmt.Println(q.Cond())
	fmt.Println(q.JSON())
}

func SimpleMongoQ() {
	q := NewQuery()
	q.And(Q("is_valid", true))
	q.And(Q("city_id", 11))

	openGeoQ := NewQuery()
	openGeoQ.And(
		NewOr(
			Q("robobus.open", 1),
			Q("minibus.open", 1)),
	)

	whiteProjects := []string{"1", "2"}
	whiteQ := NewQuery().And(
		Q("project_id__in", whiteProjects),
		NewOr(
			Q("robobus.open", 2),
			Q("minibus.open", 2)))
	openGeoQ = NewOr(openGeoQ, whiteQ)

	q.And(openGeoQ)

	s := q.JSON()
	fmt.Println(s)
}

func SimpleMixQ() {
	whiteProjects := []string{"1", "2"}
	whiteQ := MixQ(map[string]interface{}{
		"project_id__in": whiteProjects,
		"$or": []map[string]interface{}{
			{
				"robobus.open": 2,
			},
			{
				"minibus.open": 2,
			},
		},
	})

	openGeoQ := MixQ(map[string]interface{}{
		"$or": []interface{}{
			map[string]interface{}{
				"robobus.open": 1,
				"type":         2,
			},
			map[string]interface{}{
				"minibus.open": 1,
				"$and": map[string]interface{}{
					"id":   1,
					"name": "123",
				},
			},
			whiteQ,
		},
	})

	q := MixQ(map[string]interface{}{
		"is_valid": true,
		"city_id":  11,
		"$and": []*Query{
			openGeoQ,
		},
	})

	s := q.JSON()
	fmt.Println(s)
}

func SimpleMongoNotQ() {
	objID := primitive.ObjectID{}
	fmt.Println(objID.Hex(), objID.IsZero())
}
