package biz

import (
	"context"
	"errors"
	"shortURL/internal/conf"
	"shortURL/internal/data/param"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/bloom"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ShortURLRepo interface {
	GetLongURLByShortURL(ctx context.Context, shortURL string) (string, error)
	GetShortURLByLongURL(ctx context.Context, longURL string) (string, error)
	GetShortURLs(ctx context.Context) ([]string, error)
	CreateSLMap(ctx context.Context, shortURL, longURL string) error

	RediGet(ctx context.Context, key string) (string, error)
	RediSet(ctx context.Context, key string, val string, expTime ...time.Duration) error
}
type SequenceUseCase interface {
	Get(context.Context) (int64, error)
}

type ShortURLUsecase struct {
	repo         ShortURLRepo
	seq          SequenceUseCase
	log          *log.Helper
	validCharMap map[rune]struct{}
	domain       string
}

func NewShortURLUsecase(conf *conf.Biz, repo ShortURLRepo, seq SequenceUseCase, bf *bloom.Filter, logger log.Logger) *ShortURLUsecase {
	m := make(map[rune]struct{}, 62)
	baseString := "VJ7y3fWdPZ9tSqEa8uN4XcGQnH2LxK6w15iFbO0rDkYgBmTzIeMhRvUoJlC"
	for _, v := range baseString {
		m[v] = struct{}{}
	}

	// 填充bloom过滤器
	urls, err := repo.GetShortURLs(context.Background())
	if err != nil {
		panic(err)
	}
	for _, v := range urls {
		bf.Add([]byte(v))
	}
	return &ShortURLUsecase{repo: repo, seq: seq, validCharMap: m, domain: conf.Domain, log: log.NewHelper(logger)}
}

// Convert 接收一个有效长连接，将长连接转为短连接并存入mysql数据库，返回短连接和错误。
func (uc *ShortURLUsecase) Convert(ctx context.Context, longURL string) (string, error) {
	// 使用数据库参数检查，传入的不能是短链接，不能是已经存在的长链接
	urlChar := strings.Split(longURL, "/")[len(strings.Split(longURL, "/"))-1]
	_, err := uc.repo.GetLongURLByShortURL(ctx, urlChar)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err != nil {
			return "", err
		}
		return "", errors.New("url is shortURL")
	}

	_, err = uc.repo.GetShortURLByLongURL(ctx, longURL)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if err != nil {
			return "", err
		}
		return "", errors.New("longURL existed")
	}
	err = nil

	// redis取号
	// 生成短连接
	// 将结果存入mysql数据库
	seq, err := uc.seq.Get(ctx)
	if err != nil {
		return "", err
	}
	seqTemp := seq
	baseStr := "VJ7y3fWdPZ9tSqEa8uN4XcGQnH2LxK6w15iFbO0rDkYgBmTzIeMhRvUoJlC"
	sURLChar := make([]byte, 0, 8)
	for seqTemp > 0 {
		i := seqTemp % 62
		seqTemp /= 62
		sURLChar = append(sURLChar, baseStr[i])
	}

	if err := uc.repo.CreateSLMap(ctx, string(sURLChar), longURL); err != nil {
		return "", err
	}

	// 将短连接拼接后返回
	url := strings.Builder{}
	url.WriteString(uc.domain)
	if uc.domain[len(uc.domain)-1] != '/' {
		url.WriteString("/")
	}
	for _, v := range sURLChar {
		url.WriteByte(v)
	}

	return url.String(), nil
}

func (uc *ShortURLUsecase) Redirect(ctx context.Context, shortURL string) (string, error) {
	// 参数检查
	// 不能具有非法字符,同时可以过滤存在querystring的情况，目前短连接不允许qs
	for _, v := range shortURL {
		_, ok := uc.validCharMap[v]
		if !ok {
			// 不在合法字符集内
			return "", errors.New("ivnalide char in url")
		}
	}

	// 先查看缓存是否有短链接对应的长链接记录
	key := strings.Builder{}
	key.WriteString(param.KeyPreffix)
	key.WriteString(shortURL)
	sfg := singleflight.Group{}

	// Do直接返回结果，DoChan会返回一个chan，支持异步调用，防止一个请求导致所有请求堵塞
	longURL, err, shared := sfg.Do("redirect:"+shortURL, func() (interface{}, error) {

		longURL, err := uc.repo.RediGet(ctx, key.String())
		if err == nil {
			uc.log.Infow("[biz] redis", "hit")
			return longURL, nil
		}
		if err == redis.Nil {
			uc.log.Infow("[biz] redis", "miss")
			// 缓存未命中，需要去数据库中查询
			longURL, err := uc.repo.GetLongURLByShortURL(ctx, shortURL)
			if err != nil {
				uc.log.Debugw("[biz] Redirect shortURL:", shortURL)
				return "", err
			}
			if err := uc.repo.RediSet(ctx, key.String(), longURL, time.Minute*30); err != nil {
				return "", err
			}
			return longURL, nil
		}
		return "", err
	})
	if err != nil {
		uc.log.Debugw("[biz] Redirect singleflight:", err)
		return "", err
	}
	uc.log.Infow("[biz] shared", shared)
	return longURL.(string), nil
}
