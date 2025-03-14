package logger

import (
	"fmt"
	"github.com/novikoff-vvs/logger/helpers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

type Interface interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger(logFilePath, logName string, setStdoutOutput bool) (*ZapLogger, error) {
	// Убедимся, что директория существует
	if err := helpers.EnsureDir(logFilePath); err != nil {
		return nil, err
	}

	if setStdoutOutput {
		cf := zap.Config{
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
		logger, err := cf.Build()
		if err != nil {
			return nil, err
		}

		return &ZapLogger{
			logger: logger,
		}, nil
	}

	infoPath := filepath.Join(logFilePath, "info_"+logName+".log")
	errorPath := filepath.Join(logFilePath, "error_"+logName+".log")
	debugPath := filepath.Join(logFilePath, "debug_"+logName+".log")

	// Создаем WriteSyncer для информационных сообщений
	infoFile, err := os.OpenFile(infoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	infoSyncer := zapcore.AddSync(infoFile)

	// Создаем WriteSyncer для ошибок
	errorFile, err := os.OpenFile(errorPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	errorSyncer := zapcore.AddSync(errorFile)

	// Создаем WriteSyncer для дебага
	debugFile, err := os.OpenFile(debugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	debugSyncer := zapcore.AddSync(debugFile)

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zap.InfoLevel // Только Info
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zap.ErrorLevel // Ошибки и выше
	})

	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zap.DebugLevel // Только дебаг
	})

	// Создаем ядра для логирования
	infoCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		infoSyncer,
		infoLevel,
	)

	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		errorSyncer,
		errorLevel,
	)

	debugCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		debugSyncer,
		debugLevel,
	)

	// Объединяем ядра
	core := zapcore.NewTee(infoCore, errorCore, debugCore)

	// Создаем логгер
	logger := zap.New(core)

	return &ZapLogger{logger: logger}, nil
}

// Info logs an info message
func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

// Error logs an error message
func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

// Debug logs a debug message
func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

func (z *ZapLogger) Println(v ...interface{}) {
	z.logger.Debug(fmt.Sprintln(v...))
}

func (z *ZapLogger) Printf(format string, v ...interface{}) {
	z.logger.Debug(fmt.Sprintf(fmt.Sprintf(format, v...)))
}

func (z *ZapLogger) Sync() error {
	err := z.logger.Sync()
	if err != nil {
		return err
	}
	return nil
}
