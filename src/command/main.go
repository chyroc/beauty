package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
		}
	}
}

// config
func LoadConfig() ([]*FetchConfig, error) {
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
