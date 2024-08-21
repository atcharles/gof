package g2db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andeya/goutil"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"
	"github.com/unknwon/com"

	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/j2rpc"
	"github.com/atcharles/gof/v2/json"
)

// constants
const (
	SYSKey = "P6UEgd7ln9mpMz5hGWYqT21cSHOtkJQZ"

	GinContextJWTTokenKey = "JWT_RAW"

	GinContextJWTUIDKey = "JWT_UID"
)

// ItfGinContext ...
type ItfGinContext interface {
	context.Context
	Set(key string, value interface{})
	GetHeader(key string) string
	Query(key string) string
}

// Token ...
type Token struct {
	Config *g2util.Config `inject:""`

	Redis *redisObj `inject:""`
	Cache *cacheMem `inject:""`

	option *TokenOption
}

const defaultTokenTimeout = time.Hour * 24 * 10

// NewCopy ...
func (t *Token) NewCopy(key string) *Token {
	return &Token{
		Config: t.Config,
		Redis:  t.Redis,
		Cache:  t.Cache,
		option: &TokenOption{
			CacheKey:   key,
			Timeout:    defaultTokenTimeout,
			MaxRefresh: defaultTokenTimeout / 2,
			EncryptKey: []byte(SYSKey)[:16],
			MultiLogin: func(id int64) bool { return false },
		},
	}
}

// Constructor ...
func (t *Token) Constructor() {
	t.option = &TokenOption{
		CacheKey:   "h:token",
		Timeout:    defaultTokenTimeout,
		MaxRefresh: defaultTokenTimeout / 2,
		EncryptKey: []byte(SYSKey)[:16],
		MultiLogin: func(id int64) bool { return false },
	}
}

// AfterLogin ...
func (t *Token) AfterLogin(ctx context.Context, id int64) (td *TokenData, err error) {
	return t.write2redis(ctx, id)
}

// Verify ...
func (t *Token) Verify(ctx ItfGinContext, fns ...func() error) (err error) {
	return t.verify(ctx, fns...)
}

// Logout ...
func (t *Token) Logout(ctx context.Context, id int64) (err error) { return t.removeTokenData(ctx, id) }

// Option ...
func (t *Token) Option() *TokenOption { return t.option }

func (t *Token) redisCacheKey() string {
	name := t.Config.Viper().GetString("name")
	if len(name) == 0 {
		return t.option.CacheKey
	}
	return fmt.Sprintf("%s:%s", name, t.option.CacheKey)
}

// redisField ...
func (t *Token) redisField(id int64) string { return cast.ToString(id) }

// memCacheKey ...
func (t *Token) memCacheKey(id int64) string { return fmt.Sprintf("%s::%d", t.redisCacheKey(), id) }

// write2redis ...
func (t *Token) write2redis(ctx context.Context, id int64) (td *TokenData, err error) {
	if t.option.MultiLogin(id) {
		var ts1 string
		ts1, err = t.Redis.Client().HGet(ctx, t.redisCacheKey(), cast.ToString(id)).Result()
		if err == nil {
			td, err = unmarshalTokenData(ts1)
			return
		}
	}
	td = t.option.generateTd(id)
	err = td.write2redis(ctx, t)
	return
}

// verify ...
func (t *Token) verify(ctx ItfGinContext, fns ...func() error) (err error) {
	token := ctx.GetHeader("token")
	if len(token) == 0 {
		token = ctx.GetHeader("Authorization")
	}
	if len(token) == 0 {
		token = ctx.Query("token")
	}

	token = strings.Replace(token, "Bearer ", "", -1)
	if len(token) == 0 {
		return j2rpc.TokenError("非法访问")
	}

	id, err := t.option.getIDByToken(token)
	if err != nil {
		return j2rpc.TokenError(fmt.Sprintf("无效的令牌:%s", err.Error()))
	}
	tt, err := t.cacheVerify(ctx, id, token)
	if err != nil {
		return j2rpc.TokenError(fmt.Sprintf("身份认证失败:%s", err.Error()))
	}

	ctx.Set(GinContextJWTTokenKey, token)
	ctx.Set(GinContextJWTUIDKey, tt.ID)

	if len(fns) == 0 {
		return
	}

	for _, fn := range fns {
		if fn == nil {
			continue
		}
		if err = fn(); err != nil {
			return
		}
	}
	return
}

func (t *Token) cacheVerify(ctx ItfGinContext, id int64, token string) (tt *TokenData, err error) {
	redisClient := t.Redis.Client()
	redisKey, redisField := t.redisCacheKey(), t.redisField(id)
	memCacheKey := t.memCacheKey(id)
	data, err := t.Cache.GetOrStore(memCacheKey, func() ([]byte, error) {
		_s, _e := redisClient.HGet(ctx, redisKey, redisField).Result()
		if _e == redis.Nil {
			return []byte(cacheNULL), nil
		}
		if _e != nil {
			return nil, _e
		}
		return []byte(_s), nil
	})
	if err != nil {
		return
	}
	if string(data) == cacheNULL {
		return nil, errors.New("未知的令牌")
	}

	tt, err = unmarshalTokenData(string(data))
	if err != nil {
		return nil, err
	}
	if tt.Token != token {
		return nil, errors.New("无效的令牌")
	}
	now := g2util.TimeNow()
	if now.After(*tt.Expire) {
		if err = t.removeTokenData(ctx, id); err != nil {
			return
		}
		return nil, errors.New("过期的令牌")
	}

	if tt.Expire.Sub(now) < t.option.MaxRefresh {
		tt.Expire = t.option.expireTimeAddr()
		if err = tt.write2redis(ctx, t); err != nil {
			return
		}
	}
	return
}

// removeTokenData ...
func (t *Token) removeTokenData(ctx context.Context, id int64) (err error) {
	if err = t.Redis.Client().HDel(ctx, t.redisCacheKey(), t.redisField(id)).Err(); err != nil {
		return
	}
	if err = t.Redis.PubDelCache([]string{t.memCacheKey(id)}); err != nil {
		return
	}
	return
}

// TokenOption ...
type TokenOption struct {
	CacheKey string
	//有效期时间
	Timeout time.Duration
	//距离到期时间剩余n时刷新
	MaxRefresh time.Duration
	//加密秘钥
	EncryptKey []byte
	//是否允许多点登录
	MultiLogin func(id int64) bool
}

// expireTimeAddr ...
func (t *TokenOption) expireTimeAddr() *time.Time {
	to1 := g2util.TimeNow().Add(t.Timeout)
	return &to1
}

// generateTd ...
func (t *TokenOption) generateTd(id int64) (td *TokenData) {
	td = &TokenData{ID: id}
	id1 := goutil.StringToBytes(cast.ToString(id))
	td.Token = goutil.BytesToString(bytes.ToUpper(goutil.AESCBCEncrypt(t.EncryptKey, id1)))
	td.Expire = t.expireTimeAddr()
	return
}

// getIDByToken ...
func (t *TokenOption) getIDByToken(token string) (id int64, err error) {
	id1, err := goutil.AESCBCDecrypt(t.EncryptKey, bytes.ToLower([]byte(token)))
	if err != nil {
		return
	}
	id = com.StrTo(id1).MustInt64()
	return
}

func unmarshalTokenData(s string) (tt *TokenData, err error) {
	tt = new(TokenData)
	err = json.Unmarshal([]byte(s), tt)
	return
}

// TokenData ...
type TokenData struct {
	ID     int64      `json:"id,omitempty"`
	Token  string     `json:"token,omitempty"`
	Expire *time.Time `json:"expire,omitempty"`
}

func (t *TokenData) String() string { return g2util.JSONDump(t) }

// write2redis ...
func (t *TokenData) write2redis(ctx context.Context, tk *Token) (err error) {
	kv := map[string]interface{}{cast.ToString(t.ID): t.String()}
	err = tk.Redis.Client().HSet(ctx, tk.redisCacheKey(), kv).Err()
	if err != nil {
		return
	}
	if err = tk.Redis.PubDelCache([]string{tk.memCacheKey(t.ID)}); err != nil {
		return
	}
	return
}
