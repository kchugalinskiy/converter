package main

import (
	"bufio"
	"flag"
	"os"

	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/pelletier/go-toml"
	"github.com/vharitonsky/iniflags"
)

var (
	source      = flag.String("i", "stdin", "Source - filepath or stdout")
	destination = flag.String("o", "stdout", "Destination - filepath or stdout")
)

func main() {
	iniflags.Parse()

	var input *os.File
	var output *os.File
	var err error

	if "stdin" == *source {
		input = os.Stdin
	} else {
		input, err = os.OpenFile(*source, os.O_RDONLY, 0444)
		if nil != err {
			log.Errorf("Cannot open file '%v' %v", *source, err)
			return
		}
	}
	inputStat, err := input.Stat()
	if nil != err {
		log.Errorf("Cannot stat file '%v' %v", *source, err)
		return
	}
	inputContents := make([]byte, inputStat.Size()*2)
	reader := bufio.NewReader(input)
	_, err = reader.Read(inputContents)
	if nil != err {
		log.Errorf("Cannot read contents of '%v' %v", *source, err)
		return
	}
	inputContentsString := string(inputContents)

	if "stdout" == *destination {
		output = os.Stdout
	} else {
		output, err = os.OpenFile(*destination, os.O_WRONLY|os.O_CREATE, 0644)
		if nil != err {
			log.Errorf("Cannot open file '%v' %v", *destination, err)
			return
		}
	}

	writer := bufio.NewWriter(output)

	tomlTree, err := toml.Load(inputContentsString)
	if err != nil {
		// handle error
		log.Errorf("error loading toml: %v", err)
		return
	}

	goTree := convertToGoTree(tomlTree)

	jsonResult, err := json.Marshal(goTree)

	if nil != err {
		log.Errorf("error %v parsing tree", err)
		return
	}

	_, err = writer.Write(jsonResult)
	if nil != err {
		log.Errorf("writing file: %v", err)
	}
	writer.Flush()
}

func convertToGoTree(tree *toml.TomlTree) map[string]interface{} {
	result := make(map[string]interface{})
	keys := tree.Keys()
	for _, key := range keys {
		value := tree.Get(key)
		switch value.(type) {
		case *toml.TomlTree:
			result[key] = convertToGoTree(value.(*toml.TomlTree))
		case []*toml.TomlTree:
			result[key] = make(map[string]interface{}, len(value.([]*toml.TomlTree)))
			for index, item := range value.([]*toml.TomlTree) {
				result[key].([]map[string]interface{})[index] = convertToGoTree(item)
			}
		default:
			result[key] = &value
		}
	}
	return result
}
