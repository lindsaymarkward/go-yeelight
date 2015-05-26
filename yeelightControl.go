package yeelight

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var IP = "192.168.1.59"

type Light struct {
	ID                                        string
	Type, Online, LQI, R, G, B, Level, Effect int
	/*
	   id = HEX code for light ID
	   type =  0 or 1 (always 1)
	   online = 0 or 1 (1 is online)
	   lqi = LED ZigBee signal, 0-100  *I think
	   r, g, b = 0-255...
	   level = 0-100 brightness
	   effect = reserved/not implemented by Yeelight yet
	*/
}

// GetLights queries the Yeelight hub for current status of all lights and
// returns an array of Light structs with the values
func GetLights() ([]Light, error) {
	response, err := SendCommand("GL\r\n", IP)
	if err != nil {
		fmt.Println("Error is:", err) // TODO - change to log
		return []Light{}, err
	} else {
		lights := getLightsFromString(response)
		return lights, err
	}

}

// SendCommand sends a single named command to the Yeelight hub
func SendCommand(cmd string, ip string) (string, error) {

	//	log("Sending command %s to IP %s", cmd, ip)

	conn, err := net.Dial("tcp", ip+":10003")
	if err != nil {
		fmt.Println("Failed to connect: %s", err)
		return "", err
	}

	fmt.Fprintf(conn, cmd)
	status, err := bufio.NewReader(conn).ReadString('\n')
	conn.Close()
	return status, err
}

// TurnOffAllLights turns off all Yeelight bulbs on hub
func TurnOffAllLights() error {
	_, err := SendCommand("C G000,,,0,0,0\r\n", IP)
	return err
}

// SetLight sets the useful values (R, G, B, Brightness - ints) for a given light based on its ID (string)
func SetLight(id string, r, g, b, brightness int) error {
	// format string like command: "C 50F5,200,100,255,90,0\r\n"
	cmd := fmt.Sprintf("C %s,%d,%d,%d,%d,0\r\n", id, r, g, b, brightness)
	_, err := SendCommand(cmd, IP)
	return err
}

// getLightsFromString converts string response from Yeelight GL command
// into an array of Light structs with all the data for each light
func getLightsFromString(response string) []Light {
	var lights []Light
	lights = make([]Light, 0)
	if response == "" {
		fmt.Println(fmt.Errorf("Error")) // TODO: find out how to use errors
	} else {
		response = strings.TrimLeft(response, "GLB ")
		// remove last ';' so we don't get an empty light
		response = strings.TrimRight(response, ";")
		lightStrings := strings.Split(response, ";")
		//		fmt.Println(lightStrings)
		for _, lightString := range lightStrings {
			parts := strings.Split(lightString, ",")
			address := parts[0]
			// convert strings to ints for data values
			var values [8]int
			for i := 1; i < len(parts); i++ {
				values[i-1], _ = strconv.Atoi(parts[i])
			}
			newLight := Light{address, values[0], values[1], values[2], values[3], values[4], values[5], values[6], values[7]}
			lights = append(lights, newLight)
		}
	}
	return lights
}

// DiscoverHub uses SSDP (UDP) to find and return the IP address of the Yeelight hub
// returns an empty string if not found
// ref: https://groups.google.com/forum/#!topic/golang-nuts/Llfb0wMY9WI
func DiscoverHub() (string, error) {
	searchString := "M-SEARCH * HTTP/1.1\r\n HOST:239.255.255.250:1900\r\n MAN:\"ssdp:discover\"\r\n ST:yeelink:yeebox\r\n MAC:00000001\r\n MX:3\r\n\n\r\n"
	ip := ""
	ssdp, _ := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	c, _ := net.ListenPacket("udp4", ":0")
	socket := c.(*net.UDPConn)

	message := []byte(searchString)
	socket.WriteToUDP(message, ssdp)
	answerBytes := make([]byte, 1024)
	// stores result in answerBytes (pass-by-reference)
	_, _, err := socket.ReadFromUDP(answerBytes)
	if err == nil {
		response := string(answerBytes)
		// extract IP address from full response
		startIndex := strings.Index(response, "LOCATION: ") + 10
		endIndex := strings.Index(response, "MAC: ") - 2
		ip = response[startIndex:endIndex]
	}
	return ip, err
}