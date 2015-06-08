package yeelight

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"strconv"
	"strings"

	"errors"
)

type Hub struct {
	IP       string
	LightIDs []string // array of just the IDs of all lights the hub controls -- MAYBE? Maybe best in device, not here...
}

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
func GetLights(ip string) ([]Light, error) {
	response, err := SendCommand("GL\r\n", ip)
	if err != nil {
		log.Println("Error is:", err)
		return []Light{}, err
	} else {
		lights := getLightsFromString(response)
		return lights, err
	}

}

// SendCommand sends a single named command to the Yeelight hub via TCP
func SendCommand(cmd string, ip string) (string, error) {
	//	log.Printf("Sending command %s to IP %s\n", cmd, ip)
	conn, err := net.Dial("tcp", ip+":10003")
	// TODO: Figure out timeout - not working like this...
	//	raddr, err := net.ResolveTCPAddr("tcp", ip+":10003")
	//	conn, err := net.DialTCP("tcp", nil, raddr)
	//	conn.SetDeadline(time.Now().Add(100 * time.Millisecond))

	if err != nil {
		log.Println("Failed to connect: %s", err)
		return "", err
	}
	// send command to net connection, read it
	fmt.Fprintf(conn, cmd)
	response, err := bufio.NewReader(conn).ReadString('\n')
	//	conn.SetDeadline(time.Time{})
	conn.Close()
	return response, err
}

// TurnOffAllLights turns off all Yeelight bulbs on hub
func TurnOffAllLights(ip string) error {
	_, err := SendCommand("C G000,,,0,0,0\r\n", ip)
	return err
}

// SetLight sets the useful values (R, G, B, Brightness - ints) for a given light based on its ID (string)
func SetLight(id string, r, g, b, brightness int, ip string) error {
	// format string like command: "C 50F5,200,100,255,90,0\r\n"
	cmd := fmt.Sprintf("C %s,%d,%d,%d,%d,0\r\n", id, r, g, b, brightness)
	_, err := SendCommand(cmd, ip)
	return err
}

// SetOnOff sets the light to on (full brightness) when state is true, off when it is false
func SetOnOff(id string, state bool, ip string) error {
	var cmd string
	if state {
		cmd = fmt.Sprintf("C %s,,,,100,\r\n", id)
	} else {
		cmd = fmt.Sprintf("C %s,,,,0,\r\n", id)
	}
	//	fmt.Printf("%#v", cmd)
	_, err := SendCommand(cmd, ip)
	return err
}

// ToggleOnOff determines on/off state of chosen light (0 is off, anything > 0 is on)
// then calls SetOnOff to set opposite on/off state
func ToggleOnOff(id string, ip string) error {
	lights, err := GetLights(ip)
	if err != nil {
		return err
	}
	for i := 0; i < len(lights); i++ {
		if lights[i].ID == id {
			if lights[i].Level != 0 {
				err = SetOnOff(id, false, ip)
			} else {
				err = SetOnOff(id, true, ip)
			}
			break
		}
	}
	return err
}

// SetBrightness takes a float level (0-1) and sets the brightness of a light (0-100)
func SetBrightness(id string, level float64, ip string) error {
	// convert level fraction to int 0-100
	brightness := int(level * 100)
	cmd := fmt.Sprintf("C %s,,,,%d,\r\n", id, brightness)
	_, err := SendCommand(cmd, ip)
	return err
}

// SetColor takes r, g, b values (0-255) and sets the colour, leaving brightness unchanged
func SetColor(id string, r, g, b uint8, ip string) error {
	cmd := fmt.Sprintf("C %s,%d,%d,%d,,\r\n", id, r, g, b)
	_, err := SendCommand(cmd, ip)
	return err
}

// Heartbeat pings the Yeelight hub to see if it's alive, returns either nil error if it's responsive
// or error if the ack is not received from the hub.
func Heartbeat(ip string) error {
	response, err := SendCommand("HB\r\n", ip)
	if err != nil {
		return err
	}
	if response != "HACK\r\n" {
		return errors.New("Error. Hub not responding")
	}
	return nil
}

// DiscoverHub uses SSDP (UDP) to find and return the IP address of the Yeelight hub
// returns an empty string if not found
// ref for UDP code: https://groups.google.com/forum/#!topic/golang-nuts/Llfb0wMY9WI
func DiscoverHub() (string, error) {
	// TODO: Add timeout and return error when not found after a while
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

// getLightsFromString converts string response from Yeelight GL command
// into an array of Light structs with all the data for each light
func getLightsFromString(response string) []Light {
	var lights []Light
	lights = make([]Light, 0)
	if response == "" {
		fmt.Println(fmt.Errorf("Error, string is empty"))
	} else {
		response = strings.TrimLeft(response, "GLB ")
		// remove last ';\r\n' so we don't get an empty light
		response = strings.TrimRight(response, ";\r\n")
		lightStrings := strings.Split(response, ";")
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

// HSVToRGB converts an HSV triple to a RGB triple.
// from https://godoc.org/code.google.com/p/sadbox/color
// Ported from http://goo.gl/Vg1h9
func HSVToRGB(h, s, v float64) (r, g, b uint8) {
	var fR, fG, fB float64
	i := math.Floor(h * 6)
	f := h*6 - i
	p := v * (1.0 - s)
	q := v * (1.0 - f*s)
	t := v * (1.0 - (1.0-f)*s)
	switch int(i) % 6 {
	case 0:
		fR, fG, fB = v, t, p
	case 1:
		fR, fG, fB = q, v, p
	case 2:
		fR, fG, fB = p, v, t
	case 3:
		fR, fG, fB = p, q, v
	case 4:
		fR, fG, fB = t, p, v
	case 5:
		fR, fG, fB = v, p, q
	}
	r, g, b = float64ToUint8(fR), float64ToUint8(fG), float64ToUint8(fB)
	return
}

// float64ToUint8 converts a float64 to uint8.
// See: http://code.google.com/p/go/issues/detail?id=3423
func float64ToUint8(x float64) uint8 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 255
	}
	return uint8(int(x*255 + 0.5))
}

// TemperatureToRGB converts a colour temperature in kelvins in range [1000, 40000] to RGB
// https://gist.github.com/paulkaplan/5184275
// From http://www.tannerhelland.com/4435/convert-temperature-rgb-algorithm-code/
func TemperatureToRGB(kelvin float64) (r, g, b uint8) {
	temp := kelvin / 100

	var red, green, blue float64

	if temp <= 66 {

		red = 255

		green = temp
		green = 99.4708025861*math.Log(green) - 161.1195681661

		if temp <= 19 {
			blue = 0
		} else {
			blue = temp - 10
			blue = 138.5177312231*math.Log(blue) - 305.0447927307
		}

	} else {
		red = temp - 60
		red = 329.698727446 * math.Pow(red, -0.1332047592)

		green = temp - 60
		green = 288.1221695283 * math.Pow(green, -0.0755148492)

		blue = 255
	}
	return clamp(red, 0, 255), clamp(green, 0, 255), clamp(blue, 0, 255)
}

func clamp(x, min, max float64) uint8 {

	if x < min {
		return uint8(min)
	}
	if x > max {
		return uint8(max)
	}
	return uint8(x)
}
