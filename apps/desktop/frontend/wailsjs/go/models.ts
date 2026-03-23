export namespace domain {
	
	export class Action {
	    id: number;
	    ruleId: number;
	    type: string;
	    target: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new Action(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ruleId = source["ruleId"];
	        this.type = source["type"];
	        this.target = source["target"];
	        this.value = source["value"];
	    }
	}
	export class Device {
	    id: number;
	    name: string;
	    ip: string;
	    present: boolean;
	    failureCount: number;
	    // Go type: time
	    lastSeenAt: any;
	    // Go type: time
	    lastCheckedAt: any;
	    consecutiveOks: number;
	
	    static createFrom(source: any = {}) {
	        return new Device(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.present = source["present"];
	        this.failureCount = source["failureCount"];
	        this.lastSeenAt = this.convertValues(source["lastSeenAt"], null);
	        this.lastCheckedAt = this.convertValues(source["lastCheckedAt"], null);
	        this.consecutiveOks = source["consecutiveOks"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Condition {
	    id: number;
	    ruleId: number;
	    type: string;
	    key: string;
	    value: string;
	    negate: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Condition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.ruleId = source["ruleId"];
	        this.type = source["type"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.negate = source["negate"];
	    }
	}
	export class Rule {
	    id: number;
	    name: string;
	    trigger: string;
	    enabled: boolean;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    conditions: Condition[];
	    actions: Action[];
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.trigger = source["trigger"];
	        this.enabled = source["enabled"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.conditions = this.convertValues(source["conditions"], Condition);
	        this.actions = this.convertValues(source["actions"], Action);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Scene {
	    id: string;
	    name: string;
	    groupId: string;
	    group: string;
	
	    static createFrom(source: any = {}) {
	        return new Scene(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.groupId = source["groupId"];
	        this.group = source["group"];
	    }
	}
	export class Room {
	    id: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Room(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class Light {
	    id: string;
	    name: string;
	    roomId: string;
	    roomName: string;
	    on: boolean;
	    brightness: number;
	    colorHex: string;
	    reachable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Light(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.roomId = source["roomId"];
	        this.roomName = source["roomName"];
	        this.on = source["on"];
	        this.brightness = source["brightness"];
	        this.colorHex = source["colorHex"];
	        this.reachable = source["reachable"];
	    }
	}
	export class LogEntry {
	    id: number;
	    level: string;
	    source: string;
	    message: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.source = source["source"];
	        this.message = source["message"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectionStatus {
	    bridgeIp: string;
	    connected: boolean;
	    bridgeReady: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bridgeIp = source["bridgeIp"];
	        this.connected = source["connected"];
	        this.bridgeReady = source["bridgeReady"];
	        this.message = source["message"];
	    }
	}
	export class Dashboard {
	    connectionStatus: ConnectionStatus;
	    lightsCount: number;
	    activeRules: number;
	    devicesPresent: number;
	    recentLogs: LogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Dashboard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionStatus = this.convertValues(source["connectionStatus"], ConnectionStatus);
	        this.lightsCount = source["lightsCount"];
	        this.activeRules = source["activeRules"];
	        this.devicesPresent = source["devicesPresent"];
	        this.recentLogs = this.convertValues(source["recentLogs"], LogEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppState {
	    dashboard: Dashboard;
	    lights: Light[];
	    rooms: Room[];
	    scenes: Scene[];
	    rules: Rule[];
	    devices: Device[];
	    logs: LogEntry[];
	    settings: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dashboard = this.convertValues(source["dashboard"], Dashboard);
	        this.lights = this.convertValues(source["lights"], Light);
	        this.rooms = this.convertValues(source["rooms"], Room);
	        this.scenes = this.convertValues(source["scenes"], Scene);
	        this.rules = this.convertValues(source["rules"], Rule);
	        this.devices = this.convertValues(source["devices"], Device);
	        this.logs = this.convertValues(source["logs"], LogEntry);
	        this.settings = source["settings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	
	
	

}

