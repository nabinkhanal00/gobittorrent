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

func decodeBencode(bencodedString string) (interface{}, error) {

	if bencodedString[0] == 'l' {
		bencodedString = bencodedString[1 : len(bencodedString)-1]
		currentIndex := 0
		result := []interface{}{}
		s := []interface{}{}
		var oldTop *[]interface{} = nil
		current := &result
		s = append(s, &result)
		for currentIndex < len(bencodedString) {
			if bencodedString[currentIndex] == 'l' {
				newList := []interface{}{}
				s = append(s, &newList)
				oldTop = current
				current = &newList
				currentIndex++
			} else if bencodedString[currentIndex] == 'e' {
				*oldTop = append(*oldTop, *current)
				current = oldTop
				s = s[:len(s)-1]
				if len(s) == 1 {
					oldTop = nil
				} else {
					oldTop = s[len(s)-2].(*[]interface{})
				}

				currentIndex++
			} else if bencodedString[currentIndex] == 'i' {
				intValue, index, err := parseInt(bencodedString[currentIndex:])
				if err != nil {
					return nil, err
				}
				currentIndex = currentIndex + index
				*current = append(*current, intValue)
			} else if unicode.IsDigit(rune(bencodedString[currentIndex])) {
				strValue, index, err := parseString(bencodedString[currentIndex:])
				if err != nil {
					return nil, err
				}
				currentIndex = currentIndex + index
				*current = append(*current, strValue)
			} else {
			}
		}
		return result, nil
	} else if bencodedString[0] == 'd' {
		// elements := map[string]interface{}{}
		// bencodedString = bencodedString[1 : len(bencodedString)-1]
	} else if bencodedString[0] == 'i' {
		intValue, _, err := parseInt(bencodedString)
		if err != nil {
			return nil, err
		}
		return intValue, nil
	} else if unicode.IsDigit(rune(bencodedString[0])) {
		strValue, _, err := parseString(bencodedString)
		if err != nil {
			return nil, err
		}
		return strValue, nil

	} else {
		return nil, errors.New("Unsupported")
	}
	return 3, nil
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
