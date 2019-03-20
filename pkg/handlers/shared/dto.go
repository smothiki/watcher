package shared

import (
	"fmt"
	"strings"
)

// Service struct is used to describe a service
type ServicePayload struct {
	Name      string `validate:"-" json:"-"`
	Namespace string `validate:"required" json:"namespace"`
	Host      string `validate:"required,ipv4" json:"host"`
	Port      int    `validate:"required,min=1,max=65535" json:"port"`
	Protocol  string `validate:"required" json:"protocol,omitempty"`
	// FATHER LEVEL DOMAIN
	FLDomain    string `validate:"-" json:"fl_domain"`
	HealthCheck struct {
		Path string `validate:"-" json:"path,omitempty"`
		Port int    `validate:"required,min=1,max=65535" json:"port,omitempty"`
	} `json:"health_check"`
}

// Return a string consisting of host and port
func (s *ServicePayload) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Return a legal coredns parsing record
func (s *ServicePayload) DNSRecord() string {
	return fmt.Sprintf(`{"host":"%s"}`, s.Host)
}

// Return the name of Dns, which consists of FLD and name
func (s *ServicePayload) DNSName() string {
	if s.FLDomain == "" {
		return s.Name
	}

	return s.FLDomain + "/" + s.Name
}

// Return the key of Dns
func (s *ServicePayload) DNSKey() string {
	return strings.Replace(s.Host, ".", "-", -1)
}
