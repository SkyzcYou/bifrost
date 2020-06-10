package bifrost

import (
	"flag"
	"fmt"
	"github.com/apsdehal/go-logger"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	confPath = flag.String("f", "./configs/bifrost.yml", "the bifrost `config`uration file path.")
	help     = flag.Bool("h", false, "this `help`")
	//confBackupDelay = flag.Duration("b", 10, "how many minutes `delay` for backup nginx config")
	Configs  *ParserConfigs
	dbConfig DBConfig
	myLogger *logger.Logger
	Logf     *os.File
)

const (
	CRITICAL = logger.CriticalLevel
	ERROR    = logger.ErrorLevel
	WARN     = logger.WarningLevel
	NOTICE   = logger.NoticeLevel
	INFO     = logger.InfoLevel
	DEBUG    = logger.DebugLevel
)

type NGConfig struct {
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

type LogConfig struct {
	LogDir string          `yaml:"logDir"`
	Level  logger.LogLevel `yaml:"level"`
}

type ParserConfigs struct {
	//NGConfigs []NGConfig `json:"NGConfigs"`
	//NGConfigs  []NGConfig `yaml:"NGConfigs"`
	NGConfigs []NGConfig `yaml:"NGConfigs"`
	DBConfig  `yaml:"DBConfig"`
	LogConfig `yaml:"logConfig"`
}

func init() {
	// 初始化工作目录
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	workspace := filepath.Dir(ex)
	cdErr := os.Chdir(workspace)
	if cdErr != nil {
		panic(cdErr)
	}

	// 初始化应用传参
	flag.Parse()
	if *confPath == "" {
		*confPath = "./configs/bifrost.yml"
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	//confPath := "./configs/ng-conf-info.json"
	//confPath := "./configs/bifrost.yml"
	isExistConfig, pathErr := PathExists(*confPath)
	//isExistConfig, pathErr := PathExists(confPath)
	if !isExistConfig {
		if pathErr != nil {
			fmt.Println("The ngConfig file", "'"+*confPath+"'", "is not found.")
		} else {
			fmt.Println("Unkown error of the ngConfig file.")
		}
		flag.Usage()
		os.Exit(1)
	}

	// 初始化bifrost配置
	confData, readErr := readFile(*confPath)
	//confData, readErr := readFile(confPath)
	if readErr != nil {
		fmt.Println(readErr)
		flag.Usage()
		os.Exit(1)
	}

	Configs = &ParserConfigs{}
	//jsonErr := json.Unmarshal(confData, configs)
	jsonErr := yaml.Unmarshal(confData, Configs)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		flag.Usage()
		os.Exit(1)
	}

	// 初始化数据库信息
	dbConfig = Configs.DBConfig

	// 初始化日志
	logDir, absErr := filepath.Abs(Configs.LogDir)
	if absErr != nil {
		panic(absErr)
	}
	logPath := filepath.Join(logDir, "bifrost.out")
	logf, openErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if openErr != nil {
		panic(openErr)
	}

	myLogger, err = logger.New("Bifrost", Configs.Level, logf)
	if err != nil {
		panic(err)
	}
	myLogger.SetFormat("[%{module}] %{time:2006-01-02 15:04:05.000} [%{level}] %{message}\n")
}
