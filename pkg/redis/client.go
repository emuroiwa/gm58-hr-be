package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(addr, password string, db int) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &Client{rdb: rdb}
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	ctx := context.Background()

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, key, jsonData, expiration).Err()
}

func (c *Client) Get(key string, dest interface{}) error {
	ctx := context.Background()

	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Del(key string) error {
	ctx := context.Background()
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) Exists(key string) (bool, error) {
	ctx := context.Background()
	result := c.rdb.Exists(ctx, key)
	return result.Val() > 0, result.Err()
}

func (c *Client) PushJob(queue string, job interface{}) error {
	ctx := context.Background()

	jsonData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return c.rdb.LPush(ctx, queue, jsonData).Err()
}

func (c *Client) PopJob(queue string, dest interface{}) error {
	ctx := context.Background()

	val, err := c.rdb.BRPop(ctx, time.Second*10, queue).Result()
	if err != nil {
		return err
	}

	if len(val) < 2 {
		return redis.Nil
	}

	return json.Unmarshal([]byte(val[1]), dest)
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
