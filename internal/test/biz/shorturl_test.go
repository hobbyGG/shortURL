package test

import (
	"context"
	v1 "shortURL/api/shorturl/v1"
	"sync"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func TestRedirect(t *testing.T) {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("127.0.0.1:9000"),
		grpc.WithTimeout(time.Second*30),
	)
	if err != nil {
		t.Fatal(err)
	}
	client := v1.NewShortURLClient(conn)
	shorURL := "127.0.0.1:9000/3"
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for range 10000 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			if _, err := client.Redirect(
				ctx,
				&v1.RedirectRequest{ShortURL: shorURL},
			); err != nil {
				wg.Done()
				t.Error(err)
				return
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// 未增加singleflight时的耗时为0.65s
	// 增加后时间为
}
