package apimodels

type (
	// RoutesResponse defines the response of a GET routes REST API call.
	RoutesResponse struct {
		Routes []string `json:"routes"`
	}
)
