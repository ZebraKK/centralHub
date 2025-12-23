package service

type DNSService struct {
}

func NewDNSService() *DNSService {
	return &DNSService{}
}

func (ds *DNSService) Create(domainName string, owner string) error {

	// For now, we will just return nil to indicate success.
	return nil
}
