package proto

type Register struct {
	Id                string   `json:"id"`
	Name              string   `json:"name"`
	Address           string   `json:"address"`
	Tags              []string `json:"tags"`
	Port              int      `json:"port"`
	EnableTagOverride bool     `json:"enableTagOverride"`
}

type HealthCheck struct {
	Name           string `json:"name"`
	Script         string `json:"script,omitempty"`
	Interval       string `json:"interval"`
	Timeout        string `json:"Timeout"`
	Deregister_Csa string `json:"deregister_critical_service_after"`
	ServiceID      string `json:"ServiceID"`
	Tcp            string `json:"tcp,omitempty"`
}
