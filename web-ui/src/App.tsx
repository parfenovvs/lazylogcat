import { createSignal, onCleanup, onMount, For } from "solid-js";
import type { LogLine, Device, ServerMessage } from "./lib/types";
import { LogcatWebSocket, getWebSocketUrl } from "./lib/ws";

export default function App() {
  const [lines, setLines] = createSignal<LogLine[]>([]);
  const [devices, setDevices] = createSignal<Device[]>([]);
  const [connected, setConnected] = createSignal(false);
  const [wsConnected, setWsConnected] = createSignal(false);
  const [activeDevice, setActiveDevice] = createSignal<string | null>(null);
  const [error, setError] = createSignal<string | null>(null);

  let ws: LogcatWebSocket | null = null;
  let logContainer: HTMLPreElement | undefined;

  const handleMessage = (msg: ServerMessage) => {
    switch (msg.type) {
      case "lines":
        setLines((prev) => {
          const next = [...prev, ...msg.data];
          // Keep last 10000 lines in the UI
          return next.length > 10000 ? next.slice(-10000) : next;
        });
        // Auto-scroll to bottom
        requestAnimationFrame(() => {
          if (logContainer) {
            logContainer.scrollTop = logContainer.scrollHeight;
          }
        });
        break;
      case "connected":
        setConnected(true);
        setActiveDevice(msg.deviceId);
        setError(null);
        break;
      case "disconnected":
        setConnected(false);
        if (msg.error) setError(msg.error);
        break;
      case "devices":
        setDevices(msg.data);
        break;
      case "error":
        setError(msg.message);
        break;
    }
  };

  const handleStatus = (connected: boolean) => {
    setWsConnected(connected);
  };

  onMount(() => {
    ws = new LogcatWebSocket(getWebSocketUrl(), handleMessage, handleStatus);
    ws.connect();
    // Request device list on connect
    setTimeout(() => {
      ws?.send({ type: "listDevices" });
    }, 500);
  });

  onCleanup(() => {
    ws?.disconnect();
  });

  const connectToDevice = (deviceId: string) => {
    setLines([]);
    setError(null);
    ws?.send({ type: "connect", deviceId });
  };

  const disconnect = () => {
    ws?.send({ type: "disconnect" });
    setConnected(false);
    setActiveDevice(null);
  };

  const refreshDevices = () => {
    ws?.send({ type: "listDevices" });
  };

  return (
    <div class="flex flex-col h-screen bg-gray-950 text-gray-100">
      {/* Header */}
      <header class="flex items-center justify-between px-4 py-2 bg-gray-900 border-b border-gray-800">
        <h1 class="text-sm font-bold tracking-wide">lazylogcat</h1>
        <div class="flex items-center gap-3 text-xs">
          <span class={wsConnected() ? "text-green-400" : "text-red-400"}>
            {wsConnected() ? "WS connected" : "WS disconnected"}
          </span>
          {connected() && (
            <span class="text-blue-400">Device: {activeDevice()}</span>
          )}
        </div>
      </header>

      {/* Controls */}
      <div class="flex items-center gap-2 px-4 py-2 bg-gray-900/50 border-b border-gray-800">
        <button
          class="px-3 py-1 text-xs bg-gray-800 hover:bg-gray-700 rounded border border-gray-700"
          onClick={refreshDevices}
        >
          Refresh Devices
        </button>
        <For each={devices()}>
          {(device) => (
            <button
              class={`px-3 py-1 text-xs rounded border ${
                activeDevice() === device.id
                  ? "bg-blue-600 border-blue-500 text-white"
                  : "bg-gray-800 hover:bg-gray-700 border-gray-700"
              }`}
              onClick={() => connectToDevice(device.id)}
            >
              {device.name} ({device.id})
            </button>
          )}
        </For>
        {connected() && (
          <button
            class="px-3 py-1 text-xs bg-red-900 hover:bg-red-800 rounded border border-red-700"
            onClick={disconnect}
          >
            Disconnect
          </button>
        )}
      </div>

      {/* Error banner */}
      {error() && (
        <div class="px-4 py-1 text-xs bg-red-900/50 text-red-300 border-b border-red-800">
          {error()}
        </div>
      )}

      {/* Log output */}
      <pre
        ref={logContainer}
        class="flex-1 overflow-auto px-4 py-2 text-xs font-mono leading-tight"
      >
        <For each={lines()}>
          {(line) => <div class="hover:bg-gray-900/50">{line.raw}</div>}
        </For>
      </pre>

      {/* Status bar */}
      <footer class="px-4 py-1 text-xs bg-gray-900 border-t border-gray-800 text-gray-500">
        {lines().length} lines
      </footer>
    </div>
  );
}
