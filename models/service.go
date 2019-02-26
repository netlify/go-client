// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// Service service
// swagger:model service
type Service struct {

	// created at
	CreatedAt string `json:"created_at,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// environments
	Environments []string `json:"environments"`

	// events
	Events []interface{} `json:"events"`

	// icon
	Icon string `json:"icon,omitempty"`

	// id
	ID string `json:"id,omitempty"`

	// long description
	LongDescription string `json:"long_description,omitempty"`

	// manifest url
	ManifestURL string `json:"manifest_url,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// service path
	ServicePath string `json:"service_path,omitempty"`

	// slug
	Slug string `json:"slug,omitempty"`

	// tags
	Tags []string `json:"tags"`

	// updated at
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Validate validates this service
func (m *Service) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Service) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Service) UnmarshalBinary(b []byte) error {
	var res Service
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
