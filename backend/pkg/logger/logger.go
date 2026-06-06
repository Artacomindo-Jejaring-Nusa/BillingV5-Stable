package logger

import (
	"fmt"

	"os"
	"time"

	"gorm.io/gorm"
)

// LogLevel represents the severity of the log
type LogLevel string

const (
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// SystemLog matches the database structure for GORM
type SystemLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	Timestamp time.Time `gorm:"type:datetime(6);default:CURRENT_TIMESTAMP(6);index"`
	Level     string    `gorm:"type:varchar(50);not null"`
	Message   string    `gorm:"type:text;not null"`
}

func (SystemLog) TableName() string {
	return "system_logs"
}

// Logger is a structured logger that writes to both console (with color) and the system_logs DB table.
type Logger struct {
	db *gorm.DB
}

var globalLogger *Logger

// Init initializes the structured global logger
func Init(db *gorm.DB) {
	globalLogger = &Logger{
		db: db,
	}
}

// PrintBanner prints the ASCII header at server startup in a nice Cyan color
func PrintBanner() {
	banner := `
$$$$$$$\  $$$$$$\ $$\       $$\       $$$$$$\ $$\   $$\  $$$$$$\                                   
$$  __$$\ \_$$  _|$$ |      $$ |      \_$$  _|$$$\  $$ |$$  __$$\                                  
$$ |  $$ |  $$ |  $$ |      $$ |        $$ |  $$$$\ $$ |$$ /  \__|                                 
$$$$$$$\ |  $$ |  $$ |      $$ |        $$ |  $$ $$\$$ |$$ |$$$$\                                  
$$  __$$\   $$ |  $$ |      $$ |        $$ |  $$ \$$$$ |$$ |\_$$ |                                 
$$ |  $$ |  $$ |  $$ |      $$ |        $$ |  $$ |\$$$ |$$ |  $$ |                                 
$$$$$$$  |$$$$$$\ $$$$$$$$\ $$$$$$$$\ $$$$$$\ $$ | \$$ |\$$$$$$  |                                 
\_______/ \______|\________|\________|\______|\__|  \__| \______/                                  
                                                                                                   
                                                                                                   
                                                                                                   
 $$$$$$\  $$$$$$$\ $$$$$$$$\  $$$$$$\   $$$$$$\   $$$$$$\  $$\      $$\       $$\    $$\ $$$$$$$\  
$$  __$$\ $$  __$$\\__$$  __|$$  __$$\ $$  __$$\ $$  __$$\ $$$\    $$$ |      $$ |   $$ |$$  ____| 
$$ /  $$ |$$ |  $$ |  $$ |   $$ /  $$ |$$ /  \__|$$ /  $$ |$$$$\  $$$$ |      $$ |   $$ |$$ |      
$$$$$$$$ |$$$$$$$  |  $$ |   $$$$$$$$ |$$ |      $$ |  $$ |$$\$$\$$ $$ |      \$$\  $$  |$$$$$$$\  
$$  __$$ |$$  __$$<   $$ |   $$  __$$ |$$ |      $$ |  $$ |$$ \$$$  $$ |       \$$\$$  / \_____$$\ 
$$ |  $$ |$$ |  $$ |  $$ |   $$ |  $$ |$$ |  $$\ $$ |  $$ |$$ |\$  /$$ |        \$$$  /  $$\   $$ |
$$ |  $$ |$$ |  $$ |  $$ |   $$ |  $$ |\$$$$$$  | $$$$$$  |$$ | \_/ $$ |         \$  /$$\\$$$$$$  |
\__|  \__|\__|  \__|  \__|   \__|  \__| \______/  \______/ \__|     \__|          \_/ \__|\______/ 
`
	fmt.Printf("\033[36m%s\033[0m\n", banner)
}

func (l *Logger) print(level LogLevel, format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	var colorCode string
	switch level {
	case LevelInfo:
		colorCode = "\033[32m" // Green
	case LevelWarn:
		colorCode = "\033[33m" // Yellow
	case LevelError:
		colorCode = "\033[31m" // Red
	default:
		colorCode = "\033[0m"
	}

	fmt.Printf("%s%s [%-5s] %s\033[0m\n", colorCode, timestamp, level, msg)
	return msg
}

func (l *Logger) logToDB(level LogLevel, message string) {
	if l.db == nil {
		return
	}
	// Run in background to prevent blocking HTTP handler or business logic
	go func() {
		sysLog := &SystemLog{
			Timestamp: time.Now(),
			Level:     string(level),
			Message:   message,
		}
		if err := l.db.Create(sysLog).Error; err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m[LOGGER ERROR] Failed to write system log to DB: %v\033[0m\n", err)
		}
	}()
}

// Info logs messages with LevelInfo (Green)
func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		msg := globalLogger.print(LevelInfo, format, args...)
		globalLogger.logToDB(LevelInfo, msg)
	} else {
		// Fallback
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		fmt.Printf("%s [%-5s] %s\n", timestamp, LevelInfo, fmt.Sprintf(format, args...))
	}
}

// Warn logs messages with LevelWarn (Yellow)
func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		msg := globalLogger.print(LevelWarn, format, args...)
		globalLogger.logToDB(LevelWarn, msg)
	} else {
		// Fallback
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		fmt.Printf("%s [%-5s] %s\n", timestamp, LevelWarn, fmt.Sprintf(format, args...))
	}
}

// Error logs messages with LevelError (Red)
func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		msg := globalLogger.print(LevelError, format, args...)
		globalLogger.logToDB(LevelError, msg)
	} else {
		// Fallback
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		fmt.Printf("%s [%-5s] %s\n", timestamp, LevelError, fmt.Sprintf(format, args...))
	}
}
