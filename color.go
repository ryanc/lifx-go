package lifx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type (
	Color interface {
		ColorString() string
	}
)

type (
	RGBColor struct {
		R, G, B uint8
	}

	HSBKColor struct {
		H *float32 `json:"hue"`
		S *float32 `json:"saturation"`
		B *float32 `json:"brightness"`
		K *int16   `json:"kelvin"`
	}

	NamedColor string
)

const (
	KelvinCandlelight    = 1500
	KelvinSunset         = 2000
	KelvinUltraWarm      = 2500
	KelvinIncandescent   = 2700
	KelvinWarm           = 3000
	KelvinCool           = 4000
	KelvinCoolDaylight   = 4500
	KelvinSoftDaylight   = 5000
	KelvinDaylight       = 5600
	KelvinNoonDaylight   = 6000
	KelvinBrightDaylight = 6500
	KelvinCloudDaylight  = 7000
	KelvinBlueDaylight   = 7500
	KelvinBlueOvercast   = 8000
	KelvinBlueIce        = 9000

	HueWhite  = 0
	HueRed    = 0
	HueOrange = 36
	HueYellow = 60
	HueGreen  = 120
	HueCyan   = 180
	HueBlue   = 250
	HuePurple = 280
	HuePink   = 325
)

var (
	DefaultWhites = map[string]int{
		"candlelight":    KelvinCandlelight,
		"sunset":         KelvinSunset,
		"ultrawarm":      KelvinUltraWarm,
		"incandescent":   KelvinIncandescent,
		"warm":           KelvinWarm,
		"cool":           KelvinCool,
		"cooldaylight":   KelvinCoolDaylight,
		"softdaylight":   KelvinSoftDaylight,
		"daylight":       KelvinDaylight,
		"noondaylight":   KelvinNoonDaylight,
		"brightdaylight": KelvinBrightDaylight,
		"clouddaylight":  KelvinCloudDaylight,
		"bluedaylight":   KelvinBlueDaylight,
		"blueovercast":   KelvinBlueOvercast,
		"blueice":        KelvinBlueIce,
	}
)

func NewRGBColor(r, g, b uint8) (RGBColor, error) {
	var c RGBColor
	if (r < 0 || r > 255) && (g < 0 || r > 255) && (b < 0 || b > 255) {
		return c, errors.New("values must be between 0-255")
	}

	return RGBColor{R: r, G: g, B: b}, nil
}

func NewHSColor(h, s float32) (HSBKColor, error) {
	var c HSBKColor

	if h < 0 || h > 360 {
		return c, errors.New("hue must be between 0.0-360.0")
	}
	if s < 0 || s > 1 {
		return c, errors.New("saturation must be between 0.0-1.0")
	}

	c = HSBKColor{
		H: Float32Ptr(h),
		S: Float32Ptr(s),
	}

	return c, nil
}

func NewHSBColor(h, s, b float32) (HSBKColor, error) {
	var c HSBKColor

	if h < 0 || h > 360 {
		return c, errors.New("hue must be between 0.0-360.0")
	}
	if s < 0 || s > 1 {
		return c, errors.New("saturation must be between 0.0-1.0")
	}
	if b < 0 || b > 1 {
		return c, errors.New("brightness must be between 0.0-1.0")
	}

	c = HSBKColor{
		H: Float32Ptr(h),
		S: Float32Ptr(s),
		B: Float32Ptr(b),
	}

	return c, nil
}

func NewRed() (HSBKColor, error)    { return NewHSColor(HueRed, 1) }
func NewOrange() (HSBKColor, error) { return NewHSColor(HueOrange, 1) }
func NewYellow() (HSBKColor, error) { return NewHSColor(HueYellow, 1) }
func NewGreen() (HSBKColor, error)  { return NewHSColor(HueGreen, 1) }
func NewCyan() (HSBKColor, error)   { return NewHSColor(HueCyan, 1) }
func NewPurple() (HSBKColor, error) { return NewHSColor(HuePurple, 1) }
func NewPink() (HSBKColor, error)   { return NewHSColor(HuePink, 1) }

func NewWhite(k int16) (HSBKColor, error) {
	var c HSBKColor

	if k < 1500 || k > 9000 {
		return c, errors.New("kelvin must be between 1500-9000")
	}

	c = HSBKColor{
		H: Float32Ptr(HueWhite),
		S: Float32Ptr(0.0),
		K: Int16Ptr(k),
	}

	return c, nil
}

func NewWhiteString(s string) (HSBKColor, error) {
	k, ok := DefaultWhites[s]

	if !ok {
		return HSBKColor{}, fmt.Errorf("'%s' is not a valid default white", s)
	}

	return NewWhite(int16(k))
}

func (c RGBColor) ColorString() string {
	return fmt.Sprintf("rgb:%d,%d,%d", c.R, c.G, c.B)
}

func (c RGBColor) Hex() string {
	return fmt.Sprintf("#%x%x%x", c.R, c.G, c.B)
}

func (c HSBKColor) ColorString() string {
	var s []string
	if c.H != nil {
		s = append(s, fmt.Sprintf("hue:%g", *c.H))
	}
	if c.S != nil {
		s = append(s, fmt.Sprintf("saturation:%g", *c.S))
	}
	if c.B != nil {
		s = append(s, fmt.Sprintf("brightness:%g", *c.B))
	}
	if c.K != nil {
		s = append(s, fmt.Sprintf("kelvin:%d", *c.K))
	}
	return strings.Join(s, " ")
}

func (c HSBKColor) MarshalText() ([]byte, error) {
	return []byte(c.ColorString()), nil
}

func (c RGBColor) MarshalText() ([]byte, error) {
	return []byte(c.ColorString()), nil
}

func (c NamedColor) ColorString() string {
	return string(c)
}

func (c *Client) ValidateColor(color Color) (Color, error) {
	var (
		err  error
		s    *HSBKColor
		r    *http.Response
		resp *Response
	)

	if resp, err = c.validateColor(color); err != nil {
		return nil, err
	}

	resp, err = NewResponse(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}
