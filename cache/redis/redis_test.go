package redis_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wuzhengjerry/udns-api/cache"
	"github.com/wuzhengjerry/udns-api/cache/redis"
)

// adapterSuit 适配结构体
type adapterSuit struct {
	adapter cache.Cache
	testKey string
	testVal string
}

// SetUp 初始化
func (a *adapterSuit) SetUp() {
	adapter := redis.NewCache(redis.NewDefaultConfig())
	a.adapter = adapter
	a.testKey = "testkey01"
	a.testVal = "testval01"
}

// TearDown 适配关闭
func (a *adapterSuit) TearDown() {
	a.adapter.Close()
}

// TestRedisAdapterSuit 测试
func TestRedisAdapterSuit(t *testing.T) {
	suit := new(adapterSuit)
	suit.SetUp()
	defer suit.TearDown()

	t.Run("PutOK", testPutOK(suit))
	t.Run("ExistOK", testExistOK(suit))
	t.Run("ExistNotOK", testExistNotOK(suit))
	t.Run("GetOK", testGetOK(suit))
	t.Run("GetFailed", testGetFailed(suit))
	t.Run("DelOK", testDelOK(suit))
}

// testPutOK 测试
func testPutOK(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		err := a.adapter.Put(a.testKey, a.testVal)
		should.NoError(err)
	}
}

// testGetOK 测试
func testGetOK(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		val := ""
		a.adapter.Get(a.testKey, &val)
		should.Equal(a.testVal, val)
	}
}

// testGetFailed 测试
func testGetFailed(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		val := ""
		err := a.adapter.Get("xxx", &val)
		should.Equal("redis: nil", err.Error())
	}
}

// testExistOK 测试
func testExistOK(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		ok := a.adapter.IsExist(a.testKey)
		should.Equal(true, ok)
	}
}

// testExistNotOK 测试
func testExistNotOK(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		ok := a.adapter.IsExist("not exist key")
		should.Equal(false, ok)
	}
}

// testDelOK 测试
func testDelOK(a *adapterSuit) func(t *testing.T) {
	return func(t *testing.T) {
		should := require.New(t)

		err := a.adapter.Delete(a.testKey)
		should.NoError(err)
	}
}
