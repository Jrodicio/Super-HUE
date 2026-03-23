package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"superhue/internal/domain"
)

type AppService struct {
	settings domain.SettingsRepository
	hueRepo  domain.HueRepository
	ruleRepo domain.RuleRepository
	devRepo  domain.DeviceRepository
	logRepo  domain.LogRepository
	hue      domain.HueClient
	logger   *AppLogger
}

func NewAppService(settings domain.SettingsRepository, hueRepo domain.HueRepository, ruleRepo domain.RuleRepository, devRepo domain.DeviceRepository, logRepo domain.LogRepository, hue domain.HueClient, logger *AppLogger) *AppService {
	return &AppService{settings: settings, hueRepo: hueRepo, ruleRepo: ruleRepo, devRepo: devRepo, logRepo: logRepo, hue: hue, logger: logger}
}

func (s *AppService) Bootstrap(ctx context.Context) (domain.AppState, error) {
	settings, _ := s.settings.List(ctx)
	status, _ := s.hue.Status(ctx)
	lights, _ := s.hueRepo.ListLights(ctx)
	rooms, _ := s.hueRepo.ListRooms(ctx)
	scenes, _ := s.hueRepo.ListScenes(ctx)
	rules, _ := s.ruleRepo.List(ctx)
	devices, _ := s.devRepo.List(ctx)
	logs, _ := s.logRepo.List(ctx, 25)
	activeRules := 0
	devicesPresent := 0
	for _, rule := range rules {
		if rule.Enabled {
			activeRules++
		}
	}
	for _, device := range devices {
		if device.Present {
			devicesPresent++
		}
	}
	return domain.AppState{
		Dashboard: domain.Dashboard{ConnectionStatus: status, LightsCount: len(lights), ActiveRules: activeRules, DevicesPresent: devicesPresent, RecentLogs: logs},
		Lights:    lights, Rooms: rooms, Scenes: scenes, Rules: rules, Devices: devices, Logs: logs, Settings: settings,
	}, nil
}

func (s *AppService) ConnectBridge(ctx context.Context, bridgeIP string) (domain.AppState, error) {
	status, err := s.hue.Connect(ctx, bridgeIP)
	if err != nil {
		s.logger.Error(ctx, "hue", err.Error())
		return domain.AppState{}, err
	}
	s.logger.Info(ctx, "hue", status.Message)
	if err := s.RefreshHue(ctx); err != nil {
		s.logger.Error(ctx, "hue", fmt.Sprintf("conectó pero no pudo sincronizar: %v", err))
	}
	return s.Bootstrap(ctx)
}

func (s *AppService) RefreshHue(ctx context.Context) error {
	lights, err := s.hue.GetLights(ctx)
	if err != nil {
		return err
	}
	rooms, err := s.hue.GetRooms(ctx)
	if err != nil {
		return err
	}
	scenes, err := s.hue.GetScenes(ctx)
	if err != nil {
		return err
	}
	if err := s.hueRepo.SaveRooms(ctx, rooms); err != nil {
		return err
	}
	if err := s.hueRepo.SaveLights(ctx, lights); err != nil {
		return err
	}
	if err := s.hueRepo.SaveScenes(ctx, scenes); err != nil {
		return err
	}
	s.logger.Info(ctx, "hue", fmt.Sprintf("Sincronización completada: %d luces, %d escenas", len(lights), len(scenes)))
	return nil
}

func (s *AppService) SetLightPower(ctx context.Context, lightID string, on bool) error {
	if err := s.hue.SetLightPower(ctx, lightID, on); err != nil {
		return err
	}
	s.logger.Info(ctx, "lights", fmt.Sprintf("Luz %s -> on=%v", lightID, on))
	return s.RefreshHue(ctx)
}

func (s *AppService) SetBrightness(ctx context.Context, lightID string, brightness int) error {
	if err := s.hue.SetBrightness(ctx, lightID, brightness); err != nil {
		return err
	}
	s.logger.Info(ctx, "lights", fmt.Sprintf("Luz %s brillo=%d", lightID, brightness))
	return s.RefreshHue(ctx)
}

func (s *AppService) SetColorHex(ctx context.Context, lightID, hex string) error {
	if err := s.hue.SetColorHex(ctx, lightID, hex); err != nil {
		return err
	}
	s.logger.Info(ctx, "lights", fmt.Sprintf("Luz %s color=%s", lightID, hex))
	return s.RefreshHue(ctx)
}

func (s *AppService) ActivateScene(ctx context.Context, sceneID string) error {
	if err := s.hue.ActivateScene(ctx, sceneID); err != nil {
		return err
	}
	s.logger.Info(ctx, "scenes", fmt.Sprintf("Escena activada %s", sceneID))
	return s.RefreshHue(ctx)
}

func (s *AppService) SaveRule(ctx context.Context, rule domain.Rule) error {
	if strings.TrimSpace(rule.Name) == "" {
		return errors.New("el nombre de la regla es obligatorio")
	}
	if len(rule.Actions) == 0 {
		return errors.New("la regla debe tener al menos una acción")
	}
	if err := s.ruleRepo.Save(ctx, &rule); err != nil {
		return err
	}
	s.logger.Info(ctx, "automation", fmt.Sprintf("Regla guardada: %s", rule.Name))
	return nil
}

func (s *AppService) DeleteRule(ctx context.Context, id int64) error {
	if err := s.ruleRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info(ctx, "automation", fmt.Sprintf("Regla eliminada: %d", id))
	return nil
}

func (s *AppService) SaveDevice(ctx context.Context, device domain.Device) error {
	if strings.TrimSpace(device.Name) == "" || strings.TrimSpace(device.IP) == "" {
		return errors.New("nombre e IP del dispositivo son obligatorios")
	}
	if err := s.devRepo.Save(ctx, &device); err != nil {
		return err
	}
	s.logger.Info(ctx, "network", fmt.Sprintf("Dispositivo guardado: %s (%s)", device.Name, device.IP))
	return nil
}

func (s *AppService) DeleteDevice(ctx context.Context, id int64) error {
	if err := s.devRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info(ctx, "network", fmt.Sprintf("Dispositivo eliminado: %d", id))
	return nil
}

func (s *AppService) ExecuteActions(ctx context.Context, actions []domain.Action) error {
	for _, action := range actions {
		switch action.Type {
		case domain.ActionTurnOnLight:
			if err := s.hue.SetLightPower(ctx, action.Target, true); err != nil {
				return err
			}
		case domain.ActionTurnOffLight:
			if err := s.hue.SetLightPower(ctx, action.Target, false); err != nil {
				return err
			}
		case domain.ActionSetBrightness:
			var brightness int
			fmt.Sscanf(action.Value, "%d", &brightness)
			if err := s.hue.SetBrightness(ctx, action.Target, brightness); err != nil {
				return err
			}
		case domain.ActionSetColor:
			if err := s.hue.SetColorHex(ctx, action.Target, action.Value); err != nil {
				return err
			}
		case domain.ActionActivateScene:
			if err := s.hue.ActivateScene(ctx, action.Target); err != nil {
				return err
			}
		case domain.ActionTurnOffAll:
			if err := s.hue.TurnOffAll(ctx); err != nil {
				return err
			}
		default:
			return fmt.Errorf("acción no soportada: %s", action.Type)
		}
	}
	return s.RefreshHue(ctx)
}
