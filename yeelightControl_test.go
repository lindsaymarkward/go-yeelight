package yeelight

import "testing"

func TestGetLightsFromString(t *testing.T) {
    response := "GLB 0001,1,1,58,255,0,255,0,0;143E,1,1,50,199,97,255,0,0;287B,1,1,16,255,255,255,0,0;3CB8,1,1,53,0,255,255,100,0;50F5,1,1,61,255,255,255,0,0;6532,1,1,86,217,46,255,0,0;50F6,1,1,61,255,255,255,0,0;143F,1,1,11,255,255,255,0,0;"
//    response = ""
    GetLightsFromString(response)

}
