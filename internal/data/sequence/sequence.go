package sequence

import (
	"context"
	"shortURL/internal/biz"
	"shortURL/internal/data/param"

	"github.com/redis/go-redis/v9"
)

// 存放取号相关逻辑

type SeqUseCase struct {
	rdb *redis.Client
}

func (uc *SeqUseCase) Get(ctx context.Context) (int64, error) {
	// 从redis中取号
	seq, err := uc.rdb.Incr(ctx, param.KeySequence).Result()
	if err != nil {
		return -1, err
	}
	return seq, err
}

func NewSeqUseCase(rdb *redis.Client) biz.SequenceUseCase {
	return &SeqUseCase{rdb: rdb}
}
