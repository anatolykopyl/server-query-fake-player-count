package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"sort"
	"time"
)

const (
	originalServerAddr = "localhost:27015"
	proxyPort          = ":27016"
)

const (
	A2S_INFO   = 0x54
	A2S_PLAYER = 0x55
)

const (
	byteLength     = 1
	shortLength    = 16 / 8
	longLength     = 32 / 8
	floatLength    = 32 / 8
	longLongLength = 64 / 8
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", proxyPort)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("Proxy server running on", proxyPort)

	for {
		buffer := make([]byte, 1024)
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		go handleRequest(conn, buffer[:n], clientAddr)
	}
}

func handleRequest(conn *net.UDPConn, request []byte, clientAddr *net.UDPAddr) {
	if len(request) < 5 {
		return
	}

	queryType := request[byteLength*4]

	response, err := forwardToOriginalServer(request)
	if err != nil {
		log.Println("Error forwarding to original server:", err)
		return
	}

	switch queryType {
	case A2S_INFO:
		modifiedResponse := modifyInfoResponse(response)
		conn.WriteToUDP(modifiedResponse, clientAddr)
	case A2S_PLAYER:
		modifiedResponse := modifyPlayerResponse(response)
		conn.WriteToUDP(modifiedResponse, clientAddr)
	default:
		conn.WriteToUDP(response, clientAddr)
	}
}

func forwardToOriginalServer(request []byte) ([]byte, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", originalServerAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Send the request to the original server
	_, err = conn.Write(request)
	if err != nil {
		return nil, err
	}

	// Read the response from the original server
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response[:n], nil
}

func modifyInfoResponse(response []byte) []byte {
	const (
		headerOffset   = 0
		protocolOffset = headerOffset + byteLength
	)

	nameOffset := protocolOffset + byteLength
	mapOffset := nameOffset + maxStringLength(response[nameOffset:])
	folderOffset := mapOffset + maxStringLength(response[mapOffset:])
	gameOffset := folderOffset + maxStringLength(response[folderOffset:])
	iDOffset := gameOffset + maxStringLength(response[gameOffset:])

	playersOffset := iDOffset + shortLength

	originalPlayerCount := response[playersOffset]
	maxPlayers := response[playersOffset+byteLength]

	turnUp := byte(getValueBasedOnTime())

	modifiedPlayerCount := originalPlayerCount + turnUp
	if modifiedPlayerCount > maxPlayers {
		modifiedPlayerCount = maxPlayers
	}

	modifiedResponse := make([]byte, len(response))
	copy(modifiedResponse, response)

	modifiedResponse[playersOffset] = modifiedPlayerCount

	return modifiedResponse
}

func modifyPlayerResponse(response []byte) []byte {
	const (
		headerOffset  = 4
		playersOffset = headerOffset + byteLength
	)

	log.Println(response)
	if response[headerOffset] == 0x41 {
		return response
	}

	turnUp := byte(getValueBasedOnTime())

	originalPlayerCount := response[playersOffset]
	modifiedPlayerCount := originalPlayerCount + turnUp

	modifiedResponse := make([]byte, len(response))
	copy(modifiedResponse, response)

	modifiedResponse[playersOffset] = modifiedPlayerCount

	var durations = []float32{}
	for i := byte(0); i < turnUp; i++ {
		randomFloat := float32(rand.Float64()) * 10000
		durations = append(durations, randomFloat)
	}

	sort.Slice(durations, func(i, j int) bool {
		return i > j
	})

	for i := byte(0); i < turnUp; i++ {
		// duration := make([]byte, 4)
		// binary.LittleEndian.PutUint32(duration, math.Float32bits(durations[i]))

		var (
			index    = []byte{0x00}
			name     = []byte{0x00}
			score    = []byte{0x00, 0x00, 0x00, 0x00}
			duration = []byte{0x00, 0x00, 0x00, 0x01}
		)

		modifiedResponse = append(modifiedResponse, index...)
		modifiedResponse = append(modifiedResponse, name...)
		modifiedResponse = append(modifiedResponse, score...)
		modifiedResponse = append(modifiedResponse, duration...)
	}

	log.Println(modifiedResponse)

	return modifiedResponse
}

func maxStringLength(data []byte) int {
	// Calculate length of the string until the null terminator (0x00)
	length := 0
	for length < len(data) && data[length] != 0x00 {
		length++
	}
	return length + 1 // Include null terminator
}

// GetValueBasedOnTime returns an integer between 0 and 20 based on current time.
// The maximum value (20) corresponds to 20:00.
func getValueBasedOnTime() int {
	now := time.Now()

	totalMinutes := now.Hour()*60 + now.Minute()

	maxMinutes := 20 * 60
	if totalMinutes < 0 {
		return 0
	}
	if totalMinutes >= maxMinutes {
		return 20
	}

	return int(float64(totalMinutes) / float64(maxMinutes) * 20)
}
