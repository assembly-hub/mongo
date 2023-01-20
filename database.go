// package mongo
package mongo

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	*Client

	dbName string
	db     *mongo.Database
}

func (db *Database) Collection(name string) *Collection {
	c := new(Collection)
	c.Database = db
	c.collectionName = name
	c.collection = db.db.Collection(name)
	return c
}

func (db *Database) TryCollection(name string) (c *Collection, exist bool, err error) {
	names, err := db.db.ListCollectionNames(c.ctx, map[string]string{"name": name})
	if err != nil {
		return nil, false, err
	}

	exist = false
	if len(names) > 0 {
		exist = true
	}

	c = new(Collection)
	c.Database = db
	c.collectionName = name
	c.collection = db.db.Collection(name)
	return c, exist, nil
}
