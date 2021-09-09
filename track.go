package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"text/template"
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
	UserInfoUrl     = "https://www.91porn.com/uprofile.php?UID=$"
	GirlPublicUrl   = "https://www.91porn.com/uvideos.php?UID=$&type=public"
	VideoM3U8Url    = "https://cdn.91p07.com/m3u8/$/$.m3u8"
	VideoMP4Url     = "https://ccn.91p52.com/mp43/$.mp4"
	VideoUrl        = "https://www.91porn.com/view_video.php?viewkey="
	HotVideIndexUrl = "https://www.91porn.com/v.php?"
	//ffmpeg下载超时时间5分钟
	FfmpegTimeOut = 5 * 60 * 1000000
)

var videoList = make([]Video, 0)

var videoTypeMap = map[int]string{
	0:  "所有",
	1:  "当前最热",
	2:  "本月最热",
	3:  "10分钟以上",
	4:  "20分钟以上",
	5:  "本月收藏",
	6:  "收藏最多",
	7:  "最近加精",
	8:  "高清",
	9:  "本月讨论",
	10: "10分钟以上",
}

/**
用户ID
*/
var uid string

/**
视频ID
*/
var vid string
var pNumber string

/**
保存地址
*/
var savePath string

/**
类型
*/
var hotVideoType string

/**
功能：下载https://www.91porn.com/ 某个大V用户视频
配置：视频合成依赖ffmpeg，所以需要先执行 brew install ffmpeg
*/
func main() {
	flag.StringVar(&savePath, "p", "/data/91movie/", "下载的视频存放目录")
	savePath = savePath + "/"
	flag.StringVar(&uid, "uid", "", "用户ID，https://www.91porn.com/uvideos.php?UID=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul ，例如 -uid=c24dDoGZBAnwUtBbHweSJB8W6ACe8c7sJyQOJ9Af4DQ4sxul")
	flag.StringVar(&hotVideoType, "t", "", "视频类型"+
		"0:所有"+
		"1: 91原创 "+
		"2：当前最热 "+
		"3：本月最热 "+
		"4：10分钟以上 "+
		"5：20分钟以上 "+
		"6：本月收藏 "+
		"7：收藏最多 "+
		"8：最近加精 "+
		"9：高清 "+
		"10：上月最热 "+
		"11：本月讨论 "+
		"，例如 -t=1")

	flag.StringVar(&vid, "vid", "", "视频ID，https://www.91porn.com/view_video.php?viewkey=8ee92162ba6b47e1dfcf， 例如 -vid=8ee92162ba6b47e1dfcf")

	flag.StringVar(&pNumber, "n", "0", "历史数据爬去分页 例如 -n=2000")
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

	if hotVideoType != "" {
		downLoadHotVideo(hotVideoType)
		buildIndex()
		return
	}

	fmt.Println("excute end")
}

/**
创建索引
*/
func buildIndex() {
	videoType, _ := strconv.Atoi(hotVideoType)
	videTypeName := videoTypeMap[videoType]
	if len(videoList) > 0 {
		videolist := VideoList{Videolist: videoList}
		tmpl, err := template.ParseFiles("index.go.tpl")
		if err != nil {
		}
		indexHtmlSavePath := savePath + "/" + videTypeName + ".html"
		//存在
		if err == nil {
			os.Remove(indexHtmlSavePath)
		}
		file, err := os.OpenFile(indexHtmlSavePath, os.O_CREATE|os.O_WRONLY, 0755)
		_, err = os.Stat(indexHtmlSavePath)
		err = tmpl.Execute(file, videolist)
	}
}

func downLoadHotVideo(videoTypeStr string) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	chromeCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

	ctx, cancel := chromedp.NewContext(chromeCtx)
	defer cancel()
	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(HotVideIndexUrl),
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
	//https://www.91porn.com/v.php?category=hot&viewtype=basic
	hotVoideoLinkNode := doc.Find(".navbar-right").Find("a")
	if hotVoideoLinkNode != nil && len(hotVoideoLinkNode.Nodes) > 0 {
		videoType, _ := strconv.Atoi(videoTypeStr)
		//视频从第6个开始
		linkUrl := hotVoideoLinkNode.Get(videoType + 5).Attr[0].Val
		maxPageNumber := 6000
		if videoType == 0 {
			linkUrl = HotVideIndexUrl
		}
		fmt.Println(linkUrl)
		flag := true
		//pageNumer :=
		pageNumer, _ := strconv.Atoi(pNumber)
		viewMap := make(map[string]string)
		for flag {
			options := []chromedp.ExecAllocatorOption{
				chromedp.Flag("headless", true),
				chromedp.Flag("hide-scrollbars", false),
				chromedp.Flag("mute-audio", false),
				chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
			}
			options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
			chromeCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

			ctx, cancel := chromedp.NewContext(chromeCtx)
			err := chromedp.Run(ctx,
				chromedp.Navigate(linkUrl+"&page="+strconv.FormatInt(int64(pageNumer), 10)),
				//等待某个特定的元素出现
				//chromedp.Sleep(2 * time.Second),
				chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
				//生成最终的html文件并保存在htmlContent文件中
			)
			/*if err != nil {
				log.Fatal(err)
			}
			*/
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
			if err != nil {
				log.Fatal(err)
			}
			if err != nil {
				log.Println(err)
			}
			//https://www.91porn.com/v.php?category=hot&viewtype=basic
			doc.Find(".videos-text-align").Each(func(i int, s *goquery.Selection) {
				urls, _ := s.Find("a").Attr("href")
				values, err := url.ParseQuery(urls)
				//videoType为0 可能执行的时候又有视频创建导致提前重复
				if videoType != 0 && err == nil {
					viewkey := values.Get("https://91porn.com/view_video.php?viewkey")
					//存在重复的 直接跳出循环，解决分页问题
					if viewMap[viewkey] != "" {
						flag = false
						fmt.Println("repeate skip key=", viewkey)
						return
					}
					viewMap[viewkey] = viewkey
				}
				downloadSingleVideo(urls)
			})
			if pageNumer >= maxPageNumber {
				fmt.Println("break pageNumer=", pageNumer)
				break
			}
			pageNumer++
			cancel()
		}
	}
}

/**
获取用户公开视频分页数量
*/
func getUserVideoPage(uid string) int {

	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:51837", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
	}
	pageSize := 8
	url := strings.ReplaceAll(UserInfoUrl, "$", uid)
	res, err := myClient.Get(url)
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
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:51837", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
	}
	url := strings.ReplaceAll(GirlPublicUrl, "$", userId)
	url = url + "&page=" + strconv.FormatInt(int64(pageNumber), 10)
	res, err := myClient.Get(url)
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
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:51837", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
	}

	fmt.Printf("excute [%s] page track \n", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept-language", "zh-CN")
	res, err := myClient.Do(req)
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
	if userNameNodes != nil && len(userNameNodes.Nodes) > 0 && userNameNodes.Nodes[0].FirstChild != nil {
		userName = userNameNodes.Nodes[0].FirstChild.Data
	}

	userTitleNodes := doc.Find(".login_register_header")
	title := ""
	if userTitleNodes != nil && len(userTitleNodes.Nodes) > 0 {
		title = userTitleNodes.Nodes[0].FirstChild.Data
		if title == " " {
			title = userTitleNodes.Nodes[0].LastChild.Data
		}
		title = strings.Trim(title, " ")
	}
	//

	imageNodes := doc.Find("#player_one")
	videoImage := ""
	if imageNodes != nil && len(imageNodes.Nodes) > 0 {
		videoImage, _ = imageNodes.Attr("poster")
	}
	//获取m3u8视频ID
	nodes := doc.Find("#VID")
	if nodes != nil && len(nodes.Nodes) > 0 {
		vid := nodes.Nodes[1].FirstChild.Data
		createParentFile(userName)
		downLoad(vid, title, userName, url, videoImage)
	} else {
		data, _ := ioutil.ReadAll(res.Body)
		log.Println("status code error: %d %s %s", res.StatusCode, res.Status, string(data))
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
func downLoad(vid string, title string, userName string, url string, videoImage string) {
	videoUrl := strings.ReplaceAll(VideoM3U8Url, "$", vid)
	videoMP4Url := strings.ReplaceAll(VideoMP4Url, "$", vid)
	saveFilePath := savePath + userName + "/" + title + ".mp4"

	if hotVideoType != "" {
		videoList = append(videoList, Video{ImgUrl: videoImage, Title: title, Path: saveFilePath})
	}
	if checkFileExists(saveFilePath) {
		fmt.Printf("【 %s 】video exists,skip download ", saveFilePath)
		return
	}
	fmt.Printf("download video begin video=%s \n", saveFilePath)
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
		fmt.Println("skip video", url)
		return
	}

	fmt.Println("download old video:", videoMP4Url)
	//videoMP4Url 不可以用，需要动态获取

	videoMP4Url = strings.ReplaceAll(videoMP4Url, "ccn.91p52.com", "cv.91p52.com")
	resp, err := http.Get(videoMP4Url)
	if err != nil {
		//panic(err)
		return
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
	//vip地址
	videoUrl = strings.ReplaceAll(videoUrl, "cdn.91p07.com", "cv.killcovid2021.com")
	binary, lookErr := exec.LookPath("ffmpeg")
	if lookErr != nil {
		panic(lookErr)
	}
	args := []string{
		"-rw_timeout",
		strconv.Itoa(FfmpegTimeOut),
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
	dialSocksProxy, err := proxy.SOCKS5("tcp", "127.0.0.1:51837", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}
	tr := &http.Transport{Dial: dialSocksProxy.Dial}

	// Create client
	myClient := &http.Client{
		Transport: tr,
	}

	//url = strings.ReplaceAll(url, "cdn.91p07.com", "cv.91p52.com")
	res, err := myClient.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return true
	}
	return false
}

type Video struct {
	ImgUrl string `json:"imgUrl"`
	Title  string `json:"title"`
	Path   string `json:"path"`
}

type VideoList struct {
	Videolist []Video
}
