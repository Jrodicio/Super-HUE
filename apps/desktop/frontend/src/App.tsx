import { useEffect, useMemo, useState } from 'react';
import { SectionCard } from './components/SectionCard';
import { api } from './lib/wails';
import type { ActionType, AppState, Device, LogEntry, Rule, TriggerType } from './lib/types';

type View = 'dashboard' | 'lights' | 'scenes' | 'automation' | 'settings';

const emptyState: AppState = {
  dashboard: {
    connectionStatus: { bridgeIp: '', connected: false, bridgeReady: false, message: 'Sin conexión' },
    lightsCount: 0,
    activeRules: 0,
    devicesPresent: 0,
    recentLogs: [],
  },
  lights: [], rooms: [], scenes: [], rules: [], devices: [], logs: [], settings: {},
};

const defaultRule = (): Rule => ({
  name: '',
  trigger: 'PROCESS_START',
  enabled: true,
  conditions: [{ type: 'PROCESS_NAME', key: 'process', value: 'cs2.exe' }],
  actions: [{ type: 'ACTIVATE_SCENE', target: '', value: '' }],
});

const defaultDevice: Device = { name: '', ip: '', present: false, failureCount: 0, consecutiveOks: 0, lastCheckedAt: '', lastSeenAt: '' };

export function App() {
  const [state, setState] = useState<AppState>(emptyState);
  const [view, setView] = useState<View>('dashboard');
  const [bridgeIP, setBridgeIP] = useState('');
  const [ruleDraft, setRuleDraft] = useState<Rule>(defaultRule());
  const [deviceDraft, setDeviceDraft] = useState<Device>(defaultDevice);
  const [busy, setBusy] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [notice, setNotice] = useState<string>('');

  useEffect(() => {
    api.bootstrap().then((data) => {
      const nextState = normalizeAppState(data);
      setState(nextState);
      setBridgeIP(nextState.settings.bridge_ip ?? nextState.dashboard.connectionStatus.bridgeIp ?? '');
    }).catch((err) => setError(err.message));
  }, []);

  const applyState = async (promise: Promise<AppState>, success?: string) => {
    setError('');
    if (success) setNotice('');
    try {
      const data = normalizeAppState(await promise);
      setState(data);
      setBridgeIP(data.settings.bridge_ip ?? bridgeIP);
      if (success) setNotice(success);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error desconocido');
    } finally {
      setBusy('');
    }
  };

  const sortedLights = useMemo(() => {
    return [...state.lights].sort((a, b) => a.name.localeCompare(b.name));
  }, [state.lights]);

  const saveRule = () => {
    setBusy('save-rule');
    applyState(api.saveRule(ruleDraft), 'Regla guardada');
    setRuleDraft(defaultRule());
  };

  const saveDevice = () => {
    setBusy('save-device');
    applyState(api.saveDevice(deviceDraft), 'Dispositivo guardado');
    setDeviceDraft(defaultDevice);
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div>
          <div className="brand">Super HUE</div>
          <p className="brand-copy">Automatización local para Philips Hue en Windows con Wails + Go.</p>
        </div>
        <nav className="nav-list">
          {[
            ['dashboard', 'Dashboard'],
            ['lights', 'Luces'],
            ['scenes', 'Escenas'],
            ['automation', 'Automatizaciones'],
            ['settings', 'Configuración'],
          ].map(([id, label]) => (
            <button key={id} className={view === id ? 'nav-btn active' : 'nav-btn'} onClick={() => setView(id as View)}>{label}</button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <span className={state.dashboard.connectionStatus.connected ? 'status-dot online' : 'status-dot offline'} />
          {state.dashboard.connectionStatus.message}
        </div>
      </aside>

      <main className="content">
        <header className="topbar">
          <div>
            <h1>{titleFor(view)}</h1>
            <p>{subtitleFor(view)}</p>
          </div>
          <div className="topbar-actions">
            <button className="ghost-btn" onClick={() => { setBusy('refresh'); applyState(api.refreshHue(), 'Sincronización actualizada'); }}>Refresh Hue</button>
            <button className="primary-btn" onClick={() => api.pingAutomationNow()}>Probar scheduler</button>
          </div>
        </header>

        {error ? <div className="alert error">{error}</div> : null}
        {notice ? <div className="alert success">{notice}</div> : null}

        {view === 'dashboard' && (
          <div className="grid cols-4">
            <MetricCard label="Bridge" value={state.dashboard.connectionStatus.connected ? 'Conectado' : 'Pendiente'} detail={state.dashboard.connectionStatus.bridgeIp || 'Ingresá la IP'} />
            <MetricCard label="Luces" value={String(state.dashboard.lightsCount)} detail="Sincronizadas localmente" />
            <MetricCard label="Reglas activas" value={String(state.dashboard.activeRules)} detail="Motor en ejecución" />
            <MetricCard label="Dispositivos presentes" value={String(state.dashboard.devicesPresent)} detail="Basado en ping local" />
            <SectionCard title="Resumen operativo" subtitle="Estado general de la automatización">
              <ul className="detail-list">
                <li>Bridge: {state.dashboard.connectionStatus.message}</li>
                <li>Escenas disponibles: {state.scenes.length}</li>
                <li>Dispositivos registrados: {state.devices.length}</li>
                <li>SQLite local: {state.settings.bridge_ip ? 'configurada' : 'lista para usar'}</li>
              </ul>
            </SectionCard>
            <SectionCard className="span-3" title="Últimos eventos" subtitle="Logs recientes del sistema">
              <LogsList logs={state.dashboard.recentLogs} />
            </SectionCard>
          </div>
        )}

        {view === 'lights' && (
          <SectionCard title="Control de luces" subtitle="On/off, brillo y color por luz">
            <div className="lights-grid">
              {sortedLights.map((light) => (
                <div
                  className="light-card"
                  key={light.id}
                  style={{ backgroundColor: cardBackgroundColor(light.colorHex, light.on, light.reachable) }}
                >
                  <div className="light-header">
                    <div>
                      <strong>{light.name}</strong>
                      <span>{light.roomName || 'Sin room'}</span>
                    </div>
                    <label className="switch">
                      <input type="checkbox" checked={light.on} onChange={(e) => { setBusy(`light-${light.id}`); applyState(api.setLightPower(light.id, e.target.checked)); }} />
                      <span />
                    </label>
                  </div>
                  <div className="field-row">
                    <label>Brillo</label>
                    <input type="range" min={0} max={100} value={light.brightness} onChange={(e) => setState((prev) => ({ ...prev, lights: prev.lights.map((item) => item.id === light.id ? { ...item, brightness: Number(e.target.value) } : item) }))} onMouseUp={() => { setBusy(`bri-${light.id}`); applyState(api.setLightBrightness(light.id, light.brightness)); }} onTouchEnd={() => { setBusy(`bri-${light.id}`); applyState(api.setLightBrightness(light.id, light.brightness)); }} />
                    <span>{light.brightness}%</span>
                  </div>
                  <div className="field-row">
                    <label>Color</label>
                    <input
                      type="color"
                      value={light.colorHex || '#FFFFFF'}
                      style={{ backgroundColor: light.colorHex || '#FFFFFF' }}
                      onChange={(e) => { setBusy(`color-${light.id}`); applyState(api.setLightColor(light.id, e.target.value)); }}
                    />
                    <span>{light.colorHex}</span>
                  </div>
                  <span className={light.reachable ? 'chip success availability-chip' : 'chip muted availability-chip'}>{light.reachable ? 'Disponible' : 'No reachable'}</span>
                </div>
              ))}
            </div>
          </SectionCard>
        )}

        {view === 'scenes' && (
          <SectionCard title="Escenas" subtitle="Ejecutá escenas guardadas del Hue Bridge">
            <div className="scene-grid">
              {state.scenes.map((scene) => (
                <div className="scene-card" key={scene.id}>
                  <strong>{scene.name}</strong>
                  <span>{scene.group || 'Escena global'}</span>
                  <button className="primary-btn" onClick={() => { setBusy(`scene-${scene.id}`); applyState(api.activateScene(scene.id), 'Escena ejecutada'); }}>Ejecutar</button>
                </div>
              ))}
            </div>
            <p className="muted-note">TODO: el guardado de escenas personalizadas se deja preparado para una siguiente iteración mediante persistencia local y captura del estado actual.</p>
          </SectionCard>
        )}

        {view === 'automation' && (
          <div className="stack">
            <SectionCard title="Reglas" subtitle="Procesos, horarios y presencia de red">
              <div className="two-col">
                <div className="form-grid">
                  <label>
                    Nombre
                    <input value={ruleDraft.name} onChange={(e) => setRuleDraft({ ...ruleDraft, name: e.target.value })} placeholder="Gaming mode" />
                  </label>
                  <label>
                    Trigger
                    <select value={ruleDraft.trigger} onChange={(e) => setRuleDraft(buildRuleForTrigger(e.target.value as TriggerType, ruleDraft))}>
                      <option value="PROCESS_START">PROCESS_START</option>
                      <option value="PROCESS_STOP">PROCESS_STOP</option>
                      <option value="TIME_SCHEDULE">TIME_SCHEDULE</option>
                      <option value="NETWORK_PRESENCE">NETWORK_PRESENCE</option>
                    </select>
                  </label>
                  <label>
                    Condición valor
                    <input value={ruleDraft.conditions[0]?.value ?? ''} onChange={(e) => setRuleDraft({ ...ruleDraft, conditions: [{ ...ruleDraft.conditions[0], value: e.target.value }] })} placeholder={placeholderForTrigger(ruleDraft.trigger)} />
                  </label>
                  {ruleDraft.trigger === 'NETWORK_PRESENCE' ? (
                    <label>
                      Dispositivo
                      <select value={ruleDraft.conditions[0]?.key ?? ''} onChange={(e) => setRuleDraft({ ...ruleDraft, conditions: [{ ...ruleDraft.conditions[0], key: e.target.value }] })}>
                        <option value="">Seleccionar</option>
                        {state.devices.map((device) => <option key={device.id} value={device.name}>{device.name}</option>)}
                      </select>
                    </label>
                  ) : null}
                  <label>
                    Acción
                    <select value={ruleDraft.actions[0]?.type ?? 'ACTIVATE_SCENE'} onChange={(e) => setRuleDraft(buildAction(e.target.value as ActionType, ruleDraft))}>
                      <option value="ACTIVATE_SCENE">ACTIVATE_SCENE</option>
                      <option value="TURN_ON_LIGHT">TURN_ON_LIGHT</option>
                      <option value="TURN_OFF_LIGHT">TURN_OFF_LIGHT</option>
                      <option value="SET_BRIGHTNESS">SET_BRIGHTNESS</option>
                      <option value="SET_COLOR">SET_COLOR</option>
                      <option value="TURN_OFF_ALL">TURN_OFF_ALL</option>
                    </select>
                  </label>
                  <label>
                    Objetivo
                    <select value={ruleDraft.actions[0]?.target ?? ''} onChange={(e) => setRuleDraft({ ...ruleDraft, actions: [{ ...ruleDraft.actions[0], target: e.target.value }] })}>
                      <option value="">Seleccionar</option>
                      {ruleDraft.actions[0]?.type === 'ACTIVATE_SCENE' && state.scenes.map((scene) => <option key={scene.id} value={scene.id}>{scene.name}</option>)}
                      {ruleDraft.actions[0]?.type !== 'ACTIVATE_SCENE' && ruleDraft.actions[0]?.type !== 'TURN_OFF_ALL' && state.lights.map((light) => <option key={light.id} value={light.id}>{light.name}</option>)}
                    </select>
                  </label>
                  {(ruleDraft.actions[0]?.type === 'SET_BRIGHTNESS' || ruleDraft.actions[0]?.type === 'SET_COLOR') ? (
                    <label>
                      Valor
                      <input value={ruleDraft.actions[0]?.value ?? ''} onChange={(e) => setRuleDraft({ ...ruleDraft, actions: [{ ...ruleDraft.actions[0], value: e.target.value }] })} placeholder={ruleDraft.actions[0]?.type === 'SET_BRIGHTNESS' ? '30' : '#FFAA55'} />
                    </label>
                  ) : null}
                  <button className="primary-btn" disabled={busy === 'save-rule'} onClick={saveRule}>Crear regla</button>
                </div>
                <div className="rules-table">
                  {state.rules.map((rule) => (
                    <div className="rule-row" key={rule.id}>
                      <div>
                        <strong>{rule.name}</strong>
                        <p>{rule.trigger} · {rule.conditions.map((item) => `${item.key || item.type}:${item.value}`).join(', ')}</p>
                      </div>
                      <div className="rule-actions">
                        <span className={rule.enabled ? 'chip success' : 'chip muted'}>{rule.enabled ? 'Activa' : 'Inactiva'}</span>
                        <button className="ghost-btn danger" onClick={() => { setBusy(`delete-rule-${rule.id}`); applyState(api.deleteRule(rule.id!), 'Regla eliminada'); }}>Eliminar</button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </SectionCard>

            <SectionCard title="Dispositivos en red" subtitle="Detección local por ping con tolerancia de 3 fallos">
              <div className="two-col">
                <div className="form-grid compact">
                  <label>
                    Nombre
                    <input value={deviceDraft.name} onChange={(e) => setDeviceDraft({ ...deviceDraft, name: e.target.value })} placeholder="CELULAR_A" />
                  </label>
                  <label>
                    IP local
                    <input value={deviceDraft.ip} onChange={(e) => setDeviceDraft({ ...deviceDraft, ip: e.target.value })} placeholder="192.168.1.50" />
                  </label>
                  <button className="primary-btn" disabled={busy === 'save-device'} onClick={saveDevice}>Guardar dispositivo</button>
                </div>
                <div className="device-list">
                  {state.devices.map((device) => (
                    <div className="rule-row" key={device.id}>
                      <div>
                        <strong>{device.name}</strong>
                        <p>{device.ip} · Último check: {formatDate(device.lastCheckedAt)}</p>
                      </div>
                      <div className="rule-actions">
                        <span className={device.present ? 'chip success' : 'chip muted'}>{device.present ? 'Presente' : 'Ausente'}</span>
                        <button className="ghost-btn danger" onClick={() => { setBusy(`delete-device-${device.id}`); applyState(api.deleteDevice(device.id!), 'Dispositivo eliminado'); }}>Eliminar</button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </SectionCard>
          </div>
        )}

        {view === 'settings' && (
          <div className="stack">
            <SectionCard title="Hue Bridge" subtitle="Conexión local al bridge por IP">
              <div className="settings-row">
                <label>
                  IP del Hue Bridge
                  <input value={bridgeIP} onChange={(e) => setBridgeIP(e.target.value)} placeholder="192.168.1.20" />
                </label>
                <button className="primary-btn" onClick={() => { setBusy('connect'); applyState(api.connectBridge(bridgeIP), 'Bridge conectado'); }}>Conectar</button>
              </div>
              <p className="muted-note">Tip: al conectar por primera vez debés presionar el botón físico del Hue Bridge para autorizar la credencial local.</p>
            </SectionCard>
            <SectionCard title="Persistencia local" subtitle="Dónde guarda la app su SQLite y configuración">
              <DataPath />
            </SectionCard>
            <SectionCard title="Logs" subtitle="Eventos recientes, errores y ejecuciones de reglas">
              <LogsList logs={state.logs} />
            </SectionCard>
          </div>
        )}
      </main>
    </div>
  );
}

function cardBackgroundColor(colorHex: string, isOn: boolean, isReachable: boolean) {
  if (!isOn || !isReachable) {
    return 'rgba(0, 0, 0, 0.35)';
  }

  return hexToRgba(colorHex || '#000000', 0.26);
}

function hexToRgba(colorHex: string, alpha: number) {
  const normalized = colorHex.replace('#', '');

  if (normalized.length !== 6) {
    return `rgba(0, 0, 0, ${alpha})`;
  }

  const r = Number.parseInt(normalized.slice(0, 2), 16);
  const g = Number.parseInt(normalized.slice(2, 4), 16);
  const b = Number.parseInt(normalized.slice(4, 6), 16);

  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}


function normalizeAppState(data: AppState): AppState {
  return {
    dashboard: {
      connectionStatus: data.dashboard?.connectionStatus ?? emptyState.dashboard.connectionStatus,
      lightsCount: data.dashboard?.lightsCount ?? 0,
      activeRules: data.dashboard?.activeRules ?? 0,
      devicesPresent: data.dashboard?.devicesPresent ?? 0,
      recentLogs: data.dashboard?.recentLogs ?? [],
    },
    lights: data.lights ?? [],
    rooms: data.rooms ?? [],
    scenes: data.scenes ?? [],
    rules: data.rules ?? [],
    devices: data.devices ?? [],
    logs: data.logs ?? [],
    settings: data.settings ?? {},
  };
}

function buildRuleForTrigger(trigger: TriggerType, rule: Rule): Rule {
  const base = { ...rule, trigger };
  switch (trigger) {
    case 'PROCESS_START':
    case 'PROCESS_STOP':
      return { ...base, conditions: [{ type: 'PROCESS_NAME', key: 'process', value: '' }] };
    case 'TIME_SCHEDULE':
      return { ...base, conditions: [{ type: 'SCHEDULE_AT', key: 'time', value: '20:00' }] };
    case 'NETWORK_PRESENCE':
      return { ...base, conditions: [{ type: 'DEVICE_STATE', key: '', value: 'present' }] };
  }
}

function buildAction(type: ActionType, rule: Rule): Rule {
  const value = type === 'SET_BRIGHTNESS' ? '30' : type === 'SET_COLOR' ? '#FFAA55' : '';
  return { ...rule, actions: [{ type, target: '', value }] };
}

function placeholderForTrigger(trigger: TriggerType) {
  switch (trigger) {
    case 'PROCESS_START':
    case 'PROCESS_STOP': return 'cs2.exe';
    case 'TIME_SCHEDULE': return '20:00';
    case 'NETWORK_PRESENCE': return 'present o absent';
  }
}

function titleFor(view: View) {
  return ({ dashboard: 'Dashboard', lights: 'Luces', scenes: 'Escenas', automation: 'Automatizaciones', settings: 'Configuración' })[view];
}

function subtitleFor(view: View) {
  return ({
    dashboard: 'Vista general del sistema local de Philips Hue.',
    lights: 'Control manual rápido y feedback inmediato.',
    scenes: 'Ejecución directa de escenas del bridge.',
    automation: 'Reglas basadas en procesos, horarios y presencia de red.',
    settings: 'Conexión, persistencia local y logs del sistema.',
  })[view];
}

function MetricCard({ label, value, detail }: { label: string; value: string; detail: string }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <p>{detail}</p>
    </div>
  );
}

function LogsList({ logs }: { logs: LogEntry[] | null | undefined }) {
  const safeLogs = logs ?? [];
  const itemsPerPage = 10;
  const [page, setPage] = useState(1);

  useEffect(() => {
    setPage(1);
  }, [safeLogs.length]);

  const totalPages = Math.max(1, Math.ceil(safeLogs.length / itemsPerPage));
  const currentPage = Math.min(page, totalPages);
  const pagedLogs = useMemo(() => {
    const start = (currentPage - 1) * itemsPerPage;
    return safeLogs.slice(start, start + itemsPerPage);
  }, [currentPage, safeLogs]);

  return (
    <div className="log-list-wrap">
      <div className="log-list">
        {safeLogs.length === 0 ? <span className="muted-note">Todavía no hay logs.</span> : pagedLogs.map((log) => (
          <div className="log-row" key={`${log.id}-${log.createdAt}`}>
            <span className={`chip ${log.level === 'ERROR' ? 'danger' : log.level === 'WARN' ? 'warn' : 'success'}`}>{log.level}</span>
            <div>
              <strong>{log.source}</strong>
              <p>{log.message}</p>
            </div>
            <time>{formatDate(log.createdAt)}</time>
          </div>
        ))}
      </div>
      {safeLogs.length > itemsPerPage ? (
        <div className="pagination">
          <button className="ghost-btn" disabled={currentPage === 1} onClick={() => setPage((prev) => Math.max(1, prev - 1))}>Anterior</button>
          <span>Página {currentPage} de {totalPages}</span>
          <button className="ghost-btn" disabled={currentPage === totalPages} onClick={() => setPage((prev) => Math.min(totalPages, prev + 1))}>Siguiente</button>
        </div>
      ) : null}
    </div>
  );
}

function DataPath() {
  const [path, setPath] = useState('cargando...');
  useEffect(() => { api.dataPath().then(setPath).catch((err) => setPath(err.message)); }, []);
  return <code className="path-box">{path}</code>;
}

function formatDate(value: string) {
  if (!value) return 'n/d';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}
