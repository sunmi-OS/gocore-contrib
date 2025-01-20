package mlcache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto"
	ristrettoStore "github.com/eko/gocache/store/ristretto/v4"
	"github.com/spf13/cast"
	"github.com/sunmi-OS/gocore/v2/api/ecode"
)

const NotFoundCode = 404

var NotFound = ecode.NewV2(NotFoundCode, "not found")

type Info struct {
	Name string
	Age  int32
}

func IsNotFound(err error) bool {
	if se := new(ecode.ErrorV2); errors.As(err, &se) {
		return se.Code() == NotFoundCode
	}
	return false
}

func TestMlmemorydb(t *testing.T) {
	smallRCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 2e5,       // 需要是持久化数量的10倍，这里设置为20w
		MaxCost:     100 << 20, // 存储字节数
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	infoSmallRStore := ristrettoStore.NewRistretto(smallRCache)
	infoReader := NewMemoryDBReader(infoSmallRStore, func(ctx context.Context, key string) (val interface{}, err error) {
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
