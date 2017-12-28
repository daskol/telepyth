package srv

import (
	"os"
	"strconv"
	"time"
)

type LogRecord struct {
	Timestamp time.Time
	UserID    string
	EventName string
}

var logCh = make(chan LogRecord)

func RunLogger(filename string) error {
	mode := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	file, err := os.OpenFile(filename, mode, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	for record := range logCh {
		format := FormatRecord(&record)
		file.WriteString(format)
		file.Sync()
	}

	return nil
}

func FormatRecord(record *LogRecord) string {
	date := record.Timestamp.Format("2006-01-02T15:04:05Z-0700")
	timestamp := strconv.FormatInt(record.Timestamp.Unix(), 10)
	return timestamp + "\t" + date + "\t" +
		record.UserID + "\t" + record.EventName + "\n"
}

func EnqueueLogRecord(UserID int, EventName string) {
	logCh <- LogRecord{time.Now(), strconv.Itoa(UserID), EventName}
}
