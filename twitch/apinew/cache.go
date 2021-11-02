package apinew

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func cacheDBDir() string {
	cachedir, _ := os.UserCacheDir()
	return filepath.Join(cachedir, "twitchbrowser")
}

func cacheFileName(id string) string {
	return filepath.Join(cacheDBDir(), id)
}

type cache struct{}

func NewCache() (*cache, error) {
	return nil, os.MkdirAll(cacheDBDir(), 0700)
}

func (*cache) GetGameNames(ids []string) (map[string]string, []string, error) {
	var todo []string
	ret := make(map[string]string, len(ids))

	for _, id := range ids {
		if name, err := ioutil.ReadFile(cacheFileName(id)); err != nil {
			todo = append(todo, id)
		} else {
			ret[id] = string(name)
		}
	}
	return ret, todo, nil
}

func (*cache) SetGameName(id, name string) error {
	return ioutil.WriteFile(cacheFileName(id), []byte(name), 0600)
}
