package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"superhue/internal/automation"
	"superhue/internal/domain"
	"superhue/internal/infrastructure"
	"superhue/internal/services"
)

type App struct {
	ctx            context.Context
	cancel         context.CancelFunc
	store          *infrastructure.SQLiteStore
	service        *services.AppService
	logger         *services.AppLogger
	automation     *automation.Engine
	processMonitor *infrastructure.ProcessMonitor
	networkMonitor *infrastructure.NetworkMonitor
}

func NewApp() *App { return &App{} }

func (a *App) Startup(ctx context.Context) {
	a.ctx, a.cancel = context.WithCancel(ctx)
	dataDir, err := appDataDir()
	if err != nil {
		runtime.LogErrorf(ctx, "app data dir error: %v", err)
		return
	}
	store, err := infrastructure.NewSQLiteStore(dataDir)
	if err != nil {
		runtime.LogErrorf(ctx, "sqlite error: %v", err)
		return
	}
	a.store = store
	settingsRepo := infrastructure.NewSettingsRepository(store.DB)
	hueRepo := infrastructure.NewHueRepository(store.DB)
	ruleRepo := infrastructure.NewRuleRepository(store.DB)
	deviceRepo := infrastructure.NewDeviceRepository(store.DB)
	logRepo := infrastructure.NewLogRepository(store.DB)
	logger := services.NewAppLogger(logRepo)
	hueClient := infrastructure.NewHueClient(settingsRepo)
	service := services.NewAppService(settingsRepo, hueRepo, ruleRepo, deviceRepo, logRepo, hueClient, logger)
	a.logger = logger
	a.service = service
	a.automation = automation.NewEngine(ruleRepo, service, logger)
	a.processMonitor = infrastructure.NewProcessMonitor(2 * time.Second)
	a.networkMonitor = infrastructure.NewNetworkMonitor(deviceRepo, 15*time.Second, 3)
	a.automation.Start(a.ctx)
	go a.processMonitor.Start(a.ctx, a.automation.Events())
	go a.networkMonitor.Start(a.ctx, a.automation.Events())
	logger.Info(a.ctx, "system", "Super HUE inició correctamente")
}

func (a *App) Shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) Bootstrap() (domain.AppState, error) { return a.service.Bootstrap(a.ctx) }

func (a *App) ConnectBridge(ip string) (domain.AppState, error) {
	return a.service.ConnectBridge(a.ctx, ip)
}

func (a *App) RefreshHue() (domain.AppState, error) {
	if err := a.service.RefreshHue(a.ctx); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) SetLightPower(lightID string, on bool) (domain.AppState, error) {
	if err := a.service.SetLightPower(a.ctx, lightID, on); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) SetLightBrightness(lightID string, brightness int) (domain.AppState, error) {
	if err := a.service.SetBrightness(a.ctx, lightID, brightness); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) SetLightColor(lightID string, hex string) (domain.AppState, error) {
	if err := a.service.SetColorHex(a.ctx, lightID, hex); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) ActivateScene(sceneID string) (domain.AppState, error) {
	if err := a.service.ActivateScene(a.ctx, sceneID); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) SaveRule(rule domain.Rule) (domain.AppState, error) {
	if err := a.service.SaveRule(a.ctx, rule); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) DeleteRule(id int64) (domain.AppState, error) {
	if err := a.service.DeleteRule(a.ctx, id); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) SaveDevice(device domain.Device) (domain.AppState, error) {
	if err := a.service.SaveDevice(a.ctx, device); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) DeleteDevice(id int64) (domain.AppState, error) {
	if err := a.service.DeleteDevice(a.ctx, id); err != nil {
		return domain.AppState{}, err
	}
	return a.service.Bootstrap(a.ctx)
}

func (a *App) PingAutomationNow() string {
	go func() {
		a.automation.Events() <- domain.RuleEvent{Trigger: domain.TriggerTimeSchedule, Name: "manual", Value: time.Now().Format("15:04")}
	}()
	return "ok"
}

func appDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "SuperHUE"), nil
}

func (a *App) DataPath() string {
	path, err := appDataDir()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return path
}
