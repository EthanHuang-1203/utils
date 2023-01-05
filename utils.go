package utils

import (
	"fmt"
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
	Config = viper.New()
	Config.SetConfigName("config") // name of config file (without extension)
	Config.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	Config.AddConfigPath(".")      // optionally look for config in the working directory
	err := Config.ReadInConfig()   // Find and read the config file
	if err != nil {                // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}
	/*
		err = Config.Unmarshal(&config)
		if err != nil {
			fmt.Printf("unable to decode into struct, %v", err)
		}
	*/

	Config.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		/*
			atomicLevel.SetLevel(zapcore.ErrorLevel)

			logrotate.MaxSize = 1
		*/
		fmt.Println(Config.GetString("test"))
	})
	Config.WatchConfig()

	logrotate = lumberjack.Logger{
		Filename:   "./log", // 日誌文件路徑
		MaxSize:    10,      // 記錄文件保存的最大尺寸單位：M
		MaxBackups: 5,       // 日誌文件最多保存多少個備份
		MaxAge:     28,      // 最多保存多少天的文件文件最多保存多少天
		Compress:   false,   // 是否壓縮
	}

	//atomicLevel := zap.NewAtomicLevel()
	// 設置日誌級別,可動態修改
	atomicLevel.SetLevel(zapcore.InfoLevel)
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
		zapcore.NewJSONEncoder(encoderConfig),                                                // console || json, NewJSONEncoder(),NewConsoleEncoder()
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&logrotate)), // >> console and file
		atomicLevel, // defualt log level
	)

	Log = zap.New(core, zap.AddCaller()) //zap.AddCaller() ,zap.AddStacktrace(),zap.Fields(zap.String("appName", name)),...
	defer Log.Sync()

}

func Test() {
	fmt.Printf("tttttttttttttt")
}
