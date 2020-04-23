package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"path"
	"sync"
	"syscall"

	"github.com/docker/docker/api/types/plugins/logdriver"
	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/daemon/logger/loggerutils"
	protoio "github.com/gogo/protobuf/io"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tonistiigi/fifo"
)

type Driver struct {
	mu   sync.Mutex
	logs map[string]*logPair
}

type logPair struct {
	active  bool
	file    string
	info    logger.Info
	logLine jsonLogLine
	stream  io.ReadCloser
	dst     string
}

func NewDriver() *Driver {
	return &Driver{
		logs: make(map[string]*logPair),
	}
}

func (d *Driver) StartLogging(file string, logCtx logger.Info) error {
	d.mu.Lock()
	if _, exists := d.logs[path.Base(file)]; exists {
		d.mu.Unlock()
		return fmt.Errorf("logger for %q already exists", file)
	}
	d.mu.Unlock()

	logrus.WithField("id", logCtx.ContainerID).WithField("file", file).Info("Start logging")
	stream, err := fifo.OpenFifo(context.Background(), file, syscall.O_RDONLY, 0700)
	if err != nil {
		return errors.Wrapf(err, "error opening logger fifo: %q", file)
	}

	tag, err := loggerutils.ParseLogTag(logCtx, loggerutils.DefaultTemplate)
	extra, err := logCtx.ExtraAttributes(nil)
	hostname, err := logCtx.Hostname()
	if err != nil {
		return err
	}

	dstPort := readWithDefault(logCtx.Config, "port", "9017")
	dstHost := readWithDefault(logCtx.Config, "host", "127.0.0.1")
	dst := fmt.Sprintf("%s:%s", dstHost, dstPort)

	logLine := jsonLogLine{
		ContainerId:      logCtx.FullID(),
		ContainerName:    logCtx.Name(),
		ContainerCreated: jsonTime{logCtx.ContainerCreated},
		ImageId:          logCtx.ImageFullID(),
		ImageName:        logCtx.ImageName(),
		Command:          logCtx.Command(),
		Tag:              tag,
		Extra:            extra,
		Host:             hostname,
	}

	lp := &logPair{true, file, logCtx, logLine, stream, dst}

	d.mu.Lock()
	d.logs[path.Base(file)] = lp
	d.mu.Unlock()

	go consumeLog(lp)
	return nil
}

func (d *Driver) StopLogging(file string) error {
	logrus.WithField("file", file).Info("Stop logging")
	d.mu.Lock()
	lp, ok := d.logs[path.Base(file)]
	if ok {
		lp.active = false
		delete(d.logs, path.Base(file))
	} else {
		logrus.WithField("file", file).Errorf("Failed to stop logging. File %q is not active", file)
	}
	d.mu.Unlock()
	return nil
}

func shutdownLogPair(lp *logPair) {
	if lp.stream != nil {
		lp.stream.Close()
	}

	lp.active = false
}

func consumeLog(lp *logPair) {
	var buf logdriver.LogEntry

	dec := protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
	defer dec.Close()
	defer shutdownLogPair(lp)

	for {
		if !lp.active {
			logrus.WithField("id", lp.info.ContainerID).Debug("shutting down logger goroutine due to stop request")
			return
		}

		err := dec.ReadMsg(&buf)
		if err != nil {
			if err == io.EOF {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Debug("shutting down logger goroutine due to file EOF")
				return
			} else {
				logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error reading from FIFO, trying to continue")
				dec = protoio.NewUint32DelimitedReader(lp.stream, binary.BigEndian, 1e6)
				continue
			}
		}

		err = logMessage(lp, buf.Line)
		if err != nil {
			logrus.WithField("id", lp.info.ContainerID).WithError(err).Warn("error logging message, dropping it and continuing")
		}

		buf.Reset()
	}
}

func readWithDefault(m map[string]string, key string, def string) string {
	value, ok := m[key]
	if ok {
		return value
	}

	return def
}
