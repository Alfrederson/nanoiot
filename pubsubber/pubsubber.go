package pubsubber

import (
	"container/list"
	"sync"
)

type List = list.List
type Node = list.Element
type Subscriber chan string

// essa GLOBAL armazena todas as subscrições.
var subscriptions map[string]*List = make(map[string]*List)
var subMutex sync.Mutex

func Publish(topic string, payload string) error {
	// envia payload para todos os listeners do topic.
	subMutex.Lock()
	defer subMutex.Unlock()
	l, exists := subscriptions[topic]
	if !exists {
		return nil
	}
	// envia o payload pra todos os subscribers e remove eles da lista.
	for element := l.Front(); element != nil; element = element.Next() {
		subscriber := element.Value.(Subscriber)
		subscriber <- payload
	}
	l.Init()
	return nil
}

func Subscribe(topic string, subscriber Subscriber) (*Node, error) {
	subMutex.Lock()
	defer subMutex.Unlock()
	l, exists := subscriptions[topic]
	if !exists {
		l = list.New()
		subscriptions[topic] = l
	}
	return l.PushBack(subscriber), nil
}
