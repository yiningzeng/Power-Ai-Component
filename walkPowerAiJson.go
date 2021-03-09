package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/lrita/cmap"
	"github.com/msterzhang/gpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//预声明变量
var sort bool
var startNum int
var dirPath string
var threadNum int
var version = "3.3.3"
var assetsVisited cmap.Cmap
var assetsTagged cmap.Cmap
var assetsNotVisited cmap.Cmap
var tagsMap cmap.Cmap

/*
   NotVisited = 0,
   Visited = 1,
   Tagged = 2,*/
func read(fs afero.Fs, filePath string, pool *gpool.Pool) {
	defer pool.Done()
	bytes, err := afero.ReadFile(fs, filePath)
	if err != nil {
		logrus.WithField("file", filePath).Error(err.Error())
	}
	var ki PowerAiAsset
	err = json.Unmarshal(bytes,  &ki)
	if err != nil {
		logrus.WithField("file", filePath).Error(err.Error())
	}
	if ki.Asset.Name != "" {
		ki.Asset.Path = "file:" + path.Join(dirPath, ki.Asset.Name)
		assetJson, err2 := json.Marshal(ki.Asset)
		if err2 != nil {
			logrus.WithField("file", filePath).Error(err2.Error())
		}
		switch ki.Asset.State {
			case 0: // 未查看的数据
				assetsNotVisited.Store(ki.Asset.Id, "\"" + ki.Asset.Id +"\":" + string(assetJson))
				break
			case 1: // 已查看的数据
				assetsVisited.Store(ki.Asset.Id, "\"" + ki.Asset.Id +"\":" + string(assetJson))
				break
			case 2: // 已标记的数据
				assetsTagged.Store(ki.Asset.Id, "\"" + ki.Asset.Id +"\":" + string(assetJson))
				break
			default:
				break
		}
		if ki.Asset.Tags != "" {
			for _, v := range strings.Split(ki.Asset.Tags, ",") {
				if _, ok := tagsMap.Load(v); !ok {
					tagsMap.Store(v, v)
				}
			}
		}
		// 模拟生成测试的json文件
		/*ki.Asset.Name = n + ".jpg"
		ki.Asset.Id = n
		fi, _ := json.Marshal(ki)
		err := afero.WriteFile(fs,  "/home/baymin/daily-work/fucker-final/" + n +".json", fi, 0755)
		if err != nil {
			logrus.Error(err)
		}*/
	}
	startNum++
	if startNum%10 == 0 {
		runtime.GC()
		pool.Add(1)
		go func(p *gpool.Pool) {
			defer p.Done()
			exec.Command("/bin/bash", "-c", "echo " + strconv.Itoa(startNum) + " > now.txt").Run()
		}(pool)
	}
}

//绑定参数变量
func flagInit() {
	flag.StringVar(&dirPath, "path", ".", "the path must be input")
	flag.IntVar(&threadNum, "tn", 100, "the default threadNum is 100")
	flag.StringVar(&version, "v", version, "the version of monkeySun")
	flag.BoolVar(&sort, "sort", false, "sort the list")
	flag.Usage = func() {
		flag.PrintDefaults()
	}
}

func LoggerInit(debug bool) {
	path := "log/yiningzeng.log"
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	 `WithMaxAge` 设置文件清理前的最长保存时间
	 `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	if debug {
		logrus.SetFormatter(&logrus.TextFormatter{})
		//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
		//同时写文件和屏幕
		//fileAndStdoutWriter := io.MultiWriter([]io.Writer{writer, os.Stdout}...)
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		writer, _ := rotatelogs.New(
			path+".%Y%m%d%H%M",
			rotatelogs.WithLinkName(path),
			// rotatelogs.WithMaxAge(time.Duration(180)*time.),
			rotatelogs.WithRotationCount(60),
			rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		)
		logrus.SetOutput(writer)
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func main() {
	flagInit()
	LoggerInit(true)
	flag.Parse()
	var allTags []string
	if bytes, err := afero.ReadFile(afero.NewOsFs(), "colorsTemplate.json"); err == nil {
		if err = json.Unmarshal(bytes, &allTags); err != nil {
			logrus.Error(err)
			allTags = []string{"#FF3399", "#FF66FF","#99CCFF", "#33FF33", "#FFFFF", "#FFCC00", "#FFFF00", "#660099", "#0033FF", "#FF0033"}
		}
	} else {
		allTags = []string{"#FF3399", "#FF66FF","#99CCFF", "#33FF33", "#FFFFF", "#FFCC00", "#FFFF00", "#660099", "#0033FF", "#FF0033"}
	}

	exec.Command("/bin/bash", "-c", "echo 0 > now.txt").Run()
	start := time.Now()
	pool := gpool.New(threadNum)
	fs := afero.NewOsFs()
	// 由于linux上加载的远程文件夹遇到大量数据的情况下无法使用遍历文件的方法，只能先使用ls命令写入本地数据再处理
	var cmd *exec.Cmd
	if sort {
		cmd = exec.Command("/bin/sh", "-c", "ls -lr '" + dirPath + "' |grep .json|awk '{print $9}' > jsonList.txt")
	} else {
		cmd = exec.Command("/bin/sh", "-c", "ls '" + dirPath + "' |grep .json > jsonList.txt")
	}

	_, err := cmd.Output()
	if err != nil {
		logrus.Error(err)
	}

	if bytes, err := afero.ReadFile(fs, "jsonList.txt"); err == nil {
		allNum := strconv.Itoa(len(strings.Split(string(bytes), "\n")))
		_ = afero.WriteFile(fs,  "allNum.txt",  []byte(allNum), 0755)
	}

	cc, _ := os.Open("jsonList.txt")
	br := bufio.NewReader(cc)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		pool.Add(1)
		go read(fs, path.Join(dirPath, string(a)), pool)
	}
	pool.Wait()
	cc.Close()
	var mutex sync.Mutex
	fp := gpool.New(threadNum)
	projectJson := "{\"assets\": {"

	assetsTagged.Range(func(key, value interface{}) bool {
		fp.Add(1)
		go func(s string, pool2 *gpool.Pool) {
			runtime.GC()
			mutex.Lock()
			projectJson += s + ","
			mutex.Unlock()
			pool2.Done()
		}(value.(string), fp)
		return true
	})
	fp.Wait()

	assetsVisited.Range(func(key, value interface{}) bool {
		fp.Add(1)
		go func(s string, pool2 *gpool.Pool) {
			runtime.GC()
			mutex.Lock()
			projectJson += s + ","
			mutex.Unlock()
			pool2.Done()
		}(value.(string), fp)
		return true
	})
	fp.Wait()

	assetsNotVisited.Range(func(key, value interface{}) bool {
		fp.Add(1)
		go func(s string, pool2 *gpool.Pool) {
			runtime.GC()
			mutex.Lock()
			projectJson += s + ","
			mutex.Unlock()
			pool2.Done()
		}(value.(string), fp)
		return true
	})
	fp.Wait()

	projectJson = strings.TrimRight(projectJson, ",")
	projectJson += "}}"
	//logrus.Info("开始写入")
	err = afero.WriteFile(fs,  dirPath + "/.yiningzeng.assets", []byte(projectJson), 0755)
	projectJson = ""
	runtime.GC()
	if err != nil {
		logrus.Error(err)
	}
	use := time.Since(start)
	//fmt.Println(use)

	rand.Seed(time.Now().UnixNano())
	var tagsFinal []PowerAiTags
	tagsMap.Range(func(key, value interface{}) bool {
		tags := PowerAiTags{Name: value.(string), Color: allTags[rand.Intn(len(allTags) - 1)]}
		tagsFinal = append(tagsFinal, tags)
		//logrus.WithField("color", allTags[rand.Intn(127)]).Error(value.(string))
		return true
	})

	tagsByte, _ := json.Marshal(tagsFinal)
	_ = afero.WriteFile(fs,  dirPath + "/colors.color", tagsByte, 0755)
	//if err != nil {
	//	//logrus.Error(err)
	//}
	exec.Command("/bin/bash", "-c", "cat allNum.txt > now.txt").Run()
	fmt.Println("DONE", use)
}
