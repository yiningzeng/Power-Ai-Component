package main

import (
	"encoding/json"
	"fmt"
	"github.com/msterzhang/gpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"os"
	"time"
)
var fuckerList [250000]string
func read(fs afero.Fs, path string, n int, pool *gpool.Pool) {
	defer pool.Done()
	b, _ := afero.ReadFile(fs, path)
	var ki PowerAiAsset
	_ = json.Unmarshal(b,  &ki)
	fuckerList[n] = ki.Asset.Path
	logrus.WithField("id", n).Info(ki.Asset.Name)
}

func main() {
	//for i:=0; i<500000; i++ {
	//	f, _:=os.Open("../fucker/a.json")
	//	defer f.Close()
	//	tar, _ := os.Create("../fucker/" + strconv.Itoa(i) + ".json")
	//	defer tar.Close()
	//	io.Copy(tar, f)
	//}
	start := time.Now()
	pool := gpool.New(150)
	fs := afero.NewOsFs()
	n := 0
	_ = afero.Walk(fs, "../fucker", func(path string, info os.FileInfo, err error) error{
		n++
		pool.Add(1)
		go read(fs, path,n, pool)
		//logrus.WithField("n", n).Info(info.Name())
		return nil
	})
	pool.Wait()
	use := time.Since(start)
	fmt.Println(use)
}
