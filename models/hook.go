// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// Hook hook
// swagger:model hook
type Hook struct {

	// created at
	CreatedAt string `json:"created_at,omitempty"`

	// data
	Data interface{} `json:"data,omitempty"`

	// disabled
	Disabled bool `json:"disabled,omitempty"`

	// event
	Event string `json:"event,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// site id
	SiteID string `json:"site_id,omitempty"`

	// type
	Type string `json:"type,omitempty"`

	// updated at
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Validate validates this hook
func (m *Hook) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Hook) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Hook) UnmarshalBinary(b []byte) error {
	var res Hook
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}