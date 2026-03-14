import { useState, useEffect, useRef, useCallback } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Alert from "@mui/material/Alert";
import FiberManualRecord from "@mui/icons-material/FiberManualRecord";
import RadioButtonChecked from "@mui/icons-material/RadioButtonChecked";

import type { LogLine, Device, ServerMessage, Filter } from "./lib/types";
import { LogcatWebSocket, getWebSocketUrl } from "./lib/ws";
import DeviceBar from "./components/DeviceBar";
import FilterBar from "./components/FilterBar";
import LogRow from "./components/LogRow";
import StatusBar from "./components/StatusBar";

const MAX_LINES = 10_000;
const ROW_HEIGHT_ESTIMATE = 28;
const COL_HEADER_HEIGHT = 32;

export default function App() {
  const [lines, setLines] = useState<LogLine[]>([]);
  const [devices, setDevices] = useState<Device[]>([]);
  const [connected, setConnected] = useState(false);
  const [wsConnected, setWsConnected] = useState(false);
  const [activeDevice, setActiveDevice] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [autoScroll, setAutoScroll] = useState(true);

  const wsRef = useRef<LogcatWebSocket | null>(null);
  const logContainerRef = useRef<HTMLDivElement>(null);
  const autoScrollRef = useRef(true);
  const filterRef = useRef<Filter>({});

  useEffect(() => {
    autoScrollRef.current = autoScroll;
  }, [autoScroll]);

  const virtualizer = useVirtualizer({
    count: lines.length,
    getScrollElement: () => logContainerRef.current,
    estimateSize: () => ROW_HEIGHT_ESTIMATE,
    overscan: 30,
    paddingStart: COL_HEADER_HEIGHT,
  });

  useEffect(() => {
    if (autoScrollRef.current && lines.length > 0) {
      virtualizer.scrollToIndex(lines.length - 1, { align: "end" });
    }
  }, [lines.length, virtualizer]);

  const handleMessage = useCallback((msg: ServerMessage) => {
    switch (msg.type) {
      case "lines":
        setLines((prev) => {
          const next = [...prev, ...msg.data];
          return next.length > MAX_LINES ? next.slice(-MAX_LINES) : next;
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
      case "clearLines":
        setLines([]);
        break;
      case "error":
        setError(msg.message);
        break;
    }
  }, []);

  const handleStatus = useCallback((c: boolean) => {
    setWsConnected(c);
  }, []);

  useEffect(() => {
    const ws = new LogcatWebSocket(getWebSocketUrl(), handleMessage, handleStatus);
    wsRef.current = ws;
    ws.connect();
    const timer = setTimeout(() => {
      ws.send({ type: "listDevices" });
    }, 500);
    return () => {
      clearTimeout(timer);
      ws.disconnect();
    };
  }, [handleMessage, handleStatus]);

  const connectToDevice = useCallback((deviceId: string) => {
    setLines([]);
    setError(null);
    wsRef.current?.send({ type: "connect", deviceId, filter: filterRef.current });
  }, []);

  const disconnect = useCallback(() => {
    wsRef.current?.send({ type: "disconnect" });
    setConnected(false);
    setActiveDevice(null);
  }, []);

  const refreshDevices = useCallback(() => {
    wsRef.current?.send({ type: "listDevices" });
  }, []);

  const handleFilterChange = useCallback((filter: Filter) => {
    filterRef.current = filter;
    wsRef.current?.send({ type: "updateFilter", filter });
  }, []);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 60;
    setAutoScroll(atBottom);
  }, []);

  const clearLogs = useCallback(() => {
    setLines([]);
  }, []);

  const COL_HEADERS = [
    { label: "TIME", width: "7.5rem" },
    { label: "PID", width: "4.5rem", align: "right" as const },
    { label: "LVL", width: "1.6rem" },
    { label: "TAG", width: "12rem" },
    { label: "MESSAGE", flex: 1 },
  ];

  return (
    <Box display="flex" flexDirection="column" height="100vh" overflow="hidden">
      <DeviceBar
        devices={devices}
        activeDevice={activeDevice}
        connected={connected}
        wsConnected={wsConnected}
        onConnect={connectToDevice}
        onDisconnect={disconnect}
        onRefresh={refreshDevices}
      />

      <FilterBar onFilterChange={handleFilterChange} disabled={!connected} />

      {error && (
        <Alert severity="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {lines.length === 0 && (
        <Box
          flex={1}
          display="flex"
          flexDirection="column"
          alignItems="center"
          justifyContent="center"
          gap={1}
        >
          {connected ? (
            <>
              <RadioButtonChecked color="disabled" />
              <Typography variant="body2" color="text.secondary">
                waiting for logs
              </Typography>
              <Typography variant="caption" color="text.disabled">
                connected — logs will appear shortly
              </Typography>
            </>
          ) : (
            <>
              <FiberManualRecord color="disabled" />
              <Typography variant="body2" color="text.secondary">
                no device connected
              </Typography>
              <Typography variant="caption" color="text.disabled">
                {devices.length > 0
                  ? "select a device above to start streaming"
                  : "connect an Android device via ADB"}
              </Typography>
            </>
          )}
        </Box>
      )}

      {lines.length > 0 && (
        <Box
          ref={logContainerRef}
          onScroll={handleScroll}
          flex={1}
          overflow="auto"
          fontFamily="monospace"
          pb={0.5}
        >
          {/* Sticky column header */}
          <Box
            display="flex"
            alignItems="center"
            pl={1}
            pr={1}
            pb={0.5}
            position="sticky"
            top={0}
            zIndex={1}
            bgcolor="background.default"
            borderBottom={1}
            borderColor="divider"
            height={COL_HEADER_HEIGHT}
          >
            {COL_HEADERS.map((col) => (
              <Typography
                key={col.label}
                variant="caption"
                color="text.disabled"
                sx={{
                  width: col.flex ? undefined : col.width,
                  flex: col.flex,
                  textAlign: col.align ?? "left",
                  pr: 1,
                  letterSpacing: "0.08em",
                  textTransform: "uppercase",
                }}
              >
                {col.label}
              </Typography>
            ))}
          </Box>

          {/* Virtual list container */}
          <div
            style={{
              height: virtualizer.getTotalSize(),
              width: "100%",
              position: "relative",
            }}
          >
            {virtualizer.getVirtualItems().map((virtualRow) => (
              <div
                key={virtualRow.index}
                style={{
                  position: "absolute",
                  top: 0,
                  left: 0,
                  width: "100%",
                  transform: `translateY(${virtualRow.start}px)`,
                }}
              >
                <LogRow line={lines[virtualRow.index]} />
              </div>
            ))}
          </div>
        </Box>
      )}

      <StatusBar
        activeDevice={activeDevice}
        connected={connected}
        autoScroll={autoScroll}
        onToggleAutoScroll={() => setAutoScroll((v) => !v)}
        onClear={clearLogs}
      />
    </Box>
  );
}
