package storage

import (
	"context"
	"fmt"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"os"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewBlacklist(t *testing.T) {
	//
	collName := fmt.Sprintf("blacklist-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "pub",
	}
	dbCfg.Table.Blacklist.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewBlacklist(ctx, dbCfg)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	//
	clear(ctx, t, s.(blacklistMongo))
}

func clear(ctx context.Context, t *testing.T, s blacklistMongo) {
	require.Nil(t, s.coll.Drop(ctx))
	require.Nil(t, s.Close())
}

func TestBlacklist_GetPage(t *testing.T) {
	//
	collName := fmt.Sprintf("blacklist-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "pub",
	}
	dbCfg.Table.Blacklist.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewBlacklist(ctx, dbCfg)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	//
	defer clear(ctx, t, s.(blacklistMongo))

	_, err = s.(blacklistMongo).coll.InsertMany(ctx, []any{
		bson.M{
			attrPrefix:  "foo",
			attrCreated: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			attrReason:  "reason 1",
		},
		bson.M{
			attrPrefix:  "yohoho",
			attrCreated: time.Date(2025, 12, 14, 20, 18, 50, 0, time.UTC),
			attrReason:  "reason 2",
		},
	})
	require.Nil(t, err)

	cases := map[string]struct {
		limit  uint32
		cursor string
		out    []model.BlacklistEntry
		err    error
	}{
		"default": {
			limit: 10,
			out: []model.BlacklistEntry{
				{
					Prefix: "foo",
					Value: model.BlacklistValue{
						CreatedAt: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
						Reason:    "reason 1",
					},
				},
				{
					Prefix: "yohoho",
					Value: model.BlacklistValue{
						CreatedAt: time.Date(2025, 12, 14, 20, 18, 50, 0, time.UTC),
						Reason:    "reason 2",
					},
				},
			},
		},
		"limit": {
			limit: 1,
			out: []model.BlacklistEntry{
				{
					Prefix: "foo",
					Value: model.BlacklistValue{
						CreatedAt: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
						Reason:    "reason 1",
					},
				},
			},
		},
		"cursor": {
			cursor: "foo",
			out: []model.BlacklistEntry{
				{
					Prefix: "yohoho",
					Value: model.BlacklistValue{
						CreatedAt: time.Date(2025, 12, 14, 20, 18, 50, 0, time.UTC),
						Reason:    "reason 2",
					},
				},
			},
		},
	}

	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var p []model.BlacklistEntry
			p, err = s.GetPage(ctx, c.limit, c.cursor)
			assert.Equal(t, c.out, p)
			assert.ErrorIs(t, err, c.err)
		})
	}
}
