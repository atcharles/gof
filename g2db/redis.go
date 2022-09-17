package g2db

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cast"

	"github.com/atcharles/gof/v2/g2util"
	"github.com/atcharles/gof/v2/json"
)

// constants defined
const (
	redisSubChannel     = "Sub"
	redisSubDelMemCache = "DelMemCache"
	redisSubDelMemAll   = "DelMemAll"
)

type (
	//RedisSubHandlerFunc ...
	RedisSubHandlerFunc func(payload []byte)

	redisObj struct {
		Logger g2util.LevelLogger `inject:""`
		Config *g2util.Config     `inject:""`
		//Go     *g2util.GoPool     `inject:""`
		Cache *cacheMem `inject:""`

		mu sync.RWMutex
		mp sync.Map

		closeSub chan struct{}
		isClose  bool

		subHandlers map[string]RedisSubHandlerFunc
	}
	//redisSubPayload ...
	redisSubPayload struct {
		Name string           `json:"name,omitempty"`
		Data *json.RawMessage `json:"data,omitempty"`
	}
)

func (r *redisObj) AfterShutdown() {
	r.closeRedisSub()
}

// close ......
func (r *redisObj) closeRedisSub() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isClose {
		return
	}
	r.isClose = true
	close(r.closeSub)
}

// PubDelCache ...
func (r *redisObj) PubDelCache(keys []string) error { return r.Pub(r.pubDelMemName(), keys) }

// PubDelMemAll ...
func (r *redisObj) PubDelMemAll() error { return r.Pub(r.formatWithAppName(redisSubDelMemAll), nil) }

// Pub ...
func (r *redisObj) Pub(name string, data interface{}) error {
	bd, err := json.Marshal(data)
	if err != nil {
		return err
	}
	rawText := json.RawMessage(bd)
	payload := &redisSubPayload{Name: name, Data: &rawText}
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.client().Publish(context.Background(), r.subChannel(), msg).Err()
}

func (r *redisObj) SubHandle(name string, handler RedisSubHandlerFunc) { r.subHandlers[name] = handler }

// Subscribe ...
func (r *redisObj) Subscribe() {
	r.subHandlers = make(map[string]RedisSubHandlerFunc)
	r.closeSub = make(chan struct{})
	//rev handlers
	r.SubHandle(r.pubDelMemName(), r.Cache.RedisSubDelCache())
	r.SubHandle(r.formatWithAppName(redisSubDelMemAll), r.Cache.RedisSubDelMemAll())
	if e := r.subscribe(); e != nil {
		r.Logger.Fatalln(e)
	}
}

// subAction ...
func (r *redisObj) subAction(sub *redis.PubSub) {
	for {
		select {
		case <-r.closeSub:
			_ = sub.Close()
			r.Logger.Debugf("[SUB] Redis关闭订阅")
			r.mp.Range(func(_, value interface{}) bool { _ = value.(*redis.Client).Close(); return true })
			return
		case msg, ok := <-sub.Channel():
			if !ok || msg == nil {
				r.Logger.Warnf("[SUB] 接收订阅消息失败")
				return
			}
			//接收到订阅的消息,执行数据解析,与删除
			r.Logger.Debugf("[SUB] 接收到订阅消息: %s\n", msg.Payload)
			payload := new(redisSubPayload)
			if e := json.Unmarshal([]byte(msg.Payload), payload); e != nil {
				r.Logger.Errorf("[SUB] 无效的订阅内容: %s\n", msg.Payload)
				continue
			}
			//run handlerFunc
			payloadData := payload.Data
			if payloadData == nil {
				_d1 := make(json.RawMessage, 0)
				payloadData = &_d1
			}
			if handler, _ok := r.subHandlers[payload.Name]; _ok {
				handler(*payloadData)
			}
		}
	}
}

func (r *redisObj) subscribe() (err error) {
	ctx := context.Background()
	channelName := r.subChannel()
	sub := r.client().Subscribe(ctx, channelName)
	_, err = sub.ReceiveTimeout(ctx, time.Second*3)
	if err != nil {
		err = fmt.Errorf("订阅Redis失败: %s", err.Error())
		return
	}
	r.Logger.Debugf("[SUB] Redis订阅成功: %s", channelName)
	go r.subAction(sub)
	return
}

// Client ...
func (r *redisObj) Client(db ...int) *redis.Client { return r.client(db...) }
func (r *redisObj) client(dbs ...int) *redis.Client {
	db := 0
	if len(dbs) > 0 {
		db = dbs[0]
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.mp.Load(db); ok {
		return v.(*redis.Client)
	}
	cl := redis.NewClient(r.newOption(db))
	if e := cl.Ping(context.Background()).Err(); e != nil {
		panic(e)
	}
	r.mp.Store(db, cl)
	return cl
}

// newOption ...
func (r *redisObj) newOption(db int) *redis.Options {
	cfg := r.Config.Viper().GetStringMapString("redis")
	svAddr := r.Config.Viper().GetString("global.host")
	return &redis.Options{
		Addr:         strings.Replace(cfg["host"], "{host}", svAddr, -1),
		Password:     cfg["pwd"],
		DB:           db,
		MaxRetries:   cast.ToInt(cfg["max_retries"]),
		MinIdleConns: cast.ToInt(cfg["min_idle_connections"]),
		MaxConnAge:   cast.ToDuration(cfg["max_conn_age_seconds"]) * time.Second,
	}
}

// subChannel ...
func (r *redisObj) subChannel() string {
	return fmt.Sprintf("%s_%s", r.Config.Viper().GetString("name"), redisSubChannel)
}

// pubDelMemName ...
func (r *redisObj) pubDelMemName() string {
	return fmt.Sprintf("%s_%s", r.Config.Viper().GetString("name"), redisSubDelMemCache)
}

// formatWithAppName ...
func (r *redisObj) formatWithAppName(s string) string {
	return fmt.Sprintf("%s_%s", r.Config.Viper().GetString("name"), s)
}
