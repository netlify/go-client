// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewExchangeTicketParams creates a new ExchangeTicketParams object
// with the default values initialized.
func NewExchangeTicketParams() *ExchangeTicketParams {
	var ()
	return &ExchangeTicketParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewExchangeTicketParamsWithTimeout creates a new ExchangeTicketParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewExchangeTicketParamsWithTimeout(timeout time.Duration) *ExchangeTicketParams {
	var ()
	return &ExchangeTicketParams{

		timeout: timeout,
	}
}

// NewExchangeTicketParamsWithContext creates a new ExchangeTicketParams object
// with the default values initialized, and the ability to set a context for a request
func NewExchangeTicketParamsWithContext(ctx context.Context) *ExchangeTicketParams {
	var ()
	return &ExchangeTicketParams{

		Context: ctx,
	}
}

// NewExchangeTicketParamsWithHTTPClient creates a new ExchangeTicketParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewExchangeTicketParamsWithHTTPClient(client *http.Client) *ExchangeTicketParams {
	var ()
	return &ExchangeTicketParams{
		HTTPClient: client,
	}
}

/*ExchangeTicketParams contains all the parameters to send to the API endpoint
for the exchange ticket operation typically these are written to a http.Request
*/
type ExchangeTicketParams struct {

	/*TicketID*/
	TicketID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the exchange ticket params
func (o *ExchangeTicketParams) WithTimeout(timeout time.Duration) *ExchangeTicketParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the exchange ticket params
func (o *ExchangeTicketParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the exchange ticket params
func (o *ExchangeTicketParams) WithContext(ctx context.Context) *ExchangeTicketParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the exchange ticket params
func (o *ExchangeTicketParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the exchange ticket params
func (o *ExchangeTicketParams) WithHTTPClient(client *http.Client) *ExchangeTicketParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the exchange ticket params
func (o *ExchangeTicketParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithTicketID adds the ticketID to the exchange ticket params
func (o *ExchangeTicketParams) WithTicketID(ticketID string) *ExchangeTicketParams {
	o.SetTicketID(ticketID)
	return o
}

// SetTicketID adds the ticketId to the exchange ticket params
func (o *ExchangeTicketParams) SetTicketID(ticketID string) {
	o.TicketID = ticketID
}

// WriteToRequest writes these params to a swagger request
func (o *ExchangeTicketParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param ticket_id
	if err := r.SetPathParam("ticket_id", o.TicketID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
