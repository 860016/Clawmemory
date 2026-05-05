//go:build !pro

package services

type ProAPIClient struct{}

func NewProAPIClient(baseURL, licenseKey string) *ProAPIClient {
	return nil
}
