package subscriber

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var wgPool = sync.Pool{New: func() interface{} { return new(sync.WaitGroup) }}

type Subscriber chan interface{}
type TopicFunc func() Topic

type Topic interface {
	Update(string) bool
	Match(v interface{}) interface{}
}

type Publisher struct {
	m           sync.RWMutex
	topic       string
	buffer      int
	timeout     time.Duration
	subscribers map[Subscriber]Topic
	topicFunc   TopicFunc
}

func NewPubliser(topic string, buffer int, topicFunc TopicFunc) *Publisher {
	return &Publisher{
		topic:       topic,
		buffer:      buffer,
		topicFunc:   topicFunc,
		subscribers: make(map[Subscriber]Topic),
	}
}

func (p *Publisher) Subscribe() Subscriber {
	ch := make(Subscriber, p.buffer)
	p.m.Lock()
	p.subscribers[ch] = p.topicFunc()
	p.m.Unlock()
	log.Println("add sub", ch, len(p.subscribers), p, fmt.Sprintf("%p", p))
	return ch
}

func (p *Publisher) Update(s Subscriber, t string) {
	p.m.Lock()
	if topic, ok := p.subscribers[s]; ok {
		topic.Update(t)
	}
	p.m.Unlock()
}

func (p *Publisher) Evict(sub Subscriber) {
	log.Println("info:", "evict sub", sub, p, fmt.Sprintf("%p", p))
	p.m.Lock()
	if _, ok := p.subscribers[sub]; ok {
		delete(p.subscribers, sub)
		close(sub)
	}
	p.m.Unlock()
}

func (p *Publisher) Publish(v interface{}) {
	p.m.RLock()
	if len(p.subscribers) == 0 {
		p.m.RUnlock()
		return
	}
	wg := wgPool.Get().(*sync.WaitGroup)
	for sub, topic := range p.subscribers {
		if r := topic.Match(v); r != nil {
			wg.Add(1)
			go p.sendTopic(sub, r, wg)
		}
	}
	wg.Wait()
	wgPool.Put(wg)
	p.m.RUnlock()
}

func (p *Publisher) sendTopic(sub Subscriber, v interface{}, wg *sync.WaitGroup) {
	log.Println("send topic", v)
	defer wg.Done()
	if p.timeout > 0 {
		timeout := time.NewTimer(p.timeout)
		defer timeout.Stop()

		select {
		case sub <- v:
		case <-timeout.C:
		}
		return
	}
	select {
	case sub <- v:
	default:
	}
}
