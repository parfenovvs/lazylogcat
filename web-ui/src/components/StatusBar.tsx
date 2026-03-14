import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import Switch from "@mui/material/Switch";

interface StatusBarProps {
  activeDevice: string | null;
  connected: boolean;
  autoScroll: boolean;
  onToggleAutoScroll: () => void;
  onClear: () => void;
}

export default function StatusBar({
  activeDevice,
  connected,
  autoScroll,
  onToggleAutoScroll,
  onClear,
}: StatusBarProps) {
  return (
    <Box
      display="flex"
      alignItems="center"
      justifyContent="space-between"
      px={1.5}
      py={0.5}
      borderTop={1}
      borderColor="divider"
    >
      <Box display="flex" alignItems="center" gap={1}>
        {activeDevice && connected && (
          <Typography variant="body2" color="text.disabled">
            {activeDevice}
          </Typography>
        )}
      </Box>

      <Box display="flex" alignItems="center" gap={1}>
        <Button size="small" onClick={onClear}>
          clear
        </Button>

        <Box display="flex" alignItems="center">
          <Switch
            size="small"
            checked={autoScroll}
            onChange={onToggleAutoScroll}
          />
          <Typography variant="body2" color="text.secondary">
            scroll
          </Typography>
        </Box>
      </Box>
    </Box>
  );
}
