package bittorrent

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"unicode"
)

func TorrentDecode(fileName string) (map[string]interface{}, error) {
	byteContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	decoded, _, err := Decode(string(byteContent))
	m, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, errors.New("Couldnot convert bendoded data to map")
	}
	return m, nil
}

func encodeInt(input int) (string, error) {
	value := strconv.Itoa(input)
	return "i" + value + "e", nil
}
func encodeString(input string) (string, error) {
	return fmt.Sprintf("%d:%v", len(input), input), nil
}
func encodeList(input []interface{}) (string, error) {
	result := "l"
	for _, value := range input {
		v, err := Encode(value)
		if err != nil {
			return "", err
		}
		result += v
	}
	result += "e"
	return result, nil
}
func encodeDictionary(input map[string]interface{}) (string, error) {
	keys := []string{}
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := "d"
	for _, key := range keys {
		encodedKey, err := Encode(key)
		if err != nil {
			return "", nil
		}
		result += encodedKey
		value := input[key]
		encodedValue, err := Encode(value)
		if err != nil {
			return "", nil
		}
		result += encodedValue
	}
	result += "e"
	return result, nil
}
func Encode(input interface{}) (string, error) {
	var result string
	switch v := input.(type) {
	case int:
		encodedValue, err := encodeInt(v)
		if err != nil {
			return "", err
		}
		result = encodedValue
	case string:
		encodedValue, err := encodeString(v)
		if err != nil {
			return "", err
		}
		result = encodedValue
	case []interface{}:
		encodedValue, err := encodeList(v)
		if err != nil {
			return "", err
		}
		result = encodedValue
	case map[string]interface{}:
		encodedValue, err := encodeDictionary(v)
		if err != nil {
			return "", err
		}
		result = encodedValue
	default:
		return "", errors.New("Unknown type.")
	}
	return result, nil
}

func decodeInt(input string) (int, int, error) {

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
func decodeString(input string) (string, int, error) {

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

func decodeList(bencodedString string) ([]interface{}, int, error) {
	currentIndex := 1
	list := []interface{}{}
	for bencodedString[currentIndex] != 'e' {
		value, index, err := Decode(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		list = append(list, value)
		currentIndex = index + currentIndex
	}
	return list, currentIndex + 1, nil
}

func decodeDictionary(bencodedString string) (map[string]interface{}, int, error) {
	currentIndex := 1
	dict := map[string]interface{}{}
	for bencodedString[currentIndex] != 'e' {
		key, index, err := Decode(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		value, index, err := Decode(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		dict[key.(string)] = value
	}
	return dict, currentIndex + 1, nil
}

func Decode(bencodedString string) (interface{}, int, error) {

	letter := bencodedString[0]
	if letter == 'l' {
		if result, nextPos, err := decodeList(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if letter == 'd' {
		if result, nextPos, err := decodeDictionary(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if letter == 'i' {
		if result, nextPos, err := decodeInt(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	} else if unicode.IsDigit(rune(letter)) {
		if result, nextPos, err := decodeString(bencodedString); err != nil {
			return nil, nextPos, err
		} else {
			return result, nextPos, nil
		}
	}
	return nil, -1, nil
}
