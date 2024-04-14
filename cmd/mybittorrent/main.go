package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go"
)

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

func main() {
	out := flag.String("o", "", "specify output file to write torrent piece")
	flag.CommandLine.Parse(os.Args[2:])
	command := os.Args[1]
	args := flag.Args()
	if command == "decode" {
		decode(args[0])
	} else if command == "info" {
		torrentInfo(args[0])
	} else if command == "peers" {
		findPeers(args[0])
	} else if command == "handshake" {
		handShake(args[0])
	} else if command == "download_piece" {
		downloadPiece(*out, args[0], args[1])
	} else if command == "download" {
		downloadFile(*out, args[0])

	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

func handShake(fileName string) {
	torrentData, err := bittorrent.TorrentDecode(fileName)
	info := torrentData["info"].(map[string]interface{})
	encoded, err := bittorrent.Encode(info)
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
	message = append(message, []byte("00112233445566778899")...)

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
	hexString := hex.EncodeToString(message[48:68])
	fmt.Println("Peer ID:", hexString)
}

func findPeers(fileName string) {
	byteContent, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := bittorrent.Decode(string(byteContent))
	m := decoded.(map[string]interface{})
	info := m["info"].(map[string]interface{})
	encoded, err := bittorrent.Encode(info)
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
	response, _, err := bittorrent.Decode(string(body))
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
}
func torrentInfo(fileName string) {

	byteContent, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := bittorrent.Decode(string(byteContent))
	m := decoded.(map[string]interface{})
	info := m["info"].(map[string]interface{})
	fmt.Println("Tracker URL:", m["announce"])
	fmt.Println("Length:", info["length"])
	encoded, err := bittorrent.Encode(info)
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
}

func decode(bencodedValue string) {
	decoded, _, err := bittorrent.Decode(bencodedValue)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}

func downloadPiece(outFile, torrentFile, piece string) {
	pieceNumber, err := strconv.Atoi(piece)
	if err != nil {
		fmt.Println(err)
		return
	}

	byteContent, err := os.ReadFile(torrentFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := bittorrent.Decode(string(byteContent))
	m := decoded.(map[string]interface{})
	info := m["info"].(map[string]interface{})
	encoded, err := bittorrent.Encode(info)
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
	response, _, err := bittorrent.Decode(string(body))
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

	var message []byte
	message = append(message, 19)
	message = append(message, []byte("BitTorrent protocol")...)
	message = append(message, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	message = append(message, info_hash[:]...)
	message = append(message, []byte("00112233445566778899")...)

	conn, err := net.Dial("tcp", peers[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// send handshake message
	_, err = conn.Write(message)
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Sent Handshake Message")
	//receive handshake message
	_, err = conn.Read(message)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Received Handshake Message")

	pcs := info["pieces"].(string)
	pieces := [][]byte{}
	for i := 0; i < len(pcs); i += 20 {
		pieces = append(pieces, []byte(pcs[i:i+20]))
	}

	// receive bitfield message
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Received Bitfield Message:", buffer[:n])

	// send interested message
	n, err = conn.Write([]byte{0, 0, 0, 4, 2})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Sent Interested Message.")

	// receive unchoke message
	buffer = make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ask for pieces
	pieceLength := info["piece length"].(int)
	maxBlockSize := 16 * 1024
	result := []byte{}
	noOfPieces := int(math.Ceil(float64(info["length"].(int)) / float64(pieceLength)))
	if pieceNumber == noOfPieces-1 {
		pieceLength = info["length"].(int) - (noOfPieces-1)*pieceLength
	}
	noOfBlocks := int(math.Ceil(float64(pieceLength) / float64(maxBlockSize)))
	// fmt.Println("pieceLength:", pieceLength)
	// fmt.Println("noOfBlocks:", noOfBlocks)
	for i := 0; i < noOfBlocks; i++ {
		// fmt.Println("i:", i)
		downloaded := 0
		blockSize := maxBlockSize
		if i == (noOfBlocks - 1) {
			blockSize = (pieceLength - (noOfBlocks-1)*maxBlockSize)
		}
		// fmt.Println("blockSize:", blockSize)

		data := []byte{0, 0, 0, 9, 6}
		index := intToBytes(int32(pieceNumber))
		begin := intToBytes(int32(maxBlockSize * i))
		length := intToBytes(int32(blockSize))
		// fmt.Println("index", index, "length", length, "begin", begin)
		data = append(data, index...)
		data = append(data, begin...)
		data = append(data, length...)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error Here during write")
			fmt.Println(err)
			os.Exit(1)
		}

		data = make([]byte, int(math.Pow(2, 15)))
		n, err = conn.Read(data)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// fmt.Println("Current Download Piece:", data[:13], (n - 13))
		downloaded += (n - 13)
		result = append(result, data[13:n]...)
		// fmt.Println("downloaded:", downloaded)
		for downloaded < blockSize {

			data = make([]byte, int(math.Pow(2, 15)))
			n, err = conn.Read(data)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// first four bytes : prefix length
			// then the message id : request 6
			result = append(result, data[:n]...)
			downloaded += n
			// fmt.Println("downloaded:", downloaded)
		}
	}
	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = f.Write(result)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Piece %v downloaded to %v.\n", pieceNumber, outFile)
}

func downloadFile(outFile, torrentFile string) {

	byteContent, err := os.ReadFile(torrentFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	decoded, _, err := bittorrent.Decode(string(byteContent))
	m := decoded.(map[string]interface{})
	info := m["info"].(map[string]interface{})
	encoded, err := bittorrent.Encode(info)
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
	response, _, err := bittorrent.Decode(string(body))
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

	var message []byte
	message = append(message, 19)
	message = append(message, []byte("BitTorrent protocol")...)
	message = append(message, []byte{0, 0, 0, 0, 0, 0, 0, 0}...)
	message = append(message, info_hash[:]...)
	message = append(message, []byte("00112233445566778899")...)

	conn, err := net.Dial("tcp", peers[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// send handshake message
	_, err = conn.Write(message)
	if err != nil {

		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Sent Handshake Message")
	//receive handshake message
	_, err = conn.Read(message)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Received Handshake Message")

	pcs := info["pieces"].(string)
	pieces := [][]byte{}
	for i := 0; i < len(pcs); i += 20 {
		pieces = append(pieces, []byte(pcs[i:i+20]))
	}

	// receive bitfield message
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Received Bitfield Message:", buffer[:n])

	// send interested message
	n, err = conn.Write([]byte{0, 0, 0, 4, 2})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// fmt.Println("Sent Interested Message.")

	// receive unchoke message
	buffer = make([]byte, 1024)
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ask for pieces
	pieceLength := info["piece length"].(int)
	fileLength := info["length"].(int)
	maxBlockSize := 16 * 1024
	noOfPieces := int(math.Ceil(float64(info["length"].(int)) / float64(pieceLength)))

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	calculatedFileSize := 0
	for pieceNumber := 0; pieceNumber < noOfPieces; pieceNumber++ {
		result := []byte{}
		if pieceNumber == noOfPieces-1 {
			pieceLength = fileLength - (noOfPieces-1)*pieceLength
		}
		noOfBlocks := int(math.Ceil(float64(pieceLength) / float64(maxBlockSize)))
		// fmt.Println("pieceLength:", pieceLength)
		// fmt.Println("noOfBlocks:", noOfBlocks)
		for i := 0; i < noOfBlocks; i++ {
			// fmt.Println("i:", i)
			downloaded := 0
			blockSize := maxBlockSize
			if i == (noOfBlocks - 1) {
				blockSize = (pieceLength - (noOfBlocks-1)*maxBlockSize)
			}
			// fmt.Println("blockSize:", blockSize)

			data := []byte{0, 0, 0, 9, 6}
			index := intToBytes(int32(pieceNumber))
			begin := intToBytes(int32(maxBlockSize * i))
			length := intToBytes(int32(blockSize))
			// fmt.Println("index", index, "length", length, "begin", begin)
			data = append(data, index...)
			data = append(data, begin...)
			data = append(data, length...)
			_, err = conn.Write(data)
			if err != nil {
				fmt.Println("Error Here during write")
				fmt.Println(err)
				os.Exit(1)
			}

			readData := make([]byte, int(math.Pow(2, 15)))
			n, err = conn.Read(readData)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			// fmt.Println("Current Download Piece:", data[:13], (n - 13))
			downloaded += (n - 13)
			calculatedFileSize += (n - 13)
			result = append(result, readData[13:n]...)
			// fmt.Println("downloaded:", downloaded)
			for downloaded < blockSize {

				readData = make([]byte, int(math.Pow(2, 15)))
				n, err = conn.Read(readData)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				// first four bytes : prefix length
				// then the message id : request 6
				result = append(result, readData[:n]...)
				calculatedFileSize += n
				downloaded += n
				// fmt.Println("downloaded:", downloaded)
			}
		}
		_, err = f.Write(result)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	fmt.Printf("Downloaded %v to %v.\n", torrentFile, outFile)
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
func intToBytes(num int32) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, num)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return buf.Bytes()
}
