package freebox

import "fmt"

// LanHost
type LanHost struct {
	L2Ident struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"l2ident"`
	Active            bool   `json:"active"`
	ID                string `json:"id"`
	LastTimeReachable Epoch  `json:"last_time_reachable"`
	Persistent        bool   `json:"persistent"`
	VendorName        string `json:"vendor_name"`
	HostType          string `json:"host_type"`
	PrimaryName       string `json:"primary_name"`
	L3Connectivities  []struct {
		Addr              string `json:"addr"`
		Active            bool   `json:"active"`
		Reachable         bool   `json:"reachable"`
		LastActivity      Epoch  `json:"last_activity"`
		AF                string `json:"af"`
		LastTimeReachable Epoch  `json:"last_time_reachable"`
	} `json:"l3connectivities"`
	Reachable         bool   `json:"reachable"`
	LastActivity      Epoch  `json:"last_activity"`
	PrimaryNameManual bool   `json:"primary_name_manual"`
	Interface         string `json:"interface"`
}

// Get all lan hosts
func (c *Client) GetLanHosts() ([]LanHost, error) {
	payload := []LanHost{}
	err := c.GetResult("lan/browser/pub/", &payload)
	return payload, err
}

// Get a contact
func (c *Client) GetLanHost(name string) (LanHost, error) {
	payload := LanHost{}
	err := c.GetResult(fmt.Sprintf("lan/browser/pub/%s", name), &payload)
	return payload, err
}
