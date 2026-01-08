package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// 定义日志分割器（lumberjack）
type LogFileConfig struct {
	Filename   string // 日志文件路径
	MaxSize    int    // 单个文件最大大小（MB）
	MaxBackups int    // 保留旧文件的最大数量
	MaxAge     int    // 保留旧文件的最大天数
	Compress   bool   // 是否压缩旧文件
}

// InitLogger 初始化zerolog，区分开发/生产环境
func InitLogger(isProd bool) {
	// 1. 设置时间格式（zerolog默认是Unix时间戳，改为可读格式）
	zerolog.TimeFieldFormat = time.DateTime
	zerolog.TimestampFieldName = "time" // 时间字段名
	zerolog.LevelFieldName = "level"    // 级别字段名
	zerolog.MessageFieldName = "msg"    // 消息字段名

	// 2. 配置审计日志和运行日志的文件输出
	auditWriter := newLogWriter(LogFileConfig{
		Filename:   filepath.Join("logs", "audit.log"),
		MaxSize:    100,  // 单个文件100MB
		MaxBackups: 30,   // 保留30个旧文件
		MaxAge:     7,    // 保留7天
		Compress:   true, // 压缩旧文件
	}, isProd)

	runWriter := newLogWriter(LogFileConfig{
		Filename:   filepath.Join("logs", "run.log"),
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     7,
		Compress:   true,
	}, isProd)

	// 3. 初始化全局审计日志和运行日志
	// AuditLogger: 记录HTTP请求
	AuditLogger = zerolog.New(auditWriter).With().Timestamp().Logger()
	// RunLogger: 记录业务逻辑，生产环境开启调用者信息
	RunLoggerCtx := zerolog.New(runWriter).With().Timestamp()
	if isProd {
		RunLoggerCtx = RunLoggerCtx.Caller()
	}
	RunLogger := RunLoggerCtx.Logger()
	// 替换zerolog的全局log（可选，便于业务中直接使用log.Info()）
	log.Logger = RunLogger

	// 4. 设置日志级别
	if isProd {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// newLogWriter 创建日志写入器：生产环境写文件，开发环境同时写控制台+文件
func newLogWriter(cfg LogFileConfig, isProd bool) io.Writer {
	// 初始化lumberjack日志分割器
	lumberjackWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
		LocalTime:  true, // 使用本地时间命名日志文件
	}

	// 开发环境：同时输出到控制台和文件
	if !isProd {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.DateTime}
		return io.MultiWriter(consoleWriter, lumberjackWriter)
	}
	// 生产环境：仅输出到文件
	return lumberjackWriter
}

// WithReqID 返回带reqid字段的运行日志logger
func WithReqID(reqid string) zerolog.Logger {
	return log.Logger.With().Str("reqid", reqid).Logger()
}

var (
	AuditLogger zerolog.Logger // 审计日志（HTTP请求）
	RunLogger   zerolog.Logger // 运行日志（业务逻辑，复用zerolog全局log）
)
