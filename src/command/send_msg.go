package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chyroc/beauty/src/image_adapter"
	"github.com/chyroc/lark"
)

func main() {
	webhook := os.Getenv("LARK_WEBHOOK_URL_1")
	appID := os.Getenv("LARK_APP_ID_1")
	appSecret := os.Getenv("LARK_APP_SECRET_1")
	larkCli := lark.New(
		lark.WithCustomBot(webhook, ""),
		lark.WithAppCredential(appID, appSecret),
		lark.WithTimeout(time.Minute),
	)

	images, err := loadImage()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(images))

	for _, imageData := range images {
		if err = sendImageToChat(larkCli, imageData); err != nil {
			panic(err)
		}
	}
}

var maxSize = 10485760 // 1024*1024*10

func loadImage() ([]*image_adapter.Image, error) {
	_ = os.MkdirAll("./cache", 0777)

	// load
	typeDirs, err := ioutil.ReadDir("./data")
	if err != nil {
		return nil, err
	}
	images := []*image_adapter.Image{}
	for _, typeDir := range typeDirs {
		if !typeDir.IsDir() {
			continue
		}
		userDirs, err := ioutil.ReadDir("./data/" + typeDir.Name())
		if err != nil {
			return nil, err
		}
		for _, userDir := range userDirs {
			if !typeDir.IsDir() {
				continue
			}
			fs, err := ioutil.ReadDir(fmt.Sprintf("./data/%s/%s", typeDir.Name(), userDir.Name()))
			if err != nil {
				return nil, err
			}
			for _, f := range fs {
				if !strings.HasSuffix(f.Name(), ".json") || f.Name() == "index.json" {
					continue
				}
				bs, err := ioutil.ReadFile(fmt.Sprintf("./data/%s/%s/%s", typeDir.Name(), userDir.Name(), f.Name()))
				if err != nil {
					return nil, err
				}
				var image image_adapter.Image
				if err = json.Unmarshal(bs, &image); err != nil {
					return nil, err
				}
				images = append(images, &image)
			}
		}
	}

	// remove done
	newImages := []*image_adapter.Image{}
	for _, image := range images {
		if isKeyExist(genKey(image)) {
			continue
		}
		newImages = append(newImages, image)
	}
	return newImages, nil
}

func sendImageToChat(larkCli *lark.Lark, image *image_adapter.Image) error {
	key := genKey(image)

	reader, err := downloadCompressImage(image)
	if err != nil {
		return err
	}

	res, _, err := larkCli.File.UploadImage(context.Background(), &lark.UploadImageReq{
		ImageType: "message",
		Image:     reader,
	})
	if err != nil {
		fmt.Println(key, "upload err", err.Error())
		return err
	}

	_, _, err = larkCli.Message.Send().SendImage(context.Background(), res.ImageKey)
	if err != nil {
		fmt.Println(key, "send message err", err.Error())
		return err
	}

	return saveKey(key)
}

func genKey(image *image_adapter.Image) string {
	return fmt.Sprintf("%s/%s/%s", image.Type, image.UserID, image.ImageID)
}

func isKeyExist(key string) bool {
	fullKey := fmt.Sprintf("./cache/%s", key)
	_, err := os.Stat(fullKey)
	return err == nil
}

func saveKey(key string) error {
	fullKey := fmt.Sprintf("./cache/%s", key)
	dir, _ := filepath.Split(fullKey)
	_ = os.MkdirAll(dir, 0777)
	return ioutil.WriteFile(fullKey, []byte(key), 0644)
}

func compressImageResource(ext string, data []byte, quality int) []byte {
	imgSrc, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}

	if strings.HasSuffix(ext, "png") {
		newImg := image.NewRGBA(imgSrc.Bounds())
		draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
		draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
		imgSrc = newImg
	} else {
		// imgSrc = imgSrc
	}

	buf := bytes.Buffer{}
	err = jpeg.Encode(&buf, imgSrc, &jpeg.Options{Quality: quality})
	if err != nil {
		return data
	}
	if buf.Len() > len(data) {
		return data
	}
	return buf.Bytes()
}

func downloadCompressImage(image *image_adapter.Image) (io.Reader, error) {
	key := genKey(image)

	body, err := downloadImage(image.URL)
	if err != nil {
		fmt.Println(key, "download err", err.Error())
		return nil, err
	}
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		fmt.Println(key, "read body err", err.Error())
		return nil, err
	}
	if len(bs) <= maxSize {
		return bytes.NewReader(bs), nil
	}
	quality := 95
	for {
		if quality <= 0 {
			return nil, fmt.Errorf("图片无法压缩")
		}
		newbs := compressImageResource(image.URL, bs, quality)
		if len(newbs) <= maxSize {
			return bytes.NewReader(newbs), nil
		}
		if len(newbs) == len(bs) {
			quality -= 5
		}
		bs = newbs
	}
}

func downloadImage(url string) (io.ReadCloser, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, err
}

var httpClient = &http.Client{
	Timeout: time.Second * 60,
}
