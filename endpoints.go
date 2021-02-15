package lifx

import (
	"fmt"
	"net/url"
	"path"
)

func BuildURL(rawurl, rawpath string) string {
	u, _ := url.Parse(rawurl)
	u.Path = path.Join(u.Path, rawpath)
	return u.String()
}

var (
	Endpoint      = "https://api.lifx.com/v1"
	EndpointState = func(selector string) string {
		return BuildURL(Endpoint, fmt.Sprintf("/lights/%s/state", selector))
	}
	EndpointStateDelta = func(selector string) string {
		return BuildURL(Endpoint, fmt.Sprintf("/lights/%s/state/delta", selector))
	}
	EndpointListLights = func(selector string) string {
		return BuildURL(Endpoint, fmt.Sprintf("/lights/%s", selector))
	}
	EndpointStates = func() string {
		return BuildURL(Endpoint, "/lights/states")
	}
	EndpointColor = func() string {
		return BuildURL(Endpoint, "/color")
	}
	EndpointToggle = func(selector string) string {
		return BuildURL(Endpoint, fmt.Sprintf("/lights/%s/toggle", selector))
	}
)
