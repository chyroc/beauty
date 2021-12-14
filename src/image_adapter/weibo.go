package image_adapter

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/chyroc/beauty/src/helper"
)

func NewWeibo() GetImager {
	return &weiboImpl{}
}

type weiboImpl struct {
}

func (r *weiboImpl) GetImage(data string) ([]*Image, error) {
	uid := data
	userInfo, err := getUserInfo(uid)
	if err != nil {
		return nil, err
	}
	containerID := helper.GetOneMatchString(userInfo.Data.More, regexp.MustCompile(`(\d+)`))

	cards := []*getContainerRespCard{}
	for _, page := range []int{1, 2} {
		data, err := getContainerCard(uid, containerID, page)
		if err != nil {
			return nil, err
		}
		cards = append(cards, data...)
	}

	imaegs := []*Image{}
	for _, card := range cards {
		imaegs = append(imaegs, card.ToImages(uid)...)
	}
	return imaegs, err
}

func getUserInfo(uid string) (*userInfo, error) {
	uri := fmt.Sprintf("https://m.weibo.cn/profile/info?uid=%s", uid)
	resp := new(userInfo)
	err := helper.Request.New(http.MethodGet, uri).Unmarshal(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

type userInfo struct {
	Data struct {
		User struct {
			ScreenName      string `json:"screen_name"`
			Description     string `json:"description"`
			ProfileImageURL string `json:"profile_image_url"`
		}
		More string
	}
}

func getContainerCard(uid, containerID string, page int) ([]*getContainerRespCard, error) {
	uri := fmt.Sprintf("https://m.weibo.cn/api/container/getIndex?uid=%s&containerid=%s&since_id=0&page=%d", uid, containerID, page)
	resp := new(getContainerResp)
	err := helper.Request.New(http.MethodGet, uri).WithHeaders(map[string]string{
		"Referer":          "https://m.weibo.cn/u/" + uid,
		"MWeibo-Pwa":       "1",
		"X-Requested-With": "XMLHttpRequest",
	}).Unmarshal(resp)
	if err != nil {
		return nil, err
	}
	return resp.Data.Cards, nil
}

type getContainerResp struct {
	Data struct {
		Cards []*getContainerRespCard `json:"cards"`
	} `json:"data"`
}

type getContainerRespCard struct {
	Mblog *getContainerRespCardMblog `json:"mblog"`
}

func (r *getContainerRespCard) ToImages(userID string) []*Image {
	if r == nil || r.Mblog == nil || r.Mblog.RetweetedStatus != nil || len(r.Mblog.Pics) == 0 {
		return nil
	}

	imaegs := []*Image{}
	for _, pic := range r.Mblog.Pics {
		image := &Image{ImageID: pic.Pid, URL: pic.Large.URL, UserID: userID}
		if image.URL == "" {
			image.URL = pic.URL
		}
		imaegs = append(imaegs, image)
	}
	return imaegs
}

type getContainerRespCardMblog struct {
	RetweetedStatus interface{} `json:"retweeted_status"`
	// "created_at": "Sun Sep 26 23:57:09 +0800 2021",
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Mid       string `json:"mid"`
	Text      string `json:"text"`
	User      struct {
		ID         int    `json:"id"`
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	IsLongText bool `json:"isLongText"`
	PicNum     int  `json:"pic_num"`
	Pics       []struct {
		Pid   string `json:"pid"`
		URL   string `json:"url"`
		Size  string `json:"size"`
		Large struct {
			Size string `json:"size"`
			URL  string `json:"url"`
		} `json:"large"`
	} `json:"pics"`
	Bid string `json:"bid"`
}
