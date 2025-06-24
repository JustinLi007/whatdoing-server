package server

type Endpoint struct {
	Name string
	Path string
}

var endpointsInstance *Endpoints

type Endpoints struct {
	table map[string]Endpoint
}

func NewEndpoints() *Endpoints {
	if endpointsInstance != nil {
		return endpointsInstance
	}

	endpointsInstance := &Endpoints{
		table: make(map[string]Endpoint),
	}

	endpointsInstance.table["landing"] = Endpoint{
		Name: "Landing",
		Path: "/",
	}

	endpointsInstance.table["home"] = Endpoint{
		Name: "Home",
		Path: "/home",
	}

	endpointsInstance.table["signup"] = Endpoint{
		Name: "Sign Up",
		Path: "/signup",
	}

	endpointsInstance.table["login"] = Endpoint{
		Name: "Login",
		Path: "/login",
	}

	return endpointsInstance
}
