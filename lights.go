package lifx

import (
	//"crypto/tls"
	"encoding/json"
	"net/http"
	"time"
)

const (
	OK       Status = "ok"
	TimedOut Status = "timed_out"
	Offline  Status = "offline"
)

type (
	Status string

	Selector struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	Product struct {
		Name         string       `json:"name"`
		Identifier   string       `json:"identifier"`
		Company      string       `json:"company"`
		Capabilities Capabilities `json:"capabilities"`
	}

	Capabilities struct {
		HasColor             bool    `json:"has_color"`
		HasVariableColorTemp bool    `json:"has_variable_color_temp"`
		HasIR                bool    `json:"has_ir"`
		HasChain             bool    `json:"has_chain"`
		HasMultizone         bool    `json:"has_multizone"`
		MinKelvin            float64 `json:"min_kelvin"`
		MaxKelvin            float64 `json:"max_kelvin"`
	}

	Light struct {
		Id              string    `json:"id"`
		UUID            string    `json:"uuid"`
		Label           string    `json:"label"`
		Connected       bool      `json:"connected"`
		Power           string    `json:"power"`
		Color           HSBKColor `json:"color"`
		Brightness      float64   `json:"brightness"`
		Effect          string    `json:"effect"`
		Group           Selector  `json:"group"`
		Location        Selector  `json:"location"`
		Product         Product   `json:"product"`
		LastSeen        time.Time `json:"last_seen"`
		SecondsLastSeen float64   `json:"seconds_last_seen"`
	}

	State struct {
		Power      string  `json:"power,omitempty"`
		Color      Color   `json:"color,omitempty"`
		Brightness float64 `json:"brightness,omitempty"`
		Duration   float64 `json:"duration,omitempty"`
		Infrared   float64 `json:"infrared,omitempty"`
		Fast       bool    `json:"fast,omitempty"`
	}

	StateDelta struct {
		Power      *string  `json:"power,omitempty"`
		Duration   *float64 `json:"duration,omitempty"`
		Infrared   *float64 `json:"infrared,omitempty"`
		Hue        *float64 `json:"hue,omitempty"`
		Saturation *float64 `json:"saturation,omitempty"`
		Brightness *float64 `json:"brightness,omitempty"`
		Kelvin     *int     `json:"kelvin,omitempty"`
	}

	StateWithSelector struct {
		State
		Selector string `json:"selector"`
	}

	States struct {
		States   []StateWithSelector `json:"states,omitempty"`
		Defaults State               `json:"defaults,omitempty"`
	}

	Toggle struct {
		Duration float64 `json:"duration,omitempty"`
	}
)

func (c *Client) SetState(selector string, state State) (*LifxResponse, error) {
	var (
		err  error
		s    *LifxResponse
		resp *Response
	)

	if resp, err = c.setState(selector, state); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, resp.GetLifxError()
	}

	if state.Fast && resp.StatusCode == http.StatusAccepted {
		return nil, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) FastSetState(selector string, state State) (*LifxResponse, error) {
	state.Fast = true
	return c.SetState(selector, state)
}

func (c *Client) SetStates(selector string, states States) (*LifxResponse, error) {
	var (
		err  error
		s    *LifxResponse
		resp *Response
	)

	if resp, err = c.setStates(selector, states); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) StateDelta(selector string, delta StateDelta) (*LifxResponse, error) {
	var (
		err  error
		s    *LifxResponse
		resp *Response
	)

	if resp, err = c.stateDelta(selector, delta); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) Toggle(selector string, duration float64) (*LifxResponse, error) {
	var (
		err  error
		s    *LifxResponse
		resp *Response
	)

	if resp, err = c.toggle(selector, duration); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return nil, resp.GetLifxError()
	}

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) ListLights(selector string) ([]Light, error) {
	var (
		err  error
		s    []Light
		resp *Response
	)

	if resp, err = c.listLights(selector); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, resp.GetLifxError()
	}

	if err = json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Client) PowerOff(selector string) (*LifxResponse, error) {
	return c.SetState(selector, State{Power: "off"})
}

func (c *Client) FastPowerOff(selector string) {
	c.SetState(selector, State{Power: "off", Fast: true})
}

func (c *Client) PowerOn(selector string) (*LifxResponse, error) {
	return c.SetState(selector, State{Power: "on"})
}

func (c *Client) FastPowerOn(selector string) {
	c.SetState(selector, State{Power: "on", Fast: true})
}
