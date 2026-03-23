package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"superhue/internal/domain"
)

type HueClient struct {
	httpClient *http.Client
	settings   domain.SettingsRepository
	mu         sync.RWMutex
	bridgeIP   string
	username   string
}

func NewHueClient(settings domain.SettingsRepository) *HueClient {
	return &HueClient{
		httpClient: &http.Client{Timeout: 8 * time.Second},
		settings:   settings,
	}
}

func (c *HueClient) Connect(ctx context.Context, bridgeIP string) (domain.ConnectionStatus, error) {
	bridgeIP = strings.TrimSpace(bridgeIP)
	if bridgeIP == "" {
		return domain.ConnectionStatus{}, errors.New("la IP del bridge es obligatoria")
	}
	storedIP, _ := c.settings.Get(ctx, "bridge_ip")
	username, _ := c.settings.Get(ctx, "hue_username")
	if username == "" || storedIP != bridgeIP {
		created, err := c.createUser(ctx, bridgeIP)
		if err != nil {
			return domain.ConnectionStatus{BridgeIP: bridgeIP, Connected: false, BridgeReady: false, Message: err.Error()}, err
		}
		username = created
		if err := c.settings.Set(ctx, "hue_username", username); err != nil {
			return domain.ConnectionStatus{}, err
		}
	}
	if err := c.settings.Set(ctx, "bridge_ip", bridgeIP); err != nil {
		return domain.ConnectionStatus{}, err
	}
	c.mu.Lock()
	c.bridgeIP = bridgeIP
	c.username = username
	c.mu.Unlock()
	status, err := c.Status(ctx)
	if err != nil {
		return status, err
	}
	status.Message = "Bridge conectado correctamente"
	return status, nil
}

func (c *HueClient) createUser(ctx context.Context, bridgeIP string) (string, error) {
	payload := map[string]string{"devicetype": "super_hue_desktop#local"}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api", bridgeIP), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo contactar el Hue Bridge: %w", err)
	}
	defer resp.Body.Close()
	var result []map[string]map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("respuesta inválida del bridge: %w", err)
	}
	if len(result) == 0 {
		return "", errors.New("respuesta vacía del bridge")
	}
	if errInfo, ok := result[0]["error"]; ok {
		return "", fmt.Errorf("bridge rechazó el registro: %s (presioná el botón físico del bridge e intentá de nuevo)", errInfo["description"])
	}
	if success, ok := result[0]["success"]; ok {
		if username, exists := success["username"]; exists {
			return username, nil
		}
	}
	return "", errors.New("no se pudo crear la credencial local del bridge")
}

func (c *HueClient) Status(ctx context.Context) (domain.ConnectionStatus, error) {
	bridgeIP, username, err := c.loadCredentials(ctx)
	if err != nil {
		return domain.ConnectionStatus{}, err
	}
	if bridgeIP == "" || username == "" {
		return domain.ConnectionStatus{BridgeIP: bridgeIP, Connected: false, BridgeReady: false, Message: "Configurá la IP del Hue Bridge para iniciar"}, nil
	}
	var config struct {
		Name string `json:"name"`
	}
	if err := c.get(ctx, "/config", &config); err != nil {
		return domain.ConnectionStatus{BridgeIP: bridgeIP, Connected: false, BridgeReady: false, Message: err.Error()}, nil
	}
	return domain.ConnectionStatus{BridgeIP: bridgeIP, Connected: true, BridgeReady: true, Message: fmt.Sprintf("Conectado a %s", config.Name)}, nil
}

func (c *HueClient) GetLights(ctx context.Context) ([]domain.Light, error) {
	var payload map[string]struct {
		Name  string `json:"name"`
		State struct {
			On        bool `json:"on"`
			Bri       int  `json:"bri"`
			Reachable bool `json:"reachable"`
			Hue       int  `json:"hue"`
			Sat       int  `json:"sat"`
		} `json:"state"`
	}
	if err := c.get(ctx, "/lights", &payload); err != nil {
		return nil, err
	}
	rooms, _ := c.GetRooms(ctx)
	lightRoom := map[string]string{}
	roomNames := map[string]string{}
	for _, room := range rooms {
		roomNames[room.ID] = room.Name
	}
	var groups map[string]struct {
		Name   string   `json:"name"`
		Type   string   `json:"type"`
		Lights []string `json:"lights"`
	}
	_ = c.get(ctx, "/groups", &groups)
	for id, group := range groups {
		if group.Type != "Room" && group.Type != "Zone" {
			continue
		}
		for _, lightID := range group.Lights {
			lightRoom[lightID] = id
		}
	}
	lights := make([]domain.Light, 0, len(payload))
	for id, light := range payload {
		roomID := lightRoom[id]
		lights = append(lights, domain.Light{
			ID:         id,
			Name:       light.Name,
			RoomID:     roomID,
			RoomName:   roomNames[roomID],
			On:         light.State.On,
			Brightness: int(math.Round(float64(light.State.Bri) / 254 * 100)),
			ColorHex:   hueSatToHex(light.State.Hue, light.State.Sat),
			Reachable:  light.State.Reachable,
		})
	}
	return lights, nil
}

func (c *HueClient) GetRooms(ctx context.Context) ([]domain.Room, error) {
	var payload map[string]struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := c.get(ctx, "/groups", &payload); err != nil {
		return nil, err
	}
	rooms := []domain.Room{}
	for id, room := range payload {
		if room.Type != "Room" && room.Type != "Zone" {
			continue
		}
		rooms = append(rooms, domain.Room{ID: id, Name: room.Name, Type: room.Type})
	}
	return rooms, nil
}

func (c *HueClient) GetScenes(ctx context.Context) ([]domain.Scene, error) {
	var payload map[string]struct {
		Name  string `json:"name"`
		Group string `json:"group"`
	}
	if err := c.get(ctx, "/scenes", &payload); err != nil {
		return nil, err
	}
	rooms, _ := c.GetRooms(ctx)
	roomNames := map[string]string{}
	for _, room := range rooms {
		roomNames[room.ID] = room.Name
	}
	scenes := make([]domain.Scene, 0, len(payload))
	for id, scene := range payload {
		scenes = append(scenes, domain.Scene{ID: id, Name: scene.Name, GroupID: scene.Group, Group: roomNames[scene.Group]})
	}
	return scenes, nil
}

func (c *HueClient) SetLightPower(ctx context.Context, lightID string, on bool) error {
	return c.putState(ctx, fmt.Sprintf("/lights/%s/state", lightID), map[string]any{"on": on})
}

func (c *HueClient) SetBrightness(ctx context.Context, lightID string, brightness int) error {
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 100 {
		brightness = 100
	}
	bri := int(math.Round(float64(brightness) / 100 * 254))
	return c.putState(ctx, fmt.Sprintf("/lights/%s/state", lightID), map[string]any{"on": brightness > 0, "bri": bri})
}

func (c *HueClient) SetColorHex(ctx context.Context, lightID string, hex string) error {
	hue, sat := hexToHueSat(hex)
	return c.putState(ctx, fmt.Sprintf("/lights/%s/state", lightID), map[string]any{"on": true, "hue": hue, "sat": sat})
}

func (c *HueClient) ActivateScene(ctx context.Context, sceneID string) error {
	scenes, err := c.GetScenes(ctx)
	if err != nil {
		return err
	}
	for _, scene := range scenes {
		if scene.ID == sceneID {
			if scene.GroupID == "" {
				continue
			}
			return c.putState(ctx, fmt.Sprintf("/groups/%s/action", scene.GroupID), map[string]any{"scene": sceneID})
		}
	}
	return fmt.Errorf("escena %s no encontrada", sceneID)
}

func (c *HueClient) TurnOffAll(ctx context.Context) error {
	rooms, err := c.GetRooms(ctx)
	if err != nil {
		return err
	}
	for _, room := range rooms {
		if err := c.putState(ctx, fmt.Sprintf("/groups/%s/action", room.ID), map[string]any{"on": false}); err != nil {
			return err
		}
	}
	return nil
}

func (c *HueClient) putState(ctx context.Context, path string, payload map[string]any) error {
	_, err := c.request(ctx, http.MethodPut, path, payload)
	return err
}

func (c *HueClient) get(ctx context.Context, path string, target any) error {
	body, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode hue payload %s: %w", path, err)
	}
	return nil
}

func (c *HueClient) request(ctx context.Context, method, path string, payload any) ([]byte, error) {
	bridgeIP, username, err := c.loadCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if bridgeIP == "" || username == "" {
		return nil, errors.New("bridge no configurado")
	}
	url := fmt.Sprintf("http://%s/api/%s%s", bridgeIP, username, path)
	var body *bytes.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	} else {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hue request: %w", err)
	}
	defer resp.Body.Close()
	var response bytes.Buffer
	if _, err := response.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hue status %d: %s", resp.StatusCode, response.String())
	}
	return response.Bytes(), nil
}

func (c *HueClient) loadCredentials(ctx context.Context) (string, string, error) {
	c.mu.RLock()
	bridgeIP, username := c.bridgeIP, c.username
	c.mu.RUnlock()
	if bridgeIP != "" && username != "" {
		return bridgeIP, username, nil
	}
	var err error
	if bridgeIP == "" {
		bridgeIP, err = c.settings.Get(ctx, "bridge_ip")
		if err != nil {
			return "", "", err
		}
	}
	if username == "" {
		username, err = c.settings.Get(ctx, "hue_username")
		if err != nil {
			return "", "", err
		}
	}
	c.mu.Lock()
	c.bridgeIP, c.username = bridgeIP, username
	c.mu.Unlock()
	return bridgeIP, username, nil
}

func (c *HueClient) currentBridgeIP() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.bridgeIP
}

func hueSatToHex(hue, sat int) string {
	h := float64(hue) / 65535 * 360
	s := float64(sat) / 254
	v := 1.0
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return fmt.Sprintf("#%02X%02X%02X", int((r+m)*255), int((g+m)*255), int((b+m)*255))
}

func hexToHueSat(hex string) (int, int) {
	hex = strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(hex) != 6 {
		return 0, 0
	}
	rv, _ := strconv.ParseInt(hex[0:2], 16, 64)
	gv, _ := strconv.ParseInt(hex[2:4], 16, 64)
	bv, _ := strconv.ParseInt(hex[4:6], 16, 64)
	r, g, b := float64(rv)/255, float64(gv)/255, float64(bv)/255
	maxV := math.Max(r, math.Max(g, b))
	minV := math.Min(r, math.Min(g, b))
	delta := maxV - minV
	var h float64
	s := 0.0
	if maxV != 0 {
		s = delta / maxV
	}
	switch {
	case delta == 0:
		h = 0
	case maxV == r:
		h = 60 * math.Mod(((g-b)/delta), 6)
	case maxV == g:
		h = 60 * (((b - r) / delta) + 2)
	default:
		h = 60 * (((r - g) / delta) + 4)
	}
	if h < 0 {
		h += 360
	}
	return int(h / 360 * 65535), int(s * 254)
}

var _ domain.HueClient = (*HueClient)(nil)
