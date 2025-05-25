package types

type HTTPMethod string

const (
	HTTP_GET     HTTPMethod = "GET"
	HTTP_POST    HTTPMethod = "POST"
	HTTP_PUT     HTTPMethod = "PUT"
	HTTP_DELETE  HTTPMethod = "DELETE"
	HTTP_PATCH   HTTPMethod = "PATCH"
	HTTP_OPTIONS HTTPMethod = "OPTIONS"
	HTTP_HEAD    HTTPMethod = "HEAD"
)
