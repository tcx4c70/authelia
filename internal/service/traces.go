package service

import "github.com/sirupsen/logrus"

func ProvisionTraces(ctx Context) (service Provider, err error) {
	providers := ctx.GetProviders()
	if providers.Telemetry == nil {
		return
	}

	return &Traces{
		name: "traces",
		ctx:  ctx,
		log:  ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeTraces, serviceTypeTraces: "traces"}),
	}, nil
}

type Traces struct {
	name string
	ctx  Context
	log  *logrus.Entry
}

func (service *Traces) ServiceType() string {
	return serviceTypeTraces
}

func (service *Traces) ServiceName() string {
	return service.name
}

func (service *Traces) Run() (err error) {
	providers := service.ctx.GetProviders()
	if providers.Telemetry == nil {
		return
	}

	return providers.Telemetry.Start(service.ctx)
}

func (service *Traces) Shutdown() {
	providers := service.ctx.GetProviders()
	if providers.Telemetry == nil {
		return
	}

	providers.Telemetry.Stop(service.ctx)
}

func (service *Traces) Log() *logrus.Entry {
	return service.log
}
