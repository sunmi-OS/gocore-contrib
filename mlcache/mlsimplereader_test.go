package mlcache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

func newRedisCli() *redis.Client {
	ops := &redis.Options{
		Addr:     "",
		Password: "",
		DB:       0,
	}
	rc := redis.NewClient(ops)
	if err := rc.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return rc
}

func TestSimpleReader(t *testing.T) {
	cli := newRedisCli()
	infoReader := NewSimpleReader(cli, func(ctx context.Context, key string) (val interface{}, err error) {
		if cast.ToString(key) == "TestNotFound" {
			return nil, NotFound
		}
		info := &Info{
			Name: "test_" + cast.ToString(key),
			Age:  18,
		}
		return info, nil
	}, SimpleOpt{
		CacheKeyPrefix: "prefix",
		Opt:            Opt{TTL: time.Minute},
		NotFoundFunc:   IsNotFound,
	})

	ctx := context.Background()
	data := &Info{}
	cs, err := infoReader.Get(ctx, "t1", data)
	if err != nil {
		t.Errorf("err:%v", err)
	} else {
		// test_t1 18
		t.Logf("data:%v,cs:%v", data, cs) // Found=true, CacheFlag=L3
	}

	cs, err = infoReader.Get(ctx, "t1", data)
	if err != nil {
		t.Errorf("err:%v", err)
	} else {
		// test_t1 18
		t.Logf("data:%v,cs:%v", data, cs) // Found=true, CacheFlag=L2
	}

	cs, err = infoReader.Get(ctx, "TestNotFound", data)
	if err != nil {
		// will get mlcache.ErrNotFound
		t.Logf("err: %v", err) // err: error: code = -1 msg = not found metadata = map[] cause = <nil>
	} else {
		t.Errorf("should be not found !")
	}

}
