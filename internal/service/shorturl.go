package service

import (
	"context"
	"errors"
	"strings"

	pb "shortURL/api/shorturl/v1"
	"shortURL/internal/biz"
	"shortURL/third_party/httpx"
)

type ShortURLService struct {
	pb.UnimplementedShortURLServer

	uc *biz.ShortURLUsecase
}

func NewShortURLService(uc *biz.ShortURLUsecase) *ShortURLService {
	return &ShortURLService{uc: uc}
}

func (s *ShortURLService) Convert(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	// 参数提取
	longURL := req.LongURL
	// 检查url是否有效
	ok, err := httpx.Ping(longURL)
	if !ok {
		return &pb.ConvertResponse{}, err
	}

	// 调用convert业务逻辑
	shortURL, err := s.uc.Convert(ctx, longURL)
	if err != nil {
		return &pb.ConvertResponse{}, err
	}

	return &pb.ConvertResponse{ShortURL: shortURL}, nil
}
func (s *ShortURLService) Redirect(ctx context.Context, req *pb.RedirectRequest) (*pb.RedirectResponse, error) {
	// 参数处理
	shortURL := req.ShortURL

	// 忽略最后为/的情况
	// if HasSuffix(s, suffix) {
	// return s[:len(s)-len(suffix)]
	// }
	shortURL = strings.TrimSuffix(shortURL, "/")

	urlSlice := strings.Split(shortURL, "/")
	l := len(urlSlice)
	if l <= 0 {
		return nil, errors.New("invalid shortURL")
	}
	surl := urlSlice[l-1]

	longURL, err := s.uc.Redirect(ctx, surl)
	if err != nil {
		return &pb.RedirectResponse{}, err
	}

	return &pb.RedirectResponse{LongURL: longURL}, nil
}
