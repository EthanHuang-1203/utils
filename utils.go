package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Config *viper.Viper
var Log *zap.Logger
var logrotate lumberjack.Logger
var atomicLevel = zap.NewAtomicLevel()

func init() {
	content := `logrotate:
  filename: ./Log/test.log 
  maxsize: 200 # MB           
  maxbackups : 10 # 最多保留幾個檔案
  maxage: 30 # 最多保存天數
  compress: false # 是否壓縮
log:
  level: -1 # -1~4 Debug > Info > Warn > Error > DPanic > Panic`

	if _, err := os.Stat("config.yaml"); err != nil {
		fmt.Printf("File does not exist\n")
		var (
			fileName = "./config.yaml"
			err2     error
		)
		if err2 = ioutil.WriteFile(fileName, []byte(content), 0666); err2 != nil {
			fmt.Println("Writefile Error =", err2)
			return
		}
	}

	Config = viper.New()
	Config.SetConfigName("config.yaml") // name of config file (without extension)
	Config.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
	Config.AddConfigPath(".")           // optionally look for config in the working directory
	err := Config.ReadInConfig()        // Find and read the config file
	if err != nil {                     // Handle errors reading the config file

		panic(fmt.Errorf("Fatal error config file: %w \nFile content include:\n%s", err, content))
	}
	/*
		err = Config.Unmarshal(&config)
		if err != nil {
			fmt.Printf("unable to decode into struct, %v", err)
		}
	*/

	logrotate = lumberjack.Logger{
		Filename:   Config.GetString("logrotate.filename"), // 日誌文件路徑
		MaxSize:    Config.GetInt("logrotate.maxsize"),     // 記錄文件保存的最大尺寸單位：M
		MaxBackups: Config.GetInt("logrotate.maxsize"),     // 日誌文件最多保存多少個備份
		MaxAge:     Config.GetInt("logrotate.maxsize"),     // 最多保存多少天的文件文件最多保存多少天
		Compress:   Config.GetBool("logrotate.compress"),   // 是否壓縮
	}

	atomicLevel := zap.NewAtomicLevel()
	// 設置日誌級別,可動態修改
	changeLogLevel(&atomicLevel)

	//atomicLevel.SetLevel(zapcore.InfoLevel)
	// public EncoderConfig
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,                            // Level大小寫 "INFO" "info, CapitalLevelEncoder || LowercaseLevelEncoder
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // 自定義時間格式
		EncodeDuration: zapcore.SecondsDurationEncoder,                         //
		EncodeCaller:   zapcore.ShortCallerEncoder,                             // caller路徑完整或縮寫,FullCallerEncoder || ShortCallerEncoder
		EncodeName:     zapcore.FullNameEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),                                             // console || json, NewJSONEncoder(),NewConsoleEncoder()
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&logrotate)), // >> console and file
		atomicLevel, // defualt log level
	)

	Log = zap.New(core) //zap.AddCaller() ,zap.AddStacktrace(),zap.Fields(zap.String("appName", name)),...
	defer Log.Sync()

	Config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		changeLogLevel(&atomicLevel)
		Test()
	})
	Config.WatchConfig()

}

func Test() {
	Log.Info("Info")
	Log.Debug("Debug")
	Log.Warn("Warn")
	Log.Error("Error")
}

func changeLogLevel(atomicLevel *zap.AtomicLevel) {
	switch x := Config.GetInt("log.level"); x {
	case -1:
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case 0:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case 1:
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case 2:
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	case 3:
		atomicLevel.SetLevel(zapcore.DPanicLevel)
	case 4:
		atomicLevel.SetLevel(zapcore.PanicLevel)
	default:
		atomicLevel.SetLevel(zapcore.DebugLevel)

	}
}
