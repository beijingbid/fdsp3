package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	//"encoding/base64"
	"log"
	//"database/sql"
	"encoding/json"
	"flag"
	"strconv"

	"github.com/larspensjo/config"
)

var g_logger_info *log.Logger  //常规日志对象
var g_logger_debug *log.Logger //调试日志对象
var g_logger_err *log.Logger   //错误日志对象

var g_config map[string]string //配置信息

var adinfo_map map[int]adinfo
var adinfo_map_lock sync.RWMutex

// Adinfo 信息
type adinfo struct {
	Adid     int
	Price    int
	Level    int
	Weight   int
	Is_https int
	Banner   banner
	Video    video
	Native   native
	Ext      adinfo_ext
}

type adinfo_ext struct {
	Action         int
	Imptrackers    []string
	Clktrackers    []string
	Clkurl         string
	Html_snippet   string
	Inventory_type int
}

type banner struct {
	Weight       int
	Height       int
	Src          string
	BannerAdType int
	Mime         string
}

type video struct {
	Weight    int
	Height    int
	Src       string
	Mime      string
	Linearity bool
	Duration  int
	Protocol  int
}

type native struct {
	ver    string
	Assets []asset
}

type asset struct {
	Id          int
	Asset_oneof int
	Title       native_title
	Image       native_image
	Video       native_video
	Data        native_data
}

type native_title struct {
	Len   int
	Title string
}
type native_image struct {
	Width          int
	Height         int
	Src            string
	Mime           string
	ImageAssetType int
}
type native_video struct {
	Width     int
	Height    int
	Src       string
	Mime      string
	Linearity bool
	Duration  int
}
type native_data struct {
	DataAssetType int
	Len           int
	Data          string
}

// Adinfo 信息

// request 信息
type request struct {
	Id     string
	Imp    []request_imp
	App    request_app
	Device request_device
	Ext    request_ext
}

type request_imp struct {
	Id          string
	Banner      request_imp_banner
	Instl       bool
	Tagid       string
	Bidfloor    int
	Bidfloorcur string
	Ext         request_imp_banner_ext
}
type request_imp_banner struct {
	W   int
	H   int
	Pos int
}

type request_imp_banner_ext struct {
	Inventory_typs []int
	Ad_type        int
	Tag_name       string
}
type request_app struct {
	Id     string
	Name   string
	Ver    string
	Bundle string
}
type request_device struct {
	Dnt             bool
	Ua              string
	Ip              string
	Geo             request_device_geo
	Didsha1         string
	Dpidsha1        string
	Make            string
	Model           string
	Os              string
	Osv             string
	W               int
	H               int
	Ppi             int
	Connectyiontype int
	Devicetype      int
	Macsha1         string
	Ext             request_device_ext
}

type request_device_ext struct {
	Plmm        string
	Imei        string
	Imsi        string
	Mac         string
	Android_id  string
	Adid        string
	Orientation int
}
type request_device_geo struct {
	Lat     float64
	Lon     float64
	Country string
	Region  string
	City    string
	Type    int
	Ext     request_device_geo_ext
}

type request_device_geo_ext struct {
	Accu int
}
type request_ext struct {
	Version    int
	Need_https bool
}

// request 信息

// response 信息
type response struct {
	Id      string             `json:"id"`
	Seatbid []response_seatbid `json:"seatbid"`
}

type response_seatbid struct {
	Bid []response_seatbid_bid `json:"bid"`
}
type response_seatbid_bid struct {
	Id    string                   `json:"id"`
	Impid string                   `json:"impid"`
	Price float64                  `json:"price"`
	Adid  string                   `json:"adid"`
	W     int                      `json:"w"`
	H     int                      `json:"h"`
	Iurl  string                   `json:"iurl"`
	Adm   string                   `json:"adm"`
	Ext   response_seatbid_bid_ext `json:"ext"`
}

type response_seatbid_bid_ext struct {
	Clkurl         string   `json:"clkurl"`
	Imptrackers    []string `json:"imptrackers"`
	Clktrackers    []string `json:"clktrackers"`
	Title          string   `json:"title"`
	Desc           string   `json:"desc"`
	Action         int      `json:"action"`
	Html_snippet   string   `json:"html_snappet"`
	Inventory_type int      `json:"inventory_type"`
}

// response 信息

func loadConf(filename string, hp_list *[]string) {
	inputFile, inputError := os.Open(filename)
	if inputError != nil {
		fmt.Println("An error occurred on opening the inputfile\n")
		return
	}
	defer inputFile.Close()
	inputReader := bufio.NewReader(inputFile)
	for {
		inputString, readerError := inputReader.ReadString('\n')
		if readerError == io.EOF {
			return
		}
		inputString = strings.Replace(inputString, "\n", "", 1)
		inputString = strings.Replace(inputString, "\r", "", 1)
		inputString = strings.Replace(inputString, "\t", "", 1)
		(*hp_list) = append((*hp_list), inputString)
	}
	return
}

func Hello(writer http.ResponseWriter, req *http.Request) {
	g_logger_info.Println("Hello")
	fmt.Println(" this is Hello")
	fmt.Fprintln(writer, " this is Hello")
	return
}

func ListAdinfo(writer http.ResponseWriter, req *http.Request) {
	g_logger_info.Println("ListAdinfo")
	fmt.Println(" this is ListAdinfo")
	fmt.Fprintln(writer, " this is ListAdinfo")
	for key, value := range adinfo_map {
		fmt.Fprintln(writer, " key:", key, " value:", value)
	}
	return
}

func LoadAdinfo(writer http.ResponseWriter, req *http.Request) {
	g_logger_info.Println("LoadAdinfo")
	fmt.Println(" this is LoadAdinfo")
	fmt.Fprintln(writer, " this is LoadAdinfo")

	inputFile, inputError := os.Open(g_config["adinfo_file"])
	if inputError != nil {
		fmt.Fprintln(os.Stderr, time.Now(), "An error occurred on opening: ipblacklist.conf", inputError)
		return
	}
	defer inputFile.Close()
	inputReader := bufio.NewReader(inputFile)
	adinfo_map_lock.Lock()
	adinfo_map = make(map[int]adinfo)
	i := 0
	for {
		inputString, readerError := inputReader.ReadString('\n')
		inputString = strings.Replace(inputString, "\n", "", 1)
		inputString = strings.Replace(inputString, "\r", "", 1)
		inputString = strings.Replace(inputString, "\t", "", 1)
		if readerError == io.EOF {
			break
		}
		if len(inputString) > 0 {
			//adinfo_map[i] = inputString
		}
		var someOne adinfo
		if err := json.Unmarshal([]byte(inputString), &someOne); err == nil {
			fmt.Println("someOne:", someOne)
			fmt.Println("someOne.adid:", someOne.Adid)
		} else {
			fmt.Println("someOne.adid err :", err)
		}
		fmt.Println(" i:", i, " inputString:", inputString)

		fmt.Fprintln(writer, " i:", i, " inputString:", inputString)
		adinfo_map[someOne.Adid] = someOne
		i++
	}
	adinfo_map_lock.Unlock()
	return
}

func ClearAdinfo(writer http.ResponseWriter, req *http.Request) {
	g_logger_info.Println("ClearAdinfo")
	fmt.Println(" this is ClearAdinfo")
	fmt.Fprintln(writer, " this is ClearAdinfo")
	adinfo_map_lock.Lock()
	adinfo_map = make(map[int]adinfo)
	adinfo_map_lock.Unlock()
	return
}

func GetAdJson(writer http.ResponseWriter, req *http.Request) {
	g_logger_info.Println("GetAdJson")
	fmt.Println(" this is GetAdJson")
	//fmt.Fprintln(writer," this is GetAdJson")

	inputString, _ := ioutil.ReadAll(req.Body) //把  body 内容读入字符串 s
	//fmt.Fprintln(writer, "%s", string([]byte(inputString)) )       //在返回页面中显示内容。
	fmt.Println("inputString is ", string([]byte(inputString))) //在返回页面中显示内容。
	var requestJson request
	var responseJson response
	if err := json.Unmarshal([]byte(inputString), &requestJson); err == nil {
		//fmt.Fprintln(writer,"requestJson:",requestJson)
		//fmt.Fprintln(writer,"requestJson.id:",requestJson.Id)

		//responseJson = searchAd(requestJson)
		if Adinfo, err := searchAd(requestJson); err == nil {

			responseJson.Id = requestJson.Id
			// 这里继续填写广告信息

			var responseJson_seatbid response_seatbid
			var responseJson_seatbid_bid response_seatbid_bid
			var responseJson_seatbid_bid_ext response_seatbid_bid_ext
			responseJson_seatbid_bid.Id = "id-" + strconv.Itoa(Adinfo.Adid)
			responseJson_seatbid_bid.Impid = "impid-" + strconv.Itoa(Adinfo.Adid)
			responseJson_seatbid_bid.Price = float64(Adinfo.Price) / 100
			responseJson_seatbid_bid.Adid = "Adid-" + strconv.Itoa(Adinfo.Adid)
			responseJson_seatbid_bid.W = Adinfo.Banner.Weight
			responseJson_seatbid_bid.H = Adinfo.Banner.Height
			responseJson_seatbid_bid.Iurl = Adinfo.Banner.Src
			responseJson_seatbid_bid.Adm = ""

			responseJson_seatbid_bid_ext.Clkurl = Adinfo.Ext.Clkurl
			for _, tmp := range Adinfo.Ext.Imptrackers {
				responseJson_seatbid_bid_ext.Imptrackers = append(responseJson_seatbid_bid_ext.Imptrackers, tmp)
			}
			for _, tmp := range Adinfo.Ext.Clktrackers {
				responseJson_seatbid_bid_ext.Clktrackers = append(responseJson_seatbid_bid_ext.Clktrackers, tmp)
			}
			responseJson_seatbid_bid_ext.Action = Adinfo.Ext.Action
			responseJson_seatbid_bid_ext.Inventory_type = Adinfo.Ext.Inventory_type

			responseJson_seatbid_bid.Ext = responseJson_seatbid_bid_ext
			responseJson_seatbid.Bid = append(responseJson_seatbid.Bid, responseJson_seatbid_bid)

			responseJson.Seatbid = append(responseJson.Seatbid, responseJson_seatbid)
			//fmt.Printf(responseJson.Seatbid)
		}

	} else {
		fmt.Fprintln(writer, "requestJson.id err :", err)
	}

	//fmt.Fprintln(writer," responseJson is ",responseJson)
	jsonStr, err := json.Marshal(responseJson)
	if err != nil {
		fmt.Fprintln(writer, " responseJson is failed ", err)
	}
	fmt.Fprintln(writer, string(jsonStr))

	return
}

func searchAd(requestJson request) (adinfo adinfo, err error) {

	// 筛选广告

	//

	// 级别筛选

	// 比重随机

	// 获取最终结果

	if adinfo, err := getAdInfo(123); err == nil {

		fmt.Println("searchAd success ", adinfo)
		return adinfo, nil
	} else {
		fmt.Println("searchAd failed ")
		var err error
		err = errors.New("searchAd failed ")
		return adinfo, err
	}

}

func getAdInfo(adid int) (adinfo adinfo, err error) {
	if _, ok := adinfo_map[adid]; ok {
		return adinfo_map[adid], nil
	} else {
		var err error
		err = errors.New("adinfo null")
		return adinfo, err
	}
}

//-------------loadConfig-----------
//读取配置文件
//参数: domain 配置分区
//返回值: true 成功，false 失败
//--------------------------------------
func loadConfig(domain string) bool {
	pConfigFile := flag.String("configfile", "config.ini", "system configuration file")
	cfg, err := config.ReadDefault(*pConfigFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, time.Now(), "Fail to load config, Err:", err)
		return false
	}

	if cfg.HasSection(domain) {
		section, err := cfg.SectionOptions(domain)
		if err != nil {
			fmt.Fprintln(os.Stderr, time.Now(), "Fail to load config, Err:", err)
			return false
		}
		g_config = make(map[string]string)
		for _, key := range section {
			value, err := cfg.String(domain, key)
			if err == nil {
				g_config[key] = value
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, time.Now(), "Fail to load config, Err: default section not found")
		return false
	}

	return true
}

//-------------initLog-----------
//初始化日志对象
//参数: 无
//返回值: true 成功，false 失败
//--------------------------------------
func initLog() bool {
	fileSuffix := ""
	if g_config["info_log_split"] == "day" {
		fileSuffix = time.Now().Format(".2006-01-02")
	} else if g_config["info_log_split"] == "hour" {
		fileSuffix = time.Now().Format(".2006-01-02-15")
	} else if g_config["info_log_split"] == "min" {
		fileSuffix = time.Now().Format(".2006-01-02-15-04")
	}

	tasklogfile, err := os.OpenFile(g_config["info_log"]+fileSuffix, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		fmt.Printf("initLog Err:", err)
		return false
	}
	g_logger_info = log.New(tasklogfile, "[INFO]", log.Ldate|log.Ltime)
	g_logger_debug = log.New(tasklogfile, "[DEBUG]", log.Ldate|log.Ltime)

	tasklogfile, err = os.OpenFile(g_config["err_log"], os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		fmt.Printf("initLog Err:", err)
		return false
	}
	g_logger_err = log.New(tasklogfile, "[ERROR]", log.Ldate|log.Ltime|log.Lshortfile)

	return true
}

func logSplitThread() {
	for {

		currentTime := time.Now()
		//每分钟0秒判断是否需要生成新日志文件
		if currentTime.Second() == 0 {
			if g_config["info_log_split"] == "day" && currentTime.Hour() == 0 && currentTime.Minute() == 0 {
				initLog()
			} else if g_config["info_log_split"] == "hour" && currentTime.Minute() == 0 {
				initLog()
			} else if g_config["info_log_split"] == "min" {
				initLog()
			}
		}
		time.Sleep(time.Second)
	}
}
func main() {
	//读取配置文件
	if loadConfig("fdsp") == false {
		return
	}
	//日志初始化
	if initLog() == false {
		return
	}
	//开启日志拆分线程
	go logSplitThread()
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.Handle("/", http.HandlerFunc(Hello))
	http.Handle("/list", http.HandlerFunc(ListAdinfo))
	http.Handle("/load", http.HandlerFunc(LoadAdinfo))
	http.Handle("/clear", http.HandlerFunc(ClearAdinfo))
	http.Handle("/json", http.HandlerFunc(GetAdJson))

	//err := http.ListenAndServe(":8141", nil)
	err := http.ListenAndServe(":"+g_config["http_port"], nil)
	if err != nil {
		g_logger_err.Println("Err", err)
	}

}
