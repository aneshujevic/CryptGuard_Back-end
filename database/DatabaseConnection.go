package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sync"
	"time"
)

const dbURI = "mongodb://localhost:27017/CryptGuardDB"
const Name = "CryptGuardDB"
var lock = &sync.Mutex{}

type Database struct {
	Client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

var singleInstance *Database

func GetInstance() *Database {
	defer lock.Unlock()
	if singleInstance == nil {
		singleInstance = &Database{}
		if singleInstance.Client == nil {
			lock.Lock()
			singleInstance.createInstance()
		}
	}
	return singleInstance
}

func DestroyInstance() {
	singleInstance.disconnectClient()
	singleInstance.Client = nil
	singleInstance.ctx = nil
	singleInstance.cancel = nil
	singleInstance = nil
}

func (db *Database) createInstance() {
	db.ctx, db.cancel = context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(db.ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatal(err)
	}

	db.Client = client
}

func (db *Database) disconnectClient() {
	if err := db.Client.Disconnect(db.ctx); err != nil {
		log.Fatal(err)
	}
}
