// http://www.developers.meethue.com/documentation/lights-api

package hue

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "strings"
    "errors"
    "bytes"
)

type Light struct {
    State struct {
        On          bool      `json:"on"`     // On or Off state of the light ("true" or "false")
        Bri         int       `json:"bri"`    // Brightness value 1-254
        Hue         int       `json:"hue"`    // Hue value 1-65535
        Saturation  int       `json:"sat"`    // Saturation value 0-254
        Effect      string    `json:"effect"` // "None" or "Colorloop"
        XY          [2]float32 `json:"xy"`    // Coordinates of color in CIE color space
        CT          int       `json:"ct"`     // Mired Color Temperature (google it)
        Alert       string    `json:"alert"`
        ColorMode   string    `json:"colormode"`
        Reachable   bool      `json:"reachable"`
    } `json:"state"`
    Type             string     `json:"type"`
    Name             string     `json:"name"`
    ModelID          string     `json:"modelid"`
    ManufacturerName string     `json:"manufacturername"`
    UniqueID         string     `json:"uniqueid"`
    SWVersion        string     `json:"swversion"`
}

// LightState used in SetLightState to ammend light attributes.
type LightState struct {
    On  bool
    Bri uint8
    Hue uint16
    Sat uint8
    XY  [2]float32
    CT  uint16
    Alert   string
    Effect  string
    TransitionTime string
    BrightnessIncrement  int // TODO: -254 to 254
    SaturationIncrement  int // TODO: -254 to 254
    HueIncrement    int // TODO: -65534 to 65534
    CTIncrement     int // TODO: -65534 to 65534
    XYIncrement     [2]float32
}

// SetLightState will modify light attributes such as on/off, saturation,
// brightness, and more. See `SetLightState` struct.
func SetLightState(bridge *Bridge, lightID string, newState LightState) {
    // Construct the http POST
    req, err := json.Marshal(newState)
    if err != nil {
        trace("", err)
    }

    // Send the request and read the response
    uri := fmt.Sprintf("http://%s/api/%s/lights/%s/state",
        bridge.IPAddress, bridge.Username, lightID)
    resp, err := http.Post(uri, "text/json", bytes.NewReader(req))
    if err != nil {
        trace("", err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        trace("", err)
    }

    // TODO: parse the response

    fmt.Println(string(body))
}

//http://192.168.1.128/api/319b36233bd2328f3e40731b23479207/lights/

// GetAllLights retreives the state of all lights that the bridge is aware of.
func GetAllLights(bridge *Bridge) []Light {
    // Loop through all light indicies to see if they exist
    // and parse their values. Supports 100 lights.
    var lights []Light
    for index := 1; index < 101; index++ {
        response, err := http.Get(
            fmt.Sprintf("http://%s/api/%s/lights/%d", bridge.IPAddress, bridge.Username, index))
        if err != nil {
            trace("", err)
        } else if response.StatusCode != 200 {
            trace(fmt.Sprintf("Bridge status error %d", response.StatusCode), nil)
        }

        // Read the response
        body, err := ioutil.ReadAll(response.Body)
        defer response.Body.Close()
        if err != nil {
            trace("", err)
        }
        if strings.Contains(string(body), "not available") {
            // Handle end of searchable lights
            fmt.Printf("\n\n%d lights found.\n\n", index)
            break
        }

        // Parse and load the response into the light array
        data := Light{}
        err = json.Unmarshal(body, &data)
        if err != nil {
            trace("", err)
        }
        lights = append(lights, data)
    }
    return lights
}

// GetLight will return a light struct containing data on a given name.
func GetLight(bridge *Bridge, name string) (Light, error) {
    lights := GetAllLights(bridge)
    for index := 0; index < len(lights); index++ {
        if lights[index].Name == name {
            return lights[index], nil
        }
    }
    return Light{}, errors.New("Light not found.")
}
