import type { AppState, Device, Rule } from './types';

declare global {
  interface Window {
    go?: {
      backend?: {
        App?: {
          Bootstrap(): Promise<AppState>;
          ConnectBridge(ip: string): Promise<AppState>;
          RefreshHue(): Promise<AppState>;
          SetLightPower(lightID: string, on: boolean): Promise<AppState>;
          SetLightBrightness(lightID: string, brightness: number): Promise<AppState>;
          SetLightColor(lightID: string, hex: string): Promise<AppState>;
          ActivateScene(sceneID: string): Promise<AppState>;
          SaveRule(rule: Rule): Promise<AppState>;
          DeleteRule(id: number): Promise<AppState>;
          SaveDevice(device: Device): Promise<AppState>;
          DeleteDevice(id: number): Promise<AppState>;
          SaveRoom(room: { id?: string; name: string; type: string }): Promise<AppState>;
          DeleteRoom(roomID: string): Promise<AppState>;
          AssignLightRoom(lightID: string, roomID: string): Promise<AppState>;
          ScanNetworkIPs(): Promise<string[]>;
          PingAutomationNow(): Promise<string>;
          DataPath(): Promise<string>;
        };
      };
    };
  }
}

const app = () => window.go?.backend?.App;

export const api = {
  bootstrap: () => app()?.Bootstrap() ?? Promise.reject(new Error('Wails bridge no disponible')),
  connectBridge: (ip: string) => app()?.ConnectBridge(ip) ?? Promise.reject(new Error('Wails bridge no disponible')),
  refreshHue: () => app()?.RefreshHue() ?? Promise.reject(new Error('Wails bridge no disponible')),
  setLightPower: (id: string, on: boolean) => app()?.SetLightPower(id, on) ?? Promise.reject(new Error('Wails bridge no disponible')),
  setLightBrightness: (id: string, brightness: number) => app()?.SetLightBrightness(id, brightness) ?? Promise.reject(new Error('Wails bridge no disponible')),
  setLightColor: (id: string, color: string) => app()?.SetLightColor(id, color) ?? Promise.reject(new Error('Wails bridge no disponible')),
  activateScene: (sceneID: string) => app()?.ActivateScene(sceneID) ?? Promise.reject(new Error('Wails bridge no disponible')),
  saveRule: (rule: Rule) => app()?.SaveRule(rule) ?? Promise.reject(new Error('Wails bridge no disponible')),
  deleteRule: (id: number) => app()?.DeleteRule(id) ?? Promise.reject(new Error('Wails bridge no disponible')),
  saveDevice: (device: Device) => app()?.SaveDevice(device) ?? Promise.reject(new Error('Wails bridge no disponible')),
  deleteDevice: (id: number) => app()?.DeleteDevice(id) ?? Promise.reject(new Error('Wails bridge no disponible')),
  saveRoom: (room: { id?: string; name: string; type: string }) => app()?.SaveRoom(room) ?? Promise.reject(new Error('Wails bridge no disponible')),
  deleteRoom: (roomID: string) => app()?.DeleteRoom(roomID) ?? Promise.reject(new Error('Wails bridge no disponible')),
  assignLightRoom: (lightID: string, roomID: string) => app()?.AssignLightRoom(lightID, roomID) ?? Promise.reject(new Error('Wails bridge no disponible')),
  scanNetworkIPs: () => app()?.ScanNetworkIPs() ?? Promise.reject(new Error('Wails bridge no disponible')),
  pingAutomationNow: () => app()?.PingAutomationNow() ?? Promise.reject(new Error('Wails bridge no disponible')),
  dataPath: () => app()?.DataPath() ?? Promise.reject(new Error('Wails bridge no disponible')),
};
