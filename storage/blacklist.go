package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"time"
)

type Blacklist interface {
	io.Closer
	GetPage(ctx context.Context, limit uint32, cursor string) (p []model.BlacklistEntry, err error)
}

type blacklistMongoEntry struct {
	Prefix    string    `bson:"prefix"`
	CreatedAt time.Time `bson:"created"`
	Reason    string    `bson:"reason"`
}

const attrPrefix = "prefix"
const attrCreated = "created"
const attrReason = "reason"

type blacklistMongo struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

var optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
var projPage = bson.D{
	{
		Key:   attrPrefix,
		Value: 1,
	},
	{
		Key:   attrCreated,
		Value: 1,
	},
	{
		Key:   attrReason,
		Value: 1,
	},
}

func NewBlacklist(ctx context.Context, cfgDb config.DbConfig) (s Blacklist, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
	if cfgDb.Tls.Enabled {
		clientOpts = clientOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cfgDb.Tls.Insecure})
	}
	if len(cfgDb.UserName) > 0 {
		auth := options.Credential{
			Username:    cfgDb.UserName,
			Password:    cfgDb.Password,
			PasswordSet: len(cfgDb.Password) > 0,
		}
		clientOpts = clientOpts.SetAuth(auth)
	}
	conn, err := mongo.Connect(ctx, clientOpts)
	var sm blacklistMongo
	if err == nil {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Blacklist.Name)
		sm.conn = conn
		sm.db = db
		sm.coll = coll
		_, err = sm.ensureIndices(ctx)
	}
	if err == nil {
		s = sm
	}
	return
}

func (sm blacklistMongo) ensureIndices(ctx context.Context) ([]string, error) {
	return sm.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   attrPrefix,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
	})
}

func (sm blacklistMongo) Close() error {
	return sm.conn.Disconnect(context.TODO())
}

func (sm blacklistMongo) GetPage(ctx context.Context, limit uint32, cursor string) (p []model.BlacklistEntry, err error) {
	q := bson.M{
		attrPrefix: bson.M{
			"$gt": cursor,
		},
	}
	optsList := options.
		Find().
		SetLimit(int64(limit)).
		SetShowRecordID(false).
		SetProjection(projPage)
	var cur *mongo.Cursor
	cur, err = sm.coll.Find(ctx, q, optsList)
	if err == nil {
		for cur.Next(ctx) {
			var e blacklistMongoEntry
			err = errors.Join(err, cur.Decode(&e))
			if err == nil {
				p = append(p, model.BlacklistEntry{
					Prefix: e.Prefix,
					Value: model.BlacklistValue{
						CreatedAt: e.CreatedAt,
						Reason:    e.Reason,
					},
				})
			}
		}
	}
	return
}
