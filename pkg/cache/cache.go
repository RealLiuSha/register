package cache

import (
	"errors"
	"git.wolaidai.com/DevOps/register/pkg/utils/log"
	"sync"
)

type cache struct {
	mutex sync.Mutex
	envs  map[string]map[string]interface{}
}

func (c cache) GetEnv(id string) (map[string]interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if env, ok := c.envs[id]; ok {
		return env, nil
	}

	return nil, errors.New("env not found in cache")
}

func (c cache) AddEnv(id string, env map[string]interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if oldEnv, ok := c.envs[id]; ok {
		log.Info("Overwrite existing cache: ", id)
		log.Debug("old cache:", oldEnv)
		log.Debug("new cache:", env)
		c.envs[id] = env
	}

	log.Debug("new cache:", id, "-->", env)
	c.envs[id] = env
}

func (c cache) DelEnv(id string) {
	c.mutex.Lock()

	delete(c.envs, id)
	c.mutex.Unlock()

	log.Debug("Delete Cache:", id)
}

var Cache = cache{mutex: sync.Mutex{}, envs: make(map[string]map[string]interface{})}
