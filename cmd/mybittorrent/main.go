package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

type Torrent struct {
	Announce  string `json:"announce"`
	CreatedBy string `json:"created by"`
	Info      Info   `json:"info"`
}

type Info struct {
	Length      int    `json:"length"`
	Name        string `json:"name"`
	PieceLength int    `json:"piece length"`
	Pieces      string `json:"pieces"`
}

func parseInt(input string) (int, int, error) {

	value := ""
	if len(input) == 0 {
		return -1, -1, errors.New("Cannot parse empty string.")
	}
	current := 1
	for {
		if input[current] == 'e' {
			break
		}
		value = value + string(input[current])
		current++
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return -1, -1, err
	}
	return number, current + 1, nil
}
func parseString(input string) (string, int, error) {

	var firstColonIndex int

	if len(input) == 0 {
		return "", -1, errors.New("Cannot parse empty string.")
	}
	for current := 0; current < len(input); current++ {
		if input[current] == ':' {
			firstColonIndex = current
			break
		}
	}

	lengthStr := input[:firstColonIndex]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", -1, err
	}

	return input[firstColonIndex+1 : firstColonIndex+1+length], firstColonIndex + 1 + length, nil
}

func parseList(bencodedString string) (interface{}, int, error) {
	currentIndex := 1
	list := []interface{}{}
	for bencodedString[currentIndex] != 'e' {
		value, index, err := parseOne(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		list = append(list, value)
		currentIndex = index + currentIndex
	}
	return list, currentIndex, nil
}

func parseDictionary(bencodedString string) (interface{}, int, error) {
	currentIndex := 1
	dict := map[string]interface{}{}
	for bencodedString[currentIndex] != 'e' {
		key, index, err := parseOne(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		value, index, err := parseOne(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		dict[key.(string)] = value
	}
	return dict, currentIndex + 1, nil
}

func parseOne(bencodedString string) (interface{}, int, error) {

	letter := bencodedString[0]
	if letter == 'l' {
		if result, nextPos, err := parseList(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if letter == 'd' {
		if result, nextPos, err := parseDictionary(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if letter == 'i' {
		if result, nextPos, err := parseInt(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if unicode.IsDigit(rune(letter)) {
		if result, nextPos, err := parseString(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	}
	return nil, -1, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, _, err := parseOne(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		fileName := os.Args[2]
		byteContent, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		decoded, _, err := parseOne(string(byteContent))
		jsonOutput, _ := json.Marshal(decoded)
		var torrent Torrent
		json.Unmarshal(jsonOutput, &torrent)
		fmt.Println("Tracker URL:", torrent.Announce)
		fmt.Println("Length:", torrent.Info.Length)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
