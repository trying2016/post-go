package utils

import (
	"reflect"
	"sync"
)

var singleNotificationCenter *NotificationCenter

func init() {
	singleNotificationCenter = &NotificationCenter{}
	singleNotificationCenter.mapObservers = make(map[string]*Observer)
}

func GetNotificationCenter() *NotificationCenter {
	return singleNotificationCenter
}

type observerCallback func(data interface{})

type Observer struct {
	callbackLock sync.RWMutex
	callback     []observerCallback
}

func (o *Observer) init() {
	o.callback = make([]observerCallback, 0)
}
func (o *Observer) add(fn func(data interface{})) {
	o.callbackLock.Lock()
	defer o.callbackLock.Unlock()
	o.callback = append(o.callback, fn)
}

// 移出一个callback，返回是否空
func (o *Observer) remove(fn observerCallback) bool {
	o.callbackLock.Lock()
	defer o.callbackLock.Unlock()
	for index, v := range o.callback {
		if reflect.ValueOf(v) == reflect.ValueOf(fn) {
			o.callback = append(o.callback[:index], o.callback[index+1:]...)
			return len(o.callback) == 0
		}
	}
	return len(o.callback) == 0
}

func (o *Observer) do(data interface{}) {
	o.callbackLock.RLock()
	defer o.callbackLock.RUnlock()
	for _, fn := range o.callback {
		fn(data)
	}
}

type NotificationCenter struct {
	observerLock sync.RWMutex
	mapObservers map[string]*Observer
}

func (nc *NotificationCenter) AddObserver(name string, callback func(data interface{})) {
	nc.observerLock.Lock()
	defer nc.observerLock.Unlock()
	observer := nc.getObserver(name)
	observer.add(callback)
}

func (nc *NotificationCenter) RemoveObserver(name string, callback func(data interface{})) {
	nc.observerLock.Lock()
	defer nc.observerLock.Unlock()
	if observer, ok := nc.mapObservers[name]; !ok {
		return
	} else {
		if observer.remove(callback) {
			delete(nc.mapObservers, name)
		}
	}
}

func (nc *NotificationCenter) PostNotification(name string, data interface{}) {
	nc.observerLock.RLock()
	defer nc.observerLock.RUnlock()
	if observer, ok := nc.mapObservers[name]; !ok {
		return
	} else {
		observer.do(data)
	}
}

func (nc *NotificationCenter) getObserver(name string) *Observer {
	observer, ok := nc.mapObservers[name]
	if !ok {
		observer = &Observer{}
		observer.init()
		nc.mapObservers[name] = observer
	}
	return observer
}
