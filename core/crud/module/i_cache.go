package module

import (
	"context"
	"time"
)

type ICache interface {
	IModule
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get 获取缓存
	Get(ctx context.Context, key string, value any) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// MSet 批量设置
	MSet(ctx context.Context, values map[string]interface{}, expiration time.Duration) error

	// MGet 批量获取
	MGet(ctx context.Context, keys []string) (map[string]string, error)

	// MDelete 批量删除
	MDelete(ctx context.Context, keys []string) error

	// LPush 左推入列表
	LPush(ctx context.Context, key string, values ...interface{}) error

	// RPush 右推入列表
	RPush(ctx context.Context, key string, values ...interface{}) error

	// LPop 左弹出列表
	LPop(ctx context.Context, key string) (string, error)

	// RPop 右弹出列表
	RPop(ctx context.Context, key string) (string, error)

	// LRange 获取列表范围
	LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// SAdd 添加集合成员
	SAdd(ctx context.Context, key string, members ...interface{}) error

	// SMembers 获取集合所有成员
	SMembers(ctx context.Context, key string) ([]string, error)

	// SRem 删除集合成员
	SRem(ctx context.Context, key string, members ...interface{}) error

	// SIsMember 判断是否是集合成员
	SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// ZAdd 添加有序集合成员
	ZAdd(ctx context.Context, key string, score float64, member interface{}) error

	// ZRange 获取有序集合范围
	ZRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// ZRangeByScore 按分数获取有序集合范围
	ZRangeByScore(ctx context.Context, key string, min, max float64) ([]string, error)

	// HSet 设置哈希字段
	HSet(ctx context.Context, key, field string, value interface{}) error

	// HGet 获取哈希字段
	HGet(ctx context.Context, key, field string) (string, error)
}
