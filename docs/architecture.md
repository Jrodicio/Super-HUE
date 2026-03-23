# Arquitectura de Super HUE

## Capas

- `internal/domain`: entidades, enums de triggers/acciones y puertos.
- `internal/services`: casos de uso de la aplicación, orquestación del bridge, reglas, dispositivos y logging.
- `internal/infrastructure`: SQLite, cliente Philips Hue, monitores de procesos/red.
- `internal/automation`: motor central de automatización y scheduler.
- `apps/desktop`: app Wails, bindings y frontend desktop.

## Flujo principal

1. La UI Wails invoca métodos del binding `backend.App`.
2. `App` delega en `services.AppService`.
3. `AppService` usa repositorios SQLite y el cliente Hue para leer/escribir estado.
4. El motor `automation.Engine` escucha eventos de:
   - `ProcessMonitor`
   - `NetworkMonitor`
   - scheduler interno por horario
5. Cuando una regla coincide, `AppService.ExecuteActions` ejecuta acciones sobre el Hue Bridge.
6. Todas las operaciones importantes escriben logs en SQLite y se muestran en la UI.

## Persistencia

La app guarda en SQLite:

- settings
- cache de luces
- rooms
- scenes
- rules
- rule_actions
- rule_conditions
- devices
- logs

## Notas MVP

- La conexión Hue usa registro local v1 (`/api`) y requiere presionar el botón del bridge la primera vez.
- La UI mantiene la lógica de negocio en Go; el frontend solo renderiza y llama bindings.
- La detección de procesos se apoya en `tasklist` en Windows y `ps` como fallback para desarrollo.
- La detección de presencia de red usa `ping` con tolerancia de 3 fallos consecutivos.
