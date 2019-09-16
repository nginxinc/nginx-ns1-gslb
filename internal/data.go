package internal

// FeedData are the statistics fetched from NGINX Plus API in a format that NS1 API understands
type FeedData struct {
	Connections uint64 `json:"connections,omitempty"`
	Up          bool   `json:"up"`
}
