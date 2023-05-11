package main

import (
	"encoding/json"
	"log"
	"pub_sub/proto"
	"pub_sub/subscriber"
)

func contentFunc(v *proto.Message) {
	log.Println("debug:", v)
}

func main() {
	PathClient := subscriber.NewClient(contentFunc)
	PathClient.Subscribe("localhost:29898")

	paths := []string{"/sata"}
	data, _ := json.Marshal(map[string]interface{}{"paths": paths, "ftypes": []string{"pic"}})
	log.Println("debug:", string(data))

	PathClient.Topics <- &proto.Topic{Topic: "file", Rule: string(data)}

	select {}
}
