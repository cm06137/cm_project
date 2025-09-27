package starter

import (
	"github.com/cm06137/cm_project/common/container"
	"github.com/cm06137/cm_project/mcp_server/domain"
)

type Constructor struct {
}

func (c *Constructor) Name() string {
	return "mcp_server"
}

func (c *Constructor) ProvideConstructors(r container.IModuleRegister) {
	provideDomain(r)
}

func (c *Constructor) AfterInstantiation(sc container.IScopeContainer) {

}

func provideDomain(r container.IModuleRegister) {
	r.Provide(domain.NewMcpServer)
}
