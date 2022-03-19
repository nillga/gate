package cache

import "github.com/nillga/jwt-server/entity"

type GatewayCache interface {
	Clear(key string)
	Get(key string) (*entity.User, bool)
	Put(key string, value *entity.User)
}

type cache map[string]*entity.User

func NewCache() GatewayCache {
	return &cache{}
}

func (c *cache) Clear(key string) {
	(*c)[key] = nil
}

func (c *cache) Get(key string) (*entity.User, bool) {
	val, ok := (*c)[key]
	return val, ok
}

func (c *cache) Put(key string, value *entity.User) {
	(*c)[key] = value
}
