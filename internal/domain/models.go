package domain

import "time"

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConnectionStatus struct {
	BridgeIP    string `json:"bridgeIp"`
	Connected   bool   `json:"connected"`
	BridgeReady bool   `json:"bridgeReady"`
	Message     string `json:"message"`
}

type Light struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RoomID     string `json:"roomId"`
	RoomName   string `json:"roomName"`
	On         bool   `json:"on"`
	Brightness int    `json:"brightness"`
	ColorHex   string `json:"colorHex"`
	Reachable  bool   `json:"reachable"`
}

type Room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Scene struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	GroupID string `json:"groupId"`
	Group   string `json:"group"`
}

type Device struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	IP             string    `json:"ip"`
	Present        bool      `json:"present"`
	FailureCount   int       `json:"failureCount"`
	LastSeenAt     time.Time `json:"lastSeenAt"`
	LastCheckedAt  time.Time `json:"lastCheckedAt"`
	ConsecutiveOKs int       `json:"consecutiveOks"`
}

type TriggerType string

type ActionType string

type ConditionType string

const (
	TriggerProcessStart   TriggerType = "PROCESS_START"
	TriggerProcessStop    TriggerType = "PROCESS_STOP"
	TriggerTimeSchedule   TriggerType = "TIME_SCHEDULE"
	TriggerNetworkPresent TriggerType = "NETWORK_PRESENCE"
)

const (
	ActionTurnOnLight   ActionType = "TURN_ON_LIGHT"
	ActionTurnOffLight  ActionType = "TURN_OFF_LIGHT"
	ActionSetBrightness ActionType = "SET_BRIGHTNESS"
	ActionSetColor      ActionType = "SET_COLOR"
	ActionActivateScene ActionType = "ACTIVATE_SCENE"
	ActionTurnOffAll    ActionType = "TURN_OFF_ALL"
)

const (
	ConditionProcessName ConditionType = "PROCESS_NAME"
	ConditionScheduleAt  ConditionType = "SCHEDULE_AT"
	ConditionDeviceState ConditionType = "DEVICE_STATE"
)

type Rule struct {
	ID         int64       `json:"id"`
	Name       string      `json:"name"`
	Trigger    TriggerType `json:"trigger"`
	Enabled    bool        `json:"enabled"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
	Conditions []Condition `json:"conditions"`
	Actions    []Action    `json:"actions"`
}

type Condition struct {
	ID     int64         `json:"id"`
	RuleID int64         `json:"ruleId"`
	Type   ConditionType `json:"type"`
	Key    string        `json:"key"`
	Value  string        `json:"value"`
	Negate bool          `json:"negate"`
}

type Action struct {
	ID     int64      `json:"id"`
	RuleID int64      `json:"ruleId"`
	Type   ActionType `json:"type"`
	Target string     `json:"target"`
	Value  string     `json:"value"`
}

type LogLevel string

const (
	LogInfo  LogLevel = "INFO"
	LogWarn  LogLevel = "WARN"
	LogError LogLevel = "ERROR"
)

type LogEntry struct {
	ID        int64     `json:"id"`
	Level     LogLevel  `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type Dashboard struct {
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
	LightsCount      int              `json:"lightsCount"`
	ActiveRules      int              `json:"activeRules"`
	DevicesPresent   int              `json:"devicesPresent"`
	RecentLogs       []LogEntry       `json:"recentLogs"`
}

type AppState struct {
	Dashboard Dashboard         `json:"dashboard"`
	Lights    []Light           `json:"lights"`
	Rooms     []Room            `json:"rooms"`
	Scenes    []Scene           `json:"scenes"`
	Rules     []Rule            `json:"rules"`
	Devices   []Device          `json:"devices"`
	Logs      []LogEntry        `json:"logs"`
	Settings  map[string]string `json:"settings"`
}

type RuleEvent struct {
	Trigger TriggerType
	Name    string
	Value   string
}
