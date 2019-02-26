// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// DNSRecord dns record
// swagger:model dnsRecord
type DNSRecord struct {

	// hostname
	Hostname string `json:"hostname,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// priority
	Priority int64 `json:"priority,omitempty"`

	// ttl
	TTL int64 `json:"ttl,omitempty"`

	// type
	Type string `json:"type,omitempty"`

	// value
	Value string `json:"value,omitempty"`
}

// Validate validates this dns record
func (m *DNSRecord) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DNSRecord) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DNSRecord) UnmarshalBinary(b []byte) error {
	var res DNSRecord
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}