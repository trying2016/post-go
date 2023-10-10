package utils

import (
	"fmt"
	"time"

	"github.com/trying2016/common-tools/log"
)

func NewWatch(info string) *Watch {
	w := &Watch{tick: time.Now().UnixNano() / 1e6, info: info}
	return w
}

type Watch struct {
	tick int64
	info string
}

func (w Watch) Info() string {
	return fmt.Sprintf("%s consume:%v", w.info, time.Now().UnixNano()/1e6-w.tick)
}

func (w Watch) Print() {
	log.Info("%s consume:%v", w.info, time.Now().UnixNano()/1e6-w.tick)
}
