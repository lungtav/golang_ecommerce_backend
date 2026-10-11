package auth

type handler struct{
	service Service
}

func AuthHandler (service Service) *handler{
	return &handler{service: service}
}

func (h *handler) RegisterUser() error{
	return nil
}