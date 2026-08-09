package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/nekoimi/go-project-template/internal/framework"
)

func Run(configPath string) error {
	return run(configPath, framework.ScopeHTTP)
}

func RunScheduler(configPath string) error {
	return run(configPath, framework.ScopeScheduler)
}

func RunWorker(configPath string) error {
	return run(configPath, framework.ScopeWorker)
}

func run(configPath string, scopes ...framework.Scope) error {
	a, cleanup, err := initialize(configPath, scopes...)
	if err != nil {
		return err
	}
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := a.Start(ctx); err != nil {
		return err
	}

	var runtimeErr error
	select {
	case <-ctx.Done():
	case runtimeErr = <-a.runtimeErr:
	}
	return errors.Join(runtimeErr, a.Stop(context.Background()))
}

func (a *App) Start(ctx context.Context) error {
	if a == nil {
		return nil
	}
	if err := a.Boot(ctx); err != nil {
		return fmt.Errorf("boot app: %w", err)
	}
	if a.Worker != nil {
		if err := a.Worker.Start(); err != nil {
			startErr := fmt.Errorf("start task worker: %w", err)
			return errors.Join(startErr, a.Shutdown(ctx))
		}
	}
	if a.Scheduler != nil {
		a.Scheduler.Start()
	}
	if a.Engine != nil {
		a.startHTTPServer()
	}
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a == nil {
		return nil
	}
	a.Logger.Info("shutting down app")

	timeout := time.Duration(a.Config.Server.ShutdownTimeout) * time.Second
	if a.Config.TaskQueue.ShutdownTimeout > timeout {
		timeout = a.Config.TaskQueue.ShutdownTimeout
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var errs []error
	if a.HTTPServer != nil {
		if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown http server: %w", err))
		}
	}
	if a.Scheduler != nil {
		if err := a.Scheduler.Stop(shutdownCtx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown scheduler: %w", err))
		}
	}
	if a.Worker != nil {
		if err := stopWithContext(shutdownCtx, a.Worker.Shutdown); err != nil {
			errs = append(errs, fmt.Errorf("shutdown task worker: %w", err))
		}
	}
	if err := a.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, fmt.Errorf("shutdown app modules: %w", err))
	}

	shutdownErr := errors.Join(errs...)
	if shutdownErr != nil {
		a.Logger.Error("app stopped with shutdown errors", zap.Error(shutdownErr))
	} else {
		a.Logger.Info("app stopped")
	}
	return shutdownErr
}

func stopWithContext(ctx context.Context, stop func()) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		stop()
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *App) startHTTPServer() {
	a.HTTPServer = &http.Server{
		Addr:              ":" + a.Config.Server.Port,
		Handler:           a.Engine,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		a.Logger.Info("http server starting", zap.String("addr", a.HTTPServer.Addr))
		if err := a.HTTPServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.reportRuntimeError(fmt.Errorf("http server failed: %w", err))
		}
	}()
}

func (a *App) reportRuntimeError(err error) {
	if err == nil {
		return
	}
	a.Logger.Error("runtime failed", zap.Error(err))
	select {
	case a.runtimeErr <- err:
	default:
	}
}
