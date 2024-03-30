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
	currentIndex := 0
	result := []interface{}{}
	s := []interface{}{}
	var oldTop *[]interface{} = nil
	var current *[]interface{} = nil
	for currentIndex < len(bencodedString) {
		if bencodedString[currentIndex] == 'l' {
			newList := []interface{}{}
			s = append(s, &newList)
			if current != nil {
				oldTop = current
			}
			current = &newList
			currentIndex++
		} else if bencodedString[currentIndex] == 'e' {
			if oldTop != nil {
				*oldTop = append(*oldTop, *current)
				current = oldTop
				s = s[:len(s)-1]
				if len(s) == 1 {
					oldTop = nil
				} else {
					oldTop = s[len(s)-2].(*[]interface{})
				}
			} else {
				return *current, currentIndex + 1, nil
			}

			currentIndex++
		} else if bencodedString[currentIndex] == 'i' {
			intValue, index, err := parseInt(bencodedString[currentIndex:])
			if err != nil {
				return nil, -1, err
			}
			currentIndex = currentIndex + index
			*current = append(*current, intValue)
		} else if unicode.IsDigit(rune(bencodedString[currentIndex])) {
			strValue, index, err := parseString(bencodedString[currentIndex:])
			if err != nil {
				return nil, -1, err
			}
			currentIndex = currentIndex + index
			*current = append(*current, strValue)
		}
	}
	return result, currentIndex + 1, nil
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
func decodeBencode(bencodedString string) (interface{}, error) {
	letter := bencodedString[0]
	if letter == 'l' {
		if result, _, err := parseList(bencodedString); err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else if letter == 'd' {

		if result, _, err := parseDictionary(bencodedString); err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else if letter == 'i' {
		if result, _, err := parseInt(bencodedString); err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else if unicode.IsDigit(rune(letter)) {

		if result, _, err := parseString(bencodedString); err != nil {
			return nil, err
		} else {
			return result, nil
		}
	}
	return nil, nil
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
