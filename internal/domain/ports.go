package domain

import "context"

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	List(ctx context.Context) (map[string]string, error)
}

type HueRepository interface {
	SaveLights(ctx context.Context, lights []Light) error
	SaveRooms(ctx context.Context, rooms []Room) error
	SaveScenes(ctx context.Context, scenes []Scene) error
	ListLights(ctx context.Context) ([]Light, error)
	ListRooms(ctx context.Context) ([]Room, error)
	ListScenes(ctx context.Context) ([]Scene, error)
}

type RuleRepository interface {
	List(ctx context.Context) ([]Rule, error)
	Save(ctx context.Context, rule *Rule) error
	Delete(ctx context.Context, id int64) error
}

type DeviceRepository interface {
	List(ctx context.Context) ([]Device, error)
	Save(ctx context.Context, device *Device) error
	Delete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, device Device) error
}

type LogRepository interface {
	Add(ctx context.Context, entry *LogEntry) error
	List(ctx context.Context, limit int) ([]LogEntry, error)
}

type HueClient interface {
	Connect(ctx context.Context, bridgeIP string) (ConnectionStatus, error)
	GetLights(ctx context.Context) ([]Light, error)
	GetRooms(ctx context.Context) ([]Room, error)
	GetScenes(ctx context.Context) ([]Scene, error)
	SetLightPower(ctx context.Context, lightID string, on bool) error
	SetBrightness(ctx context.Context, lightID string, brightness int) error
	SetColorHex(ctx context.Context, lightID string, hex string) error
	ActivateScene(ctx context.Context, sceneID string) error
	TurnOffAll(ctx context.Context) error
	Status(ctx context.Context) (ConnectionStatus, error)
}
