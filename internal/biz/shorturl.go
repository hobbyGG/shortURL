package biz

import (
	"context"
	"errors"
	"shortURL/internal/conf"
	"strings"

	"gorm.io/gorm"
)

type ShortURLRepo interface {
	GetLongURLByShortURL(ctx context.Context, shortURL string) (string, error)
	GetShortURLByLongURL(ctx context.Context, longURL string) (string, error)
	CreateSLMap(ctx context.Context, shortURL, longURL string) error
}
type SequenceUseCase interface {
	Get(context.Context) (int64, error)
}

type ShortURLUsecase struct {
	repo         ShortURLRepo
	seq          SequenceUseCase
	validCharMap map[rune]struct{}
	domain       string
}

func NewShortURLUsecase(conf *conf.Biz, repo ShortURLRepo, seq SequenceUseCase) *ShortURLUsecase {
	m := make(map[rune]struct{}, 62)
	baseString := "VJ7y3fWdPZ9tSqEa8uN4XcGQnH2LxK6w15iFbO0rDkYgBmTzIeMhRvUoJlC"
	for _, v := range baseString {
		m[v] = struct{}{}
	}
	return &ShortURLUsecase{repo: repo, seq: seq, validCharMap: m, domain: conf.Domain}
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

	// 根据短连接查长连接
	return uc.repo.GetLongURLByShortURL(ctx, shortURL)
}
