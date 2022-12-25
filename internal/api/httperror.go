package api

type HTTPError struct {
	Error string `json:"error"`
}

type HTTPMessageError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
