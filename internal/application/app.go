package application

import "context"

type App struct {
	publicServer *PublicServer
	container    *Container
}

func NewApp(publicServer *PublicServer, container *Container) *App {
	return &App{
		publicServer: publicServer,
		container:    container,
	}
}

func (a *App) Run(ctx context.Context) error {
	go func() {
		if err := a.publicServer.Start(); err != nil {
			a.container.Logger.ErrorContext(ctx, "public server stopped", "error", err)
		}
	}()
	return nil
}

func (a *App) ShutDown(ctx context.Context) error {
	if err := a.publicServer.ShutDown(ctx); err != nil {
		a.container.Logger.Error("failed to shutdown server", "error", err)
	}
	a.container.Close()
	return nil
}
