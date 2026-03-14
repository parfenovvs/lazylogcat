import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import ToggleButton from "@mui/material/ToggleButton";
import IconButton from "@mui/material/IconButton";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Box from "@mui/material/Box";
import CloseIcon from "@mui/icons-material/Close";
import RefreshIcon from "@mui/icons-material/Refresh";

import type { Device } from "../lib/types";
import ThemeSwitcher from "./ThemeSwitcher";

interface DeviceBarProps {
  devices: Device[];
  activeDevice: string | null;
  connected: boolean;
  wsConnected: boolean;
  onConnect: (deviceId: string) => void;
  onDisconnect: () => void;
  onRefresh: () => void;
}

export default function DeviceBar({
  devices,
  activeDevice,
  connected,
  wsConnected,
  onConnect,
  onDisconnect,
  onRefresh,
}: DeviceBarProps) {
  const handleChange = (_: React.MouseEvent<HTMLElement>, deviceId: string | null) => {
    if (deviceId === null) return;
    if (deviceId === activeDevice && connected) return;
    onConnect(deviceId);
  };

  return (
    <AppBar position="static" color="default" elevation={1}>
      <Toolbar variant="dense">
        <Typography variant="subtitle2" fontFamily="monospace" noWrap sx={{ mr: 2 }}>
          lazylogcat
        </Typography>

        <Chip
          size="small"
          label={wsConnected ? "live" : "offline"}
          color={wsConnected ? "success" : "error"}
          variant="outlined"
        />

        <Divider orientation="vertical" flexItem sx={{ mx: 1 }} />

        {devices.length > 0 ? (
          <Box display="flex" alignItems="center" gap={0.5} flex={1} overflow="hidden">
            <ToggleButtonGroup
              value={activeDevice ?? ""}
              exclusive
              onChange={handleChange}
              size="small"
            >
              {devices.map((device) => (
                <ToggleButton key={device.id} value={device.id}>
                  {device.name || device.id}
                </ToggleButton>
              ))}
            </ToggleButtonGroup>

            {activeDevice && connected && (
              <IconButton size="small" onClick={onDisconnect} title="Disconnect">
                <CloseIcon fontSize="small" />
              </IconButton>
            )}
          </Box>
        ) : (
          <Box flex={1}>
            <Typography variant="body2" color="text.disabled">
              no devices
            </Typography>
          </Box>
        )}

        <Divider orientation="vertical" flexItem sx={{ mx: 1 }} />

        <ThemeSwitcher />

        <IconButton onClick={onRefresh} title="Refresh device list" size="small">
          <RefreshIcon fontSize="small" />
        </IconButton>
      </Toolbar>
    </AppBar>
  );
}
