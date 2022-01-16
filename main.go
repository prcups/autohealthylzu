package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/thedevsaddam/gojsonq/v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ST struct {
	State int    `json:state`
	Msg   string `json:msg`
	Data  string `json:data`
}

var (
	st          ST
	accessToken string
	req         *http.Request
	md5         string
	info        map[string]interface{}
	err         error
	cardID      string
	username    string
	password    string
)

func GetST() chromedp.ActionFunc {
	var simpleCookies []http.Cookie
	return func(ctx context.Context) error {
		cookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		for _, cookie := range cookies {
			var simpleCookie http.Cookie
			simpleCookie.Name = cookie.Name
			simpleCookie.Value = cookie.Value
			simpleCookie.Domain = cookie.Domain
			simpleCookie.Path = cookie.Path
			simpleCookie.HttpOnly = cookie.HTTPOnly
			simpleCookie.Secure = cookie.Secure
			simpleCookies = append(simpleCookies, simpleCookie)
		}

		req, err = http.NewRequest("POST", "http://my.lzu.edu.cn/isExpire", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;")
		req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`)
		for i := len(simpleCookies) - 1; i >= 0; i-- {
			req.AddCookie(&simpleCookies[i])
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("POST", "http://my.lzu.edu.cn/api/getST", strings.NewReader("service=http://127.0.0.1"))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;")
		req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`)
		for i := len(simpleCookies) - 1; i >= 0; i-- {
			req.AddCookie(&simpleCookies[i])
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resStr, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resStr, &st)
		if err != nil {
			return err
		}
		//log.Println("st=" + st.Data)

		return nil
	}
}

func GetMD5() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var ok bool
		res, err := http.Get("https://appservice.lzu.edu.cn/dailyReportAll/api/auth/login?st=" + st.Data + "&PersonID=" + cardID)
		if err != nil {
			return err
		}
		resStr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		accessToken, ok = gojsonq.New().FromString(string(resStr)).Find("data.accessToken").(string)
		if !ok {
			return errors.New("accessToken Invaild")
		}
		//log.Println("accesstoken=" + accessToken)

		req, err = http.NewRequest("POST", "https://appservice.lzu.edu.cn/dailyReportAll/api/encryption/getMD5",
			strings.NewReader("cardId="+cardID))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;")
		req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`)
		req.Header.Set("Authorization", accessToken)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resStr, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		md5, ok = gojsonq.New().FromString(string(resStr)).Find("data").(string)
		if !ok {
			return errors.New("MD5 Invaild")
		}
		//log.Println("md5=" + md5)

		return nil
	}
}

func GetInfo() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		req, err = http.NewRequest("POST", "https://appservice.lzu.edu.cn/dailyReportAll/api/grtbMrsb/getInfo",
			strings.NewReader("cardId="+cardID+"&md5="+md5))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;")
		req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`)
		req.Header.Set("Authorization", accessToken)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resStr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		info = make(map[string]interface{})
		info = gojsonq.New().FromString(string(resStr)).Find("data.list.[0]").(map[string]interface{})
		info["sjd"] = gojsonq.New().FromString(string(resStr)).Find("data.sjd").(string)
		info["initLat"] = ""
		info["initLng"] = ""
		if err != nil {
			return err
		}
		//log.Printf("info=%+v\n", info)

		return nil
	}
}

func handleValue(value interface{}) string {
	switch value.(type) {
	case int:
		return strconv.Itoa(value.(int))
	case bool:
		return strconv.FormatBool(value.(bool))
	case nil:
		return ""
	case string:
		return value.(string)
	default:
		return ""
	case float64:
		return strconv.FormatFloat(value.(float64), 'f', -1, 64)
	}
}

func Submit() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var postForm string
		postForm += "bh=" + handleValue(info["bh"])
		postForm += "&xykh=" + handleValue(info["xykh"])
		postForm += "&twfw=" + handleValue(info["twfw"])
		postForm += "&jkm=" + handleValue(info["jkm"])
		postForm += "&sfzx=" + handleValue(info["sfzx"])
		postForm += "&sfgl=" + handleValue(info["sfgl"])
		postForm += "&szsf=" + handleValue(info["szsf"])
		postForm += "&szds=" + handleValue(info["szds"])
		postForm += "&szxq=" + handleValue(info["szxq"])
		postForm += "&sfcg=" + handleValue(info["sfcg"])
		postForm += "&cgdd=" + handleValue(info["cgdd"])
		postForm += "&gldd=" + handleValue(info["gldd"])
		postForm += "&jzyy=" + handleValue(info["jzyy"])
		postForm += "&bllb=" + handleValue(info["bllb"])
		postForm += "&sfjctr=" + handleValue(info["sfjctr"])
		postForm += "&jcrysm=" + handleValue(info["jcrysm"])
		postForm += "&xgjcjlsj=" + handleValue(info["xgjcjlsj"])
		postForm += "&xgjcjldd=" + handleValue(info["xgjcjldd"])
		postForm += "&xgjcjlsm=" + handleValue(info["xgjcjlsm"])
		postForm += "&zcwd=" + handleValue(info["zcwd"])
		postForm += "&zwwd=" + handleValue(info["zwwd"])
		postForm += "&wswd=" + handleValue(info["wswd"])
		postForm += "&sbr=" + handleValue(info["sbr"])
		postForm += "&sjd=" + handleValue(info["sjd"])
		postForm += "&initLng=" + handleValue(info["initLng"])
		postForm += "&initLat=" + handleValue(info["initLat"])
		postForm += "&dwfs=" + handleValue(info["dwfs"])
		//log.Println("send=" + postForm)

		req, err = http.NewRequest("POST", "https://appservice.lzu.edu.cn/dailyReportAll/api/grtbMrsb/submit",
			strings.NewReader(postForm))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;")
		req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/97.0.4692.71 Safari/537.36`)
		req.Header.Set("Authorization", accessToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resStr, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		code, ok := gojsonq.New().FromString(string(resStr)).Find("message").(string)
		if !ok {
			return errors.New("submit answer invalid")
		}
		if code != "成功" {
			return errors.New("get fail answer")
		} else {
			log.Println(time.Now().Format(time.RubyDate) + " 打卡成功")
		}
		return nil
	}
}

func main() {
	flag.StringVar(&cardID, "id", "", "your card id")
	flag.StringVar(&username, "username", "", "your username (no need to input @lzu.edu.cn)")
	flag.StringVar(&password, "password", "", "your password")
	flag.Parse()
	if cardID == "" {
		log.Fatal("card id is empty")
	}
	if username == "" {
		log.Fatal("username is empty")
	}
	if password == "" {
		log.Fatal("password is empty")
	}

	ctx, _ := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer chromedp.Cancel(ctx)

	err :=
		chromedp.Run(
			ctx,
			chromedp.Navigate("http://my.lzu.edu.cn:8080/login?service=http://my.lzu.edu.cn"),
			chromedp.WaitVisible(`#username`, chromedp.ByID),
			chromedp.WaitVisible(`#password`, chromedp.ByID),
			chromedp.SendKeys(`#username`, username, chromedp.ByID),
			chromedp.SendKeys(`#password`, password, chromedp.ByID),
			chromedp.Click(`#loginForm > div.btn-box > button`, chromedp.NodeVisible),
			chromedp.WaitReady("html"),
			GetST(),
			GetMD5(),
			GetInfo(),
			Submit(),
		)
	if err != nil {
		log.Fatal(err)
	}
	err = req.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
}
