package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	//https://github.com/PuerkitoBio/goquery
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

const (
	/**
	用户公开的视频地址
	*/
	UserInfoUrl   = "https://www.91porn.com/uprofile.php?UID=$"
	GirlPublicUrl = "https://www.91porn.com/uvideos.php?UID=$&type=public"
	VideoM3U8Url  = "https://cdn.91p07.com/m3u8/$/$.m3u8"
	VideoMP4Url   = "https://ccn.91p52.com/mp43/$.mp4"
	VideoUrl      = "https://www.91porn.com/view_video.php?viewkey="
)

var uid string
var vid string
var savePath string

/**
功能：下载https://www.91porn.com/ 某个大V用户视频
配置：视频合成依赖ffmpeg，所以需要先执行 brew install ffmpeg
*/
func main() {
	flag.StringVar(&savePath, "p", "/data/91movie/", "下载的视频存放目录")
	savePath = savePath + "/"
	flag.StringVar(&uid, "uid", "", "用户ID，例如https://www.91porn.com/uvideos.php?UID=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul ，-uid=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul")
	flag.StringVar(&vid, "vid", "", "视频ID，例如https://www.91porn.com/view_video.php?viewkey=8ee92162ba6b47e1dfcf， -vid=8ee92162ba6b47e1dfcf")
	flag.Parse()
	if uid != "" {
		pageNumber := getUserVideoPage(uid)
		for i := 1; i <= pageNumber; i++ {
			downloadAllVideo(uid, i)
		}
		return
	}
	if vid != "" {
		downloadSingleVideo(VideoUrl + vid)
		return
	}
	fmt.Println("excute end")
}

/**
获取用户公开视频分页数量
*/
func getUserVideoPage(uid string) int {
	pageSize := 8
	url := strings.ReplaceAll(UserInfoUrl, "$", uid)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
	}
	videoNumNode := doc.Find("a[href='" + strings.ReplaceAll(GirlPublicUrl, "$", uid) + "']")
	total := 1
	if videoNumNode != nil && len(videoNumNode.Nodes) > 0 {
		total, err = strconv.Atoi(videoNumNode.Nodes[1].FirstChild.Data)
	}
	return (total / pageSize) + 1
}

/**
下载某个用户所有公开视频
https://www.91porn.com/uvideos.php?UID=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul&type=public&page=2
*/
func downloadAllVideo(userId string, pageNumber int) {
	url := strings.ReplaceAll(GirlPublicUrl, "$", userId)
	url = url + "&page=" + strconv.FormatInt(int64(pageNumber), 10)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
	}

	doc.Find(".well-sm").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Find("a").Attr("href")
		downloadSingleVideo(url)
	})

}

/**
https://www.91porn.com/view_video.php?viewkey=d2e97bf0276d3f7ed6b0
*/
func downloadSingleVideo(url string) {
	fmt.Printf("excute [%s] page track", url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept-language", "zh-CN")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
	}

	userNameNodes := doc.Find(".title")
	userName := ""
	if userNameNodes != nil && len(userNameNodes.Nodes) > 0 {
		userName = userNameNodes.Nodes[0].FirstChild.Data
	}

	userTitleNodes := doc.Find(".login_register_header")
	title := ""
	if userTitleNodes != nil && len(userTitleNodes.Nodes) > 0 {
		title = userTitleNodes.Nodes[0].FirstChild.Data
		title = strings.Trim(title, " ")
	}

	//获取m3u8视频ID
	nodes := doc.Find("#VID")
	if nodes != nil && len(nodes.Nodes) > 0 {
		vid := nodes.Nodes[1].FirstChild.Data
		createParentFile(userName)
		downLoad(vid, title, userName, url)
	} else {
		fmt.Println("skip video", url)
	}
}

func createParentFile(name string) {
	_, err := os.Stat(savePath + name)
	//如果返回的错误为nil,说明文件或文件夹存在
	if err == nil {
	} else {
		err := os.MkdirAll(savePath+name, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
}

func checkFileExists(path string) bool {
	_, err := os.Stat(path)
	//存在
	if err == nil {
		return true
	}
	return false
}

/**
下载视频
*/
func downLoad(vid string, title string, userName string, url string) {
	videoUrl := strings.ReplaceAll(VideoM3U8Url, "$", vid)
	videoMP4Url := strings.ReplaceAll(VideoMP4Url, "$", vid)
	saveFilePath := savePath + userName + "/" + title + ".mp4"
	if checkFileExists(saveFilePath) {
		fmt.Printf("【 %s 】video exists,skip downlaod \n", saveFilePath)
		return
	}
	fmt.Printf("download video begin video=%s \n", title)
	//是否是老的视频
	if checkVideoUrlIsOld(videoUrl) {
		downLoadOld(videoMP4Url, saveFilePath, url)
	} else {
		downLoadNew(videoUrl, saveFilePath)
	}
	fmt.Println("download video end")
}

/**
下载老的视频 https://ccn.91p52.com/mp43/384739.mp4
老的视频需要调用chromedp，页面的<source>节点是js动态生成的
*/
func downLoadOld(videoMP4Url string, saveFilePath string, url string) {
	chromeCtx, cancel := chromedp.NewContext(context.Background(), chromedp.WithLogf(log.Printf))
	timeOutCtx, cancel := context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(timeOutCtx,
		chromedp.Navigate(url),
		//等待某个特定的元素出现
		//chromedp.Sleep(2 * time.Second),
		chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		//生成最终的html文件并保存在htmlContent文件中
	)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal(err)
	}
	//获取m3u8视频ID
	nodes := doc.Find("source")
	if nodes != nil && len(nodes.Nodes) > 0 {
		videoMP4Url, _ = nodes.Attr("src")
	} else {
		fmt.Println("get skip video", url)
		return
	}

	fmt.Println("download old video:", videoMP4Url)
	//videoMP4Url 不可以用，需要动态获取
	resp, err := http.Get(videoMP4Url)
	if err != nil {
		//panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//panic(err)
	}
	ioutil.WriteFile(saveFilePath, data, 0644)
}

/**
调用mac shell  执行下载
ffmpeg -i https://cdn.91p07.com//m3u8/424352/424352.m3u8 -c copy -bsf:a aac_adtstoasc output2.mp4
*/
func downLoadNew(videoUrl string, saveFilePath string) {
	binary, lookErr := exec.LookPath("ffmpeg")
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{
		"-i",
		videoUrl,
		"-acodec",
		"copy",
		"-vcodec",
		"copy",
		saveFilePath,
	}
	cmd := exec.Command(binary, args...)

	stderr, _ := cmd.StderrPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}
	cmd.Wait()
}

/**
视频链接有两种
https://ccn.91p52.com/mp43/384739.mp4
下面这种只能用ffmpeg shell下载
https://cdn.91p07.com//m3u8/425408/425408.m3u8
*/
func checkVideoUrlIsOld(url string) bool {
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return true
	}
	return false
}
