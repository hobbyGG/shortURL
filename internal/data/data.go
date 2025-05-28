package data

import (
	"context"
	"shortURL/internal/conf"
	"shortURL/internal/data/query"
	"shortURL/internal/data/sequence"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewShortURLRepo, NewDB, NewRedisCli, sequence.NewSeqUseCase)

// Data .
type Data struct {
	mdb *query.Query
	rdb *redis.Client
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	query.SetDefault(db)
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{mdb: query.Q, rdb: rdb}, cleanup, nil
}
func NewDB(conf *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(conf.Database.Source))
	if err != nil {
		panic(err)
	}
	return db
}
func NewRedisCli(conf *conf.Data) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: conf.Redis.Addr,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	return rdb
}
