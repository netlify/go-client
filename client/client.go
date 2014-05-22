package client

type method string

var (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

func Request(method method, path string, token string, headers map[string]string) (string, error) {
	return "", error
}
