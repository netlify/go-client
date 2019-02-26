// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/netlify/go-client/models"
)

// UpdateSiteReader is a Reader for the UpdateSite structure.
type UpdateSiteReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateSiteReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewUpdateSiteOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		result := NewUpdateSiteDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewUpdateSiteOK creates a UpdateSiteOK with default headers values
func NewUpdateSiteOK() *UpdateSiteOK {
	return &UpdateSiteOK{}
}

/*UpdateSiteOK handles this case with default header values.

OK
*/
type UpdateSiteOK struct {
	Payload *models.Site
}

func (o *UpdateSiteOK) Error() string {
	return fmt.Sprintf("[PATCH /sites/{site_id}][%d] updateSiteOK  %+v", 200, o.Payload)
}

func (o *UpdateSiteOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Site)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateSiteDefault creates a UpdateSiteDefault with default headers values
func NewUpdateSiteDefault(code int) *UpdateSiteDefault {
	return &UpdateSiteDefault{
		_statusCode: code,
	}
}

/*UpdateSiteDefault handles this case with default header values.

error
*/
type UpdateSiteDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the update site default response
func (o *UpdateSiteDefault) Code() int {
	return o._statusCode
}

func (o *UpdateSiteDefault) Error() string {
	return fmt.Sprintf("[PATCH /sites/{site_id}][%d] updateSite default  %+v", o._statusCode, o.Payload)
}

func (o *UpdateSiteDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}