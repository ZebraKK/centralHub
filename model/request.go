package model

type AddDomainRequest struct {
	Domain XLDomain `json:"domain"`
}

type AddDomainResponse struct {
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
