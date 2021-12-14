package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/chyroc/beauty/src/image_adapter"
)

func main() {
	configs, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	for _, config := range configs {
		cli := config.NewAdapter()
		for _, user := range config.Users {
			images, err := cli.GetImage(user.Id)
			if err != nil {
				panic(err)
			}
			for _, v := range images {
				err = config.Save(user.Id, v)
				if err != nil {
					panic(err)
				}
			}
			if err = config.Index(user.Id); err != nil {
				panic(err)
			}
		}
	}

	if err = configs.Index(); err != nil {
		panic(err)
	}
}

// config
func LoadConfig() (FetchConfigs, error) {
	bs, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}
	var config []*FetchConfig
	if err = json.Unmarshal(bs, &config); err != nil {
		return nil, err
	}
	return config, nil
}

type FetchConfig struct {
	Type  string `json:"type"`
	Users []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"users"`
}

func (r *FetchConfig) NewAdapter() image_adapter.GetImager {
	switch r.Type {
	case "weibo":
		return image_adapter.NewWeibo()
	default:
		panic("unsupported")
	}
}

func (r *FetchConfig) Save(userID string, image *image_adapter.Image) error {
	dirname := fmt.Sprintf("./data/%s/%s", r.Type, userID)
	filename := fmt.Sprintf("%s/%s.json", dirname, image.ImageID)

	if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
		return err
	}

	bs, err := json.MarshalIndent(image, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, bs, os.ModePerm)
}

func (r *FetchConfig) Index(userID string) error {
	dirname := fmt.Sprintf("./data/%s/%s", r.Type, userID)
	fs, err := ioutil.ReadDir(dirname)
	if err != nil {
		return err
	}
	done := map[string]string{}
	for _, f := range fs {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		bs, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dirname, f.Name()))
		if err != nil {
			return err
		}
		var image image_adapter.Image
		if err := json.Unmarshal(bs, &image); err != nil {
			return err
		}
		if image.URL == "" || image.ImageID == "" {
			continue
		}
		done[image.ImageID] = image.URL
	}
	bs, err := json.MarshalIndent(done, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf("%s/index.json", dirname), []byte(bs), os.ModePerm)
}

type FetchConfigs []*FetchConfig

func (r FetchConfigs) Index() error {
	host := "https://chyroc.cn/beauty/data"
	filename := fmt.Sprintf("./data/index.json")
	res := []string{}
	for _, v := range r {
		for _, user := range v.Users {
			if user.Id == "" {
				continue
			}
			res = append(res, fmt.Sprintf("%s/%s/%s/", host, v.Type, user.Id))
		}
	}
	bs, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(bs), os.ModePerm)
}
