package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Port     int
	NodeId   int
	Peer_ips []string
	Peer_ids []int
}

func GetConfig() *Config {
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		log.Println("No config path arg")
		return nil
	}

	jsonFile, err := os.Open(argsWithoutProg[0])
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// we initialize our Users array
	var config Config

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if config.Port == 0 {
		log.Println("Port not specified")
		return nil
	}

	if len(config.Peer_ips) != len(config.Peer_ids) {
		log.Println("Peers must be same length")
		return nil
	}

	fmt.Printf("My port: %d\n", config.Port)
	fmt.Printf("My Node Id: %d\n", config.NodeId)

	for i := range config.Peer_ids {
		fmt.Println("Peer Id: " + config.Peer_ips[i])
		fmt.Printf("Peer Ip: %d\n", config.Peer_ids[i])
		fmt.Println("")
	}

	return &config
}
