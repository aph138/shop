package myredis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const DAY = time.Hour * 24

type myRedis struct {
	client  *redis.Client
	options *opt
}
type opt struct {
	timeOut time.Duration
}
type optFunc func(*opt)

func defaultOption() *opt {
	return &opt{
		timeOut: time.Millisecond * 200,
	}
}

func WithTimeOut(t time.Duration) optFunc {
	return func(o *opt) {
		o.timeOut = t
	}
}
func NewRedis(opt *redis.Options, opts ...optFunc) (*myRedis, error) {
	option := defaultOption()
	for _, o := range opts {
		o(option)
	}
	rdb := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5000)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &myRedis{
		client:  rdb,
		options: option,
	}, nil
}

func (r *myRedis) Set(key string, value any, expiration ...time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.options.timeOut)
	defer cancel()
	v, err := json.Marshal(value)
	if err != nil {
		return err
	}
	exp := DAY
	if len(expiration) > 0 {
		exp = expiration[0]
	}
	if err = r.client.Set(ctx, key, v, exp).Err(); err != nil {
		return err
	}
	return nil
}
func (r *myRedis) Get(key string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.options.timeOut)
	defer cancel()
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(result), value); err != nil {
		return err
	}
	return nil
}
