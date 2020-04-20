package main

// DONE:1.权限管理
// DOING:2.nginx配置定期备份机制
// TODO:3.日志规范化输出

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ClessLi/go-nginx-conf-parser/pkg/resolv"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	confPath = flag.String("f", "./configs/ng-conf-info.yml", "go-nginx-conf-parser ng-`conf`-info.y(a)ml path.")
	help     = flag.Bool("h", false, "this `help`")
	//confBackupDelay = flag.Duration("b", 10, "how many minutes `delay` for backup nginx config")
	dbConfig DBConfig
)

const (
	ERROR      = "ERROR"
	WARN       = "WARN"
	INFO       = "INFO"
	DEBUG      = "DEBUG"
	timeFormat = "2006-01-02 15:04:05"
)

type ParserConfig struct {
	//Name         string `json:"name"`
	//RelativePath string `json:"relative_path"`
	//Port         int    `json:"port"`
	//ConfPath     string `json:"conf_path"`
	Name         string `yaml:"name"`
	RelativePath string `yaml:"relativePath"`
	Port         int    `yaml:"port"`
	ConfPath     string `yaml:"confPath"`
	NginxBin     string `yaml:"nginxBin"`
}

type DBConfig struct {
	DBName   string `yaml:"DBName"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type ParserConfigs struct {
	//Configs []ParserConfig `json:"configs"`
	Configs  []ParserConfig `yaml:"configs"`
	DBConfig `yaml:"DBConfig"`
}

//func init() {
//	flag.Usage = usage
//}

//func usage() {
//	fmt.Fprintf(os.Stdout, `go-nginx-conf-parser version: v0.0.1`)
//	flag.Usage()
//}

func main() {
	flag.Parse()
	if *confPath == "" {
		*confPath = "./configs/ng-conf-info.yml"
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	//confPath := "./configs/ng-conf-info.json"
	//confPath := "./configs/ng-conf-info.yml"
	isExist, pathErr := PathExists(*confPath)
	//isExist, pathErr := PathExists(confPath)
	if !isExist {
		if pathErr != nil {
			fmt.Println("The config file", "'"+*confPath+"'", "is not found.")
		} else {
			fmt.Println("Unkown error of the config file.")
		}
		flag.Usage()
		os.Exit(1)
	}
	confData, readErr := readFile(*confPath)
	//confData, readErr := readFile(confPath)
	if readErr != nil {
		fmt.Println(readErr)
		flag.Usage()
		os.Exit(1)
	}

	configs := &ParserConfigs{}
	//jsonErr := json.Unmarshal(confData, configs)
	jsonErr := yaml.Unmarshal(confData, configs)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		flag.Usage()
		os.Exit(1)
	}

	dbConfig = configs.DBConfig

	for _, config := range configs.Configs {
		ng, err := resolv.Load(config.ConfPath)

		if err != nil {
			fmt.Println(err)
			continue
		}

		errChan := make(chan error)

		go run(&config, ng, errChan)

		err = <-errChan
		if err != nil {
			log(ERROR, fmt.Sprintf("%s's coroutine has been stoped. Cased by <%s>", config.Name, err))
		} else {
			log(INFO, fmt.Sprintf("%s's coroutine has been stoped", config.Name))
		}
	}

}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func run(appConfig *ParserConfig, ngConfig *resolv.Config, errChan chan error) {
	_, jerr := json.Marshal(ngConfig)
	//confBytes, jerr := json.Marshal(ngConfig)
	//confBytes, jerr := json.MarshalIndent(ngConfig, "", "    ")
	if jerr != nil {
		errChan <- jerr
	}

	loginURI := "/login/:username/:password"
	verifyURI := "/verify/:token"
	refreshURI := "/refresh/:token"
	apiURI := fmt.Sprintf("%s/:token", appConfig.RelativePath)

	ngBin, absErr := filepath.Abs(appConfig.NginxBin)
	if absErr != nil {
		errChan <- absErr
	}

	// 创建备份协程管道及启动备份协程
	bakChan := make(chan int)
	go Bak(appConfig, ngConfig, bakChan)

	router := gin.Default()
	// login
	router.GET(loginURI, login)
	// verify
	router.GET(verifyURI, verify)
	// refresh
	router.GET(refreshURI, refresh)
	// view
	router.GET(apiURI, func(c *gin.Context) {
		h := view(appConfig.Name, ngConfig, c)
		c.JSON(http.StatusOK, &h)
	})
	// update
	router.POST(apiURI, func(c *gin.Context) {
		h := update(appConfig.Name, ngBin, ngConfig, c)
		c.JSON(http.StatusOK, &h)
	})

	rErr := router.Run(fmt.Sprintf(":%d", appConfig.Port))
	if rErr != nil {
		// 关闭备份
		bakChan <- 9
		// 输出子任务运行错误
		errChan <- rErr
	}

	// 关闭备份
	bakChan <- 9
	errChan <- nil

}

func Bak(appConfig *ParserConfig, ngConfig *resolv.Config, c chan int) {
	for {
		select {
		case <-time.NewTicker(5 * time.Minute).C:
			bak(appConfig, ngConfig)
		case signal := <-c:
			if signal == 9 {
				log(INFO, fmt.Sprintf("[%s] Nginx Config backup is stop.", appConfig.Name))
				break
			}

		}
	}
}

func bak(appConfig *ParserConfig, ngConfig *resolv.Config) {
	bakDate := time.Now().Format("20060102")
	bakName := fmt.Sprintf("nginx.conf.%s.tgz", bakDate)

	bakPath, bErr := resolv.Backup(ngConfig, bakName)
	if bErr != nil && !os.IsExist(bErr) {
		message := fmt.Sprintf("[%s] Nginx Config backup to %s, but failed. <%s>", appConfig.Name, bakPath, bErr)
		log(ERROR, message)
		log(INFO, fmt.Sprintf("[%s] Nginx Config backup is stop.", appConfig.Name))
	} else if bErr == nil {
		log(INFO, fmt.Sprintf("[%s] Nginx Config backup to %s", appConfig.Name, bakPath))
	}

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, err
	} else {
		return false, nil
	}
}

func log(level, message string) {
	fmt.Printf("[%s] [%s] %s\n", level, time.Now().Format(timeFormat), message)

}
