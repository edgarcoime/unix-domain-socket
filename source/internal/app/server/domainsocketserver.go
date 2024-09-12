package server

type DomainSocketServer struct{}

func NewDomainSocketServer() (*DomainSocketServer, error) {
	dss := &DomainSocketServer{}

	return dss, nil
}
