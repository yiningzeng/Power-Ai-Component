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
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//预声明变量
var dirPath string
var threadNum int
var m cmap.Cmap
var tagsMap cmap.Cmap

func read(fs afero.Fs, path string, n string, pool *gpool.Pool) {
	defer pool.Done()
	b, err := afero.ReadFile(fs, path)
	if err != nil {
		logrus.WithField("file", path).Error(err.Error())
	}
	var ki PowerAiAsset
	err = json.Unmarshal(b,  &ki)
	if err != nil {
		logrus.WithField("file", path).Error(err.Error())
	}
	if ki.Asset.Name != "" {
		ki.Asset.Path = "file:" + dirPath + "/" + ki.Asset.Name
		assetJson, err2 := json.Marshal(ki.Asset)
		if err2 != nil {
			logrus.WithField("file", path).Error(err2.Error())
		}
		m.Store(ki.Asset.Id, "\"" + ki.Asset.Id +"\":" + string(assetJson))
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
	//logrus.WithField("id", n).Info(ki.Asset.Name)
}

//绑定参数变量
func flagInit() {
	flag.StringVar(&dirPath, "path", "", "the path must be input")
	flag.IntVar(&threadNum, "tn", 100, "the default threadNum is 100")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
}

func LoggerInit(debug bool) {
	path := "."
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
	allTags := [...]string{"#FFB6C1","#FFC0CB","#DC143C","#FFF0F5","#DB7093","#FF69B4","#FF1493","#C71585","#DA70D6","#D8BFD8","#DDA0DD","#EE82EE","#FF00FF","#FF00FF","#8B008B","#800080","#BA55D3","#9400D3","#9932CC","#4B0082","#8A2BE2","#9370DB","#7B68EE","#6A5ACD","#483D8B","#E6E6FA","#F8F8FF","#0000FF","#0000CD","#191970","#00008B","#000080","#4169E1","#6495ED","#B0C4DE","#778899","#708090","#1E90FF","#F0F8FF","#4682B4","#87CEFA","#87CEEB","#00BFFF","#ADD8E6","#B0E0E6","#5F9EA0","#F0FFFF","#E1FFFF","#AFEEEE","#00FFFF","#D4F2E7","#00CED1","#2F4F4F","#008B8B","#008080","#48D1CC","#20B2AA","#40E0D0","#7FFFAA","#00FA9A","#00FF7F","#F5FFFA","#3CB371","#2E8B57","#F0FFF0","#90EE90","#98FB98","#8FBC8F","#32CD32","#00FF00","#228B22","#008000","#006400","#7FFF00","#7CFC00","#ADFF2F","#556B2F","#F5F5DC","#FAFAD2","#FFFFF0","#FFFFE0","#FFFF00","#808000","#BDB76B","#FFFACD","#EEE8AA","#F0E68C","#FFD700","#FFF8DC","#DAA520","#FFFAF0","#FDF5E6","#F5DEB3","#FFE4B5","#FFA500","#FFEFD5","#FFEBCD","#FFDEAD","#FAEBD7","#D2B48C","#DEB887","#FFE4C4","#FF8C00","#FAF0E6","#CD853F","#FFDAB9","#F4A460","#D2691E","#8B4513","#FFF5EE","#A0522D","#FFA07A","#FF7F50","#FF4500","#E9967A","#FF6347","#FFE4E1","#FA8072","#FFFAFA","#F08080","#BC8F8F","#CD5C5C","#FF0000","#A52A2A","#B22222","#8B0000","#800000"}

	//for i:=0; i<500000; i++ {
	//	f, _:=os.Open("../fucker/a.json")
	//	defer f.Close()
	//	tar, _ := os.Create("../fucker/" + strconv.Itoa(i) + ".json")
	//	defer tar.Close()
	//	io.Copy(tar, f)
	//}

	// 定义几个变量，用于接收命令行的参数值
	//var path       string
	//var threadNum    int
	//var host        string
	//var port        int
	//// &user 就是接收命令行中输入 -u 后面的参数值，其他同理
	//flag.StringVar(&user, "u", "root", "账号，默认为root")
	//flag.StringVar(&password, "p", "", "密码，默认为空")
	//flag.StringVar(&host, "h", "localhost", "主机名，默认为localhost")
	//flag.IntVar(&port, "P", 3306, "端口号，默认为3306")

	start := time.Now()
	pool := gpool.New(threadNum)
	fs := afero.NewOsFs()
	n := 0
	cmd := exec.Command("/bin/sh", "-c", "ls " + dirPath + " |grep .json > jsonList.txt")
	_, err := cmd.Output()
	if err!=nil{
		logrus.Error(err)
	}

	cc, _ := os.Open("jsonList.txt")
	defer cc.Close()
	br := bufio.NewReader(cc)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		pool.Add(1)
		go read(fs, dirPath + "/" + string(a), strconv.Itoa(n), pool)
	}



	//_ = afero.Walk(fs, dirPath, func(path string, info os.FileInfo, err error) error{
	//	if strings.Contains(info.Name(), ".json") == true {
	//		n++
	//		logrus.Info(info.Name())
	//		//if n > 19000 {
	//		//	return nil
	//		//} else {
	//			pool.Add(1)
	//			go read(fs, path, strconv.Itoa(n), pool)
	//		//}
	//	}
	//	//logrus.WithField("n", n).Info(info.Name())
	//	return nil
	//})
	pool.Wait()
	var mutex sync.Mutex
	fp := gpool.New(threadNum)
	projectJson := "{\"assets\": {"
	m.Range(func(key, value interface{}) bool {
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
		tags := PowerAiTags{Name: value.(string), Color: allTags[rand.Intn(127)]}
		tagsFinal = append(tagsFinal, tags)
		//logrus.WithField("color", allTags[rand.Intn(127)]).Error(value.(string))
		return true
	})

	tagsByte, _ := json.Marshal(tagsFinal)
	_ = afero.WriteFile(fs,  dirPath + "/colors.color", tagsByte, 0755)
	//if err != nil {
	//	//logrus.Error(err)
	//}
	fmt.Println("DONE", use)
}
