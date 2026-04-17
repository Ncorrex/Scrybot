package server

import (
	"encoding/json"
	"io"
	"strings"
	"time"
)

// LogLine is the JSON payload pushed to WebSocket clients for each log event.
type LogLine struct {
	TS    time.Time `json:"ts"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

// LogWriter implements io.Writer. It forwards every write to origin and
// broadcasts a JSON LogLine to all connected WebSocket clients.
type LogWriter struct {
	hub    *Hub
	origin io.Writer
}

func NewLogWriter(hub *Hub, origin io.Writer) *LogWriter {
	return &LogWriter{hub: hub, origin: origin}
}

func (w *LogWriter) Write(p []byte) (int, error) {
	n, err := w.origin.Write(p)

	msg := strings.TrimRight(string(p), "\n\r")
	if msg == "" {
		return n, err
	}

	upper := strings.ToUpper(msg)
	level := "INFO"
	switch {
	case strings.Contains(upper, "ERROR"), strings.Contains(upper, "FATAL"):
		level = "ERROR"
	case strings.Contains(upper, "WARN"):
		level = "WARN"
	}

	line := LogLine{TS: time.Now(), Level: level, Msg: msg}
	data, _ := json.Marshal(line)
	w.hub.Broadcast(data)

	return n, err
}
