package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"shortURL/internal/biz"
	"shortURL/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type ShortURLRepo struct {
	data *Data
	log  *log.Helper
}

func NewShortURLRepo(data *Data, logger log.Logger) biz.ShortURLRepo {
	// 指定gen的db
	return &ShortURLRepo{data: data, log: log.NewHelper(logger)}
}

// GetLongURLByShortURL
func (r *ShortURLRepo) GetLongURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	table := r.data.mdb.ShortURLMap
	slMap, err := table.WithContext(ctx).Where(table.Surl.Eq(shortURL)).First()
	if err != nil {
		// err类型判断交给biz层
		r.log.Debugw("[mysql] GetLongURLByShortURL shortURL:", shortURL)
		return "", err
	}
	return slMap.Lurl, err
}

// GetShortURLByLongURL
func (r *ShortURLRepo) GetShortURLByLongURL(ctx context.Context, longURL string) (string, error) {
	table := r.data.mdb.ShortURLMap
	slMap, err := table.WithContext(ctx).Where(table.Lurl.Eq(longURL)).First()
	if err != nil {
		return "", err
	}
	return slMap.Surl, nil
}
func (r *ShortURLRepo) GetShortURLs(ctx context.Context) ([]string, error) {
	result, err := r.data.mdb.ShortURLMap.WithContext(ctx).
		Select(r.data.mdb.ShortURLMap.Surl).
		Find()
	if err != nil {
		return nil, err
	}
	ss := make([]string, 0, len(result))
	for _, r := range result {
		ss = append(ss, r.Surl)
	}
	return ss, nil
}
func (r *ShortURLRepo) CreateSLMap(ctx context.Context, shortURL, longURL string) error {
	table := r.data.mdb.ShortURLMap
	lurlByte := []byte(longURL)
	hash := sha256.New()
	if _, err := hash.Write(lurlByte); err != nil {
		return err
	}
	hashSum := hash.Sum(nil)
	lurlSHA := hex.EncodeToString(hashSum)
	slMap := &model.ShortURLMap{
		Lurl:    longURL,
		LurlMd5: lurlSHA, //不使用md5加密，使用sha256加密
		Surl:    shortURL,
	}
	return table.WithContext(ctx).Create(slMap)
}

// redis功能
func (r *ShortURLRepo) RediGet(ctx context.Context, key string) (string, error) {
	return r.data.rdb.Get(ctx, key).Result()
}
func (r *ShortURLRepo) RediSet(ctx context.Context, key, val string, expTime ...time.Duration) error {
	if len(expTime) == 0 {
		// 没有传参，默认1小时
		return r.data.rdb.Set(ctx, key, val, time.Hour).Err()
	}
	return r.data.rdb.Set(ctx, key, val, expTime[0]).Err()
}
