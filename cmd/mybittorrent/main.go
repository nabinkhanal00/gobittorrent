package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func decodeBencode(bencodedString string) (interface{}, string, error) {
	if bencodedString[0] == 'l' {
		var elements []interface{}
		remaining := bencodedString[1 : len(bencodedString)-1]
		if remaining == "" {
			return elements, "", nil
		}
		for {
			element, rem, err := decodeBencode(remaining)
			if err != nil {
				return elements, "", nil
			}
			remaining = rem
			elements = append(elements, element)
			if remaining == "" {
				break
			}
		}
		return elements, "", nil

	} else if bencodedString[0] == 'i' {
		value := ""
		var i int
		for i = 1; i < len(bencodedString); i++ {
			if bencodedString[i] == 'e' {
				break
			}
			value = value + string(bencodedString[i])
		}
		number, err := strconv.Atoi(value)
		if err != nil {
			return "", bencodedString[i+1:], err
		}
		return number, bencodedString[i+1:], nil
	} else if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", bencodedString[firstColonIndex+1+length:], err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], bencodedString[firstColonIndex+1+length:], nil
	} else {
		return "", "", fmt.Errorf("Only strings are supported at the moment")
	}
}

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
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
