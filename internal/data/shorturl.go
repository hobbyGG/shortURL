package data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"shortURL/internal/biz"
	"shortURL/internal/data/model"

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

func (r *ShortURLRepo) GetLongURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	table := r.data.mdb.ShortURLMap
	slMap, err := table.WithContext(ctx).Where(table.Surl.Eq(shortURL)).First()
	if err != nil {
		// err类型判断交给biz层
		return "", err
	}
	return slMap.Lurl, err
}
func (r *ShortURLRepo) GetShortURLByLongURL(ctx context.Context, longURL string) (string, error) {
	table := r.data.mdb.ShortURLMap
	slMap, err := table.WithContext(ctx).Where(table.Lurl.Eq(longURL)).First()
	if err != nil {
		return "", err
	}
	return slMap.Surl, nil
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
