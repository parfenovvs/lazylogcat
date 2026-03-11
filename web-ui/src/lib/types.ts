// TypeScript types mirroring Go JSON structs

export interface LogLine {
  date: string;
  time: string;
  pid: string;
  tid: string;
  level: string;
  tag: string;
  message: string;
  raw: string;
}

export interface Device {
  id: string;
  name: string;
}

export enum TextFilterMode {
  Contains = 0,
  Exact = 1,
  Regex = 2,
}

export interface TextFilter {
  value: string;
  mode: TextFilterMode;
}

export interface Filter {
  packageName?: TextFilter | string;
  level?: string;
  tag?: TextFilter | string;
  text?: TextFilter | string;
}

// Server -> Client messages

export interface LinesMessage {
  type: "lines";
  data: LogLine[];
}

export interface ConnectedMessage {
  type: "connected";
  deviceId: string;
}

export interface DisconnectedMessage {
  type: "disconnected";
  error?: string;
}

export interface DevicesMessage {
  type: "devices";
  data: Device[];
}

export interface ErrorMessage {
  type: "error";
  message: string;
}

export type ServerMessage =
  | LinesMessage
  | ConnectedMessage
  | DisconnectedMessage
  | DevicesMessage
  | ErrorMessage;

// Client -> Server messages

export interface ConnectCommand {
  type: "connect";
  deviceId: string;
  filter?: Filter;
}

export interface DisconnectCommand {
  type: "disconnect";
}

export interface UpdateFilterCommand {
  type: "updateFilter";
  filter: Filter;
}

export interface ListDevicesCommand {
  type: "listDevices";
}

export type ClientMessage =
  | ConnectCommand
  | DisconnectCommand
  | UpdateFilterCommand
  | ListDevicesCommand;
