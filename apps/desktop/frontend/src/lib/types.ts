export type TriggerType = 'PROCESS_START' | 'PROCESS_STOP' | 'TIME_SCHEDULE' | 'NETWORK_PRESENCE';
export type ActionType = 'TURN_ON_LIGHT' | 'TURN_OFF_LIGHT' | 'SET_BRIGHTNESS' | 'SET_COLOR' | 'ACTIVATE_SCENE' | 'TURN_OFF_ALL';

export interface ConnectionStatus {
  bridgeIp: string;
  connected: boolean;
  bridgeReady: boolean;
  message: string;
}

export interface Light {
  id: string;
  name: string;
  roomId: string;
  roomName: string;
  on: boolean;
  brightness: number;
  colorHex: string;
  reachable: boolean;
}

export interface Room {
  id: string;
  name: string;
  type: string;
}

export interface Scene {
  id: string;
  name: string;
  groupId: string;
  group: string;
}

export interface Condition {
  id?: number;
  ruleId?: number;
  type: 'PROCESS_NAME' | 'SCHEDULE_AT' | 'DEVICE_STATE';
  key: string;
  value: string;
  negate?: boolean;
}

export interface Action {
  id?: number;
  ruleId?: number;
  type: ActionType;
  target: string;
  value: string;
}

export interface Rule {
  id?: number;
  name: string;
  trigger: TriggerType;
  enabled: boolean;
  conditions: Condition[];
  actions: Action[];
}

export interface Device {
  id?: number;
  name: string;
  ip: string;
  present: boolean;
  failureCount: number;
  consecutiveOks: number;
  lastSeenAt: string;
  lastCheckedAt: string;
}

export interface LogEntry {
  id: number;
  level: 'INFO' | 'WARN' | 'ERROR';
  source: string;
  message: string;
  createdAt: string;
}

export interface Dashboard {
  connectionStatus: ConnectionStatus;
  lightsCount: number;
  activeRules: number;
  devicesPresent: number;
  recentLogs: LogEntry[];
}

export interface AppState {
  dashboard: Dashboard;
  lights: Light[];
  rooms: Room[];
  scenes: Scene[];
  recentApps: string[];
  rules: Rule[];
  devices: Device[];
  logs: LogEntry[];
  settings: Record<string, string>;
}
