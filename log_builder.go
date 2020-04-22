package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type jsonTime struct {
	time.Time
}

type jsonLogLine struct {
	Message          string            `json:"message"`
	ContainerId      string            `json:"container_id"`
	ContainerName    string            `json:"container_name"`
	ContainerCreated jsonTime          `json:"container_created"`
	ImageId          string            `json:"image_id"`
	ImageName        string            `json:"image_name"`
	Command          string            `json:"command"`
	Tag              string            `json:"tag"`
	Extra            map[string]string `json:"extra"`
	Host             string            `json:"host"`
	Timestamp        jsonTime          `json:"time"`
}

func logMessage(lp *logPair, message []byte) error {
	lp.logLine.Message = string(message[:])
	lp.logLine.Timestamp = jsonTime{time.Now()}

	bytes, err := json.Marshal(lp.logLine)
	if err != nil {
		return err
	}

	sendUDP(lp, bytes)
	return nil
}

func (t jsonTime) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf("\"%s\"", t.Format(time.RFC3339Nano))
	return []byte(str), nil
}

func sendUDP(lp *logPair, message []byte) {
	conn, err := net.Dial("udp", lp.dst)
	if err != nil {
		logrus.WithField("id", lp.info.ContainerID).Error(fmt.Sprintf("Error opening UDP connection %s", err))
		return
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, string(message))
	if err != nil {
		logrus.WithField("id", lp.info.ContainerID).Error(fmt.Sprintf("Error sending: %s", err))
		return
	}
}
