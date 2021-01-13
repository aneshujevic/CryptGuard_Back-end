package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"sync"
	"time"
)

const dbURI = "mongodb://localhost:27017/CryptGuardDB"
var lock = &sync.Mutex{}

type Database struct {
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

var singleInstance *Database

func (db *Database) GetInstance() *Database {
	defer lock.Unlock()

	if db.client == nil {
		lock.Lock()
		db.createInstance()
	}
	return singleInstance
}

func (db *Database) createInstance() {
	db.ctx, db.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(db.ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		panic(err)
	}

	// Ping the primary
	if err := client.Ping(db.ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")
	db.client = client
}

func (db *Database) destroyConnection() {
	if err := db.client.Disconnect(db.ctx); err != nil {
		panic(err)
	}
}
