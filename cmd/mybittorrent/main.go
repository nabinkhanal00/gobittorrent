package main

import (
	// Uncomment this line to pass the first stage
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

func createURL(urlstring string, params map[string]string) string {
	if urlstring[len(urlstring)-1] != '/' {
		urlstring += "/"
	}
	urlstring += "?"
	for key, value := range params {
		urlstring += url.QueryEscape(key) + "=" + url.QueryEscape(value) + "&"

	}
	return urlstring

}
func torrentDecode(fileName string) (map[string]interface{}, error) {
	byteContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	decoded, _, err := decodeOne(string(byteContent))
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
		v, err := encode(value)
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
		encodedKey, err := encode(key)
		if err != nil {
			return "", nil
		}
		result += encodedKey
		value := input[key]
		encodedValue, err := encode(value)
		if err != nil {
			return "", nil
		}
		result += encodedValue
	}
	result += "e"
	return result, nil
}
func encode(input interface{}) (string, error) {
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
		value, index, err := decodeOne(bencodedString[currentIndex:])
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
		key, index, err := decodeOne(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		value, index, err := decodeOne(bencodedString[currentIndex:])
		if err != nil {
			return nil, -1, err
		}
		currentIndex = index + currentIndex
		dict[key.(string)] = value
	}
	return dict, currentIndex + 1, nil
}

func decodeOne(bencodedString string) (interface{}, int, error) {

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

func main() {
	command := os.Args[1]

	if command == "decode" {
		// Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, _, err := decodeOne(bencodedValue)
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
		decoded, _, err := decodeOne(string(byteContent))
		m := decoded.(map[string]interface{})
		info := m["info"].(map[string]interface{})
		fmt.Println("Tracker URL:", m["announce"])
		fmt.Println("Length:", info["length"])
		encoded, err := encode(info)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Info Hash:", fmt.Sprintf("%x", sha1.Sum([]byte(encoded))))
		}
		fmt.Println("Piece Length:", info["piece length"])
		fmt.Println("Piece Hashes:")
		pieces := info["pieces"].(string)
		for i := 0; i < len(pieces); i += 20 {
			fmt.Printf("%x\n", pieces[i:i+20])
		}
	} else if command == "peers" {
		fileName := os.Args[2]
		byteContent, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			return
		}
		decoded, _, err := decodeOne(string(byteContent))
		m := decoded.(map[string]interface{})
		info := m["info"].(map[string]interface{})
		encoded, err := encode(info)
		info_hash := sha1.Sum([]byte(encoded))
		queryParams := map[string]string{
			"info_hash":  string(info_hash[:]),
			"peer_id":    "00112233445566778899",
			"port":       "6881",
			"uploaded":   "0",
			"downloaded": "0",
			"left":       fmt.Sprint(info["piece length"].(int)),
			"compact":    "1",
		}

		resp, err := http.Get(createURL(m["announce"].(string), queryParams))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		response, _, err := decodeOne(string(body))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		peersString := response.(map[string]interface{})["peers"].(string)
		peers := []string{}
		for i := 0; i < len(peersString); i += 6 {
			result := ""
			result += fmt.Sprint(int(peersString[i])) + "."
			result += fmt.Sprint(int(peersString[i+1])) + "."
			result += fmt.Sprint(int(peersString[i+2])) + "."
			result += fmt.Sprint(int(peersString[i+3])) + ":"
			result += fmt.Sprint(int(peersString[i+4])<<8 | int(peersString[i+5]))
			peers = append(peers, result)
		}
		for _, peer := range peers {
			fmt.Println(peer)
		}

	} else if command == "handshake" {
		fileName := os.Args[2]
		torrentData, err := torrentDecode(fileName)
		info := torrentData["info"].(map[string]interface{})
		encoded, err := encode(info)
		info_hash := sha1.Sum([]byte(encoded))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		address := os.Args[3]
		var message []byte
		message = append(message, 19)
		message = append(message, []byte("BitTorrent protocol")...)
		message = append(message, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
		message = append(message, info_hash[:]...)
		message = append(message, []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}...)

		conn, err := net.Dial("tcp", address)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = conn.Write(message)
		if err != nil {

			fmt.Println(err)
			os.Exit(1)
		}
		_, err = conn.Read(message)

		if err != nil {

			fmt.Println(err)
			os.Exit(1)
		}
		hexString := hex.EncodeToString(message[20:40])
		fmt.Println("Peer ID:", hexString)

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
