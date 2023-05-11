package main

import (
	"encoding/json"
	"pub_sub/subscriber"
	"strings"
	"time"
)

var FilePublisher *subscriber.Publisher = subscriber.NewPubliser("file", 1000, NewFileTopic)

type FileTopic struct {
	Paths []string `json:"paths"`
	Ftype []string `json:"ftypes"`
}

type FileTopicValue struct {
	Path  string `json:"path"`
	Ftype string `json:"ftype"`
	IsDir int    `json:"isdir"`
}

func NewFileTopic() subscriber.Topic {
	return &FileTopic{}
}

func (topic *FileTopic) Update(data string) bool {
	json.Unmarshal([]byte(data), topic)
	return true
}

func (topic *FileTopic) Match(v interface{}) (r interface{}) {
	tv := v.(*FileTopicValue)
	ftypeMatched := false
	if tv.IsDir == 0 {
		if len(topic.Ftype) > 0 {
			for _, ftype := range topic.Ftype {
				if ftype == tv.Ftype {
					ftypeMatched = true
					break
				}
			}
		} else {
			ftypeMatched = true
		}
	} else {
		ftypeMatched = true
	}
	if !ftypeMatched {
		return nil
	}
	pathMatched := false
	for _, path := range topic.Paths {
		if path == tv.Path {
			pathMatched = true
			break
		}
		if ok := strings.HasPrefix(tv.Path, path); ok {
			pathMatched = true
			break
		}
	}
	if !pathMatched {
		return nil
	}
	if tv.IsDir == 1 {
		return tv.Path + "/"
	}
	return tv.Path
}

func main() {
	server := subscriber.NewServer(FilePublisher)
	go server.Run("localhost:29898")

	time.AfterFunc(time.Second*10, func() {
		path := "/sata/public"
		ftype := "pic"
		FilePublisher.Publish(&FileTopicValue{Path: path, IsDir: 1, Ftype: ftype})
	})

	select {}
}
