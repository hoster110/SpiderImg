package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
	"strings"
	"net/http"
	"regexp"
	"sync"
	"time"
)

type Config struct{
	Addr string
	FindInfoStr string
	SaveImg string
	ThreadNum int
}

var config Config

func ReadJson(path string){
	bytess , err := ioutil.ReadFile(path)
	if err != nil{
		fmt.Println(err)
		os.Exit(-1)
	}

	if err = json.Unmarshal(bytess,&config);err!=nil{
		fmt.Println(err)
		os.Exit(-1)
	}
}

func RemoveDuplicatesAndEmpty(a []string) (ret []string){
	a_len := len(a)
	for i:=0; i < a_len; i++{
		if (i > 0 && a[i-1] == a[i]) || len(a[i])==0{
			continue;
		}
		ret = append(ret, a[i])
	}
	return
}


var wg sync.WaitGroup
func main(){
	ReadJson("./SpiderBaiduImgOpt/config.json")
	DebugInfo(config)

	err := os.MkdirAll(config.SaveImg,os.ModePerm)
	if err != nil{
		DebugInfo(err)
	}

	config.Addr = strings.Replace(config.Addr,"{word}",config.FindInfoStr,-1)

	//http://img1.imgtn.bdimg.com/it/u=2492274450,283114429&fm=200&gp=0.jpg
	var j int
	for{
		temp := strings.Replace(config.Addr,"{pn}",fmt.Sprintf("%+v",j),-1)
		fmt.Println(temp)
		if resp ,err :=http.Get(temp);err ==nil{
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil{
				DebugInfo(err)
				continue
			}

			reg := regexp.MustCompile(`[a-z]+\:\/{2}(([a-z]|\d+|\.)*\/)*u=\d+,\d+&fm=\d+&gp=\d+.jpg`)
			param := reg.FindAllString(string(body),-1)

			//去重
			param = RemoveDuplicatesAndEmpty(param)

			if len(param) == 0{
				DebugInfo("Down pic End!, DownPage =",j)
				break
			}

			for t :=0 ;t<config.ThreadNum ;t++{
				wg.Add(1)
				go func(index int,imglist []string){
					savepath := fmt.Sprintf("%+v/%+v",config.SaveImg,index)
					os.MkdirAll(savepath,os.ModePerm)

					for i := 0;i < len(imglist);i++{
						if resp ,err :=http.Get(imglist[i]);err ==nil && (i%config.ThreadNum ==  index){
							body, err := ioutil.ReadAll(resp.Body)
							if err != nil || len(body)<=4096{
								DebugInfo(err,"|| pic <= 4k")
								continue
							}
							//保存图片t
							ioutil.WriteFile(fmt.Sprintf("%+v/%+vThread_%+vList_%+vIndex.jpg",savepath,index,j,i),body,666)
							DebugInfo(fmt.Sprintf("%+v/%+vThread_%+vList_%+vIndex.jpg",savepath,index,j,i))
						}
					}
					DebugInfo(index,"thread end!")
					wg.Done()

				}(t,param)
			}
			wg.Wait()
			DebugInfo("**==Down pic End!, DownPage =",j,"==**")
		}
		time.Sleep(time.Millisecond*100)
		j++
	}

}
