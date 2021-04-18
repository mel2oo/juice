package juice

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type App struct {
	opts   options
	ctx    context.Context
	cancel func()
}

func NewApp(opts ...Option) *App {
	options := options{
		ctx: context.Background(),
		// sigs: []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
	}

	for _, o := range opts {
		o(&options)
	}

	ctx, cancel := context.WithCancel(options.ctx)
	return &App{
		opts:   options,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (a *App) Run() error {
	g, ctx := errgroup.WithContext(a.ctx)
	for _, srv := range a.opts.servers {
		srv := srv

		g.Go(func() error {
			<-ctx.Done()
			return srv.Stop()
		})

		g.Go(func() error {
			return srv.Start()
		})
	}

	// c := make(chan os.Signal, 1)
	// signal.Notify(c, a.opts.sigs...)
	// g.Go(func() error {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return ctx.Err()
	// 		case <-c:
	// 			a.Stop()
	// 		}
	// 	}
	// })

	// if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
	// 	return err
	// }
	return nil
}

func (a *App) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}
