package main

import (
	"github.com/lindsaymarkward/go-yeelight"
	"fmt"
)

// main shows examples of how to use the library
func main() {
	ip, err := yeelight.DiscoverHub()
	if err == nil {
		fmt.Printf("Hub found at %s\n", ip)
	} else {
		fmt.Printf("Error: %s", err)
	}

	/*
		// get and display the current lights and their values
		lights, _ := yeelight.GetLights()
		fmt.Println(lights)

		// send a raw command, print the response ("CB")
		response, _ := yeelight.SendCommand("C 50F5,0,200,255,80,0\r\n", yeelight.IP)
		fmt.Println(response)

		// set a particular light to medium bright magenta
		yeelight.SetLight(lights[3].ID, 255, 0, 255, 50)

		// loop and fade one light up and another down
		//	for i := 0; i < 100; i++ {
		//		yeelight.SetLight("3CB8", 250, 0, 250, 100-i)
		//		yeelight.SetLight("50F5", 250, 250, 0, i)
		//		fmt.Print(time.Second)
		//		time.Sleep(50 * time.Millisecond)
		//	}

		time.Sleep(3 * time.Second)

		// turn off all the lights
		yeelight.TurnOffAllLights()
	*/
}
