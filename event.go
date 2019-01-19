package event

import (
	"sync"
)

var (
	EventAll = "*"

	DefaultEventMgr *EventMgr = New(nil)
	instanceMap     map[interface{}]*EventMgr
	instanceMutex   = sync.Mutex{}
)

type EventHandler struct {
	once    bool
	handler func(evt interface{}, args ...interface{})
}

type EventMgr struct {
	listenerMap map[interface{}]interface{}
	listeners   map[interface{}]map[interface{}]EventHandler
	sync.Mutex
	valid bool
}

func eventHandler(handler func(evt interface{}, args ...interface{}), event interface{}, args ...interface{}) {
	defer handlePanic()
	handler(event, args...)
}

func (eventMgr *EventMgr) Subscrib(tag interface{}, event interface{}, handler func(evt interface{}, args ...interface{})) bool {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if _, ok := eventMgr.listenerMap[tag]; ok {
		logDebug("NewListener Error: listener %v exist!", tag)
		return false
	}

	eventMgr.listenerMap[tag] = event
	if eventMgr.listeners[event] == nil {
		eventMgr.listeners[event] = make(map[interface{}]EventHandler)
	}
	eventMgr.listeners[event][tag] = EventHandler{false, handler}

	return true
}

func (eventMgr *EventMgr) SubscribOnce(tag interface{}, event interface{}, handler func(evt interface{}, args ...interface{})) bool {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if _, ok := eventMgr.listenerMap[tag]; ok {
		logDebug("NewListener Error: listener %v exist!", tag)
		return false
	}

	eventMgr.listenerMap[tag] = event
	if eventMgr.listeners[event] == nil {
		eventMgr.listeners[event] = make(map[interface{}]EventHandler)
	}
	eventMgr.listeners[event][tag] = EventHandler{true, handler}

	return true
}

func (eventMgr *EventMgr) UnsubscribWithoutLock(tag interface{}) {
	if event, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[event], tag)
		if len(eventMgr.listeners[event]) == 0 {
			delete(eventMgr.listeners, event)
		}
	}
}

func (eventMgr *EventMgr) Unsubscrib(tag interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	if event, ok := eventMgr.listenerMap[tag]; ok {
		delete(eventMgr.listenerMap, tag)
		delete(eventMgr.listeners[event], tag)
		if len(eventMgr.listeners[event]) == 0 {
			delete(eventMgr.listeners, event)
		}
	}
}

func (eventMgr *EventMgr) Publish(event interface{}, args ...interface{}) {
	eventMgr.Lock()
	defer eventMgr.Unlock()

	/*
		els := eventMgr.listeners[event]
		aels := eventMgr.listeners[EventAll]
		all := make([]EventHandler, len(els)+len(aels))
		i := 0
		for _, l := range els {
			all[i] = l
			i++
		}
		for _, l := range aels {
			all[i] = l
			i++
		}

		eventMgr.Unlock()

		for _, l := range all {
			eventHandler(l, event, args)
		}
	*/
	if listeners, ok := eventMgr.listeners[event]; ok {
		for tag, listener := range listeners {
			eventHandler(listener.handler, event, args...)
			if listener.once {
				delete(eventMgr.listenerMap, tag)
				delete(listeners, tag)
				if len(listeners) == 0 {
					delete(eventMgr.listeners, event)
				}
			}
		}
	}
	if listeners, ok := eventMgr.listeners[EventAll]; ok {
		for tag, listener := range listeners {
			eventHandler(listener.handler, event, args...)
			if listener.once {
				delete(listeners, tag)
			}
		}
	}
}

func Subscrib(tag interface{}, event interface{}, handler func(evt interface{}, args ...interface{})) bool {
	return DefaultEventMgr.Subscrib(tag, event, handler)
}

func SubscribOnce(tag interface{}, event interface{}, handler func(evt interface{}, args ...interface{})) bool {
	return DefaultEventMgr.SubscribOnce(tag, event, handler)
}

func UnsubscribWithoutLock(tag interface{}) {
	DefaultEventMgr.UnsubscribWithoutLock(tag)
}

func Unsubscrib(tag interface{}) {
	DefaultEventMgr.Unsubscrib(tag)
}

func Publish(event interface{}, args ...interface{}) {
	DefaultEventMgr.Publish(event, args...)
}

func New(tag interface{}) *EventMgr {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if tag != nil {
		if _, ok := instanceMap[tag]; ok {
			logDebug("NewEventMgr Error: EventMgr %v exist!", tag)
			return nil
		}
	}

	eventMgr := &EventMgr{
		listenerMap: make(map[interface{}]interface{}),
		listeners:   make(map[interface{}]map[interface{}]EventHandler),
		//mutex:       sync.Mutex{},
		valid: true,
	}

	if tag != nil {
		if instanceMap == nil {
			instanceMap = make(map[interface{}]*EventMgr)
		}
		instanceMap[tag] = eventMgr
	}

	return eventMgr
}

func DeleteInstance(tag interface{}) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if eventMgr, ok := instanceMap[tag]; ok {
		eventMgr.Lock()
		defer eventMgr.Unlock()

		for k, e := range eventMgr.listenerMap {
			if emap, ok := eventMgr.listeners[e]; ok {
				for kk, _ := range emap {
					delete(emap, kk)
				}
			}
			delete(eventMgr.listeners, k)
		}
		delete(instanceMap, tag)
	}
}

func GetInstanceByTag(tag interface{}) (*EventMgr, bool) {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if eventMgr, ok := instanceMap[tag]; ok {
		return eventMgr, true
	}
	return nil, false
}
