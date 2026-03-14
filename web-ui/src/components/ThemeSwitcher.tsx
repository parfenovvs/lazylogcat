import { useState, type MouseEvent } from "react";

import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import { useColorScheme } from "@mui/material/styles";
import BrightnessAutoIcon from "@mui/icons-material/BrightnessAuto";
import LightModeIcon from "@mui/icons-material/LightMode";
import DarkModeIcon from "@mui/icons-material/DarkMode";

type Mode = "system" | "light" | "dark";

const THEMES: { value: Mode; icon: React.ReactNode; label: string }[] = [
  { value: "system", icon: <BrightnessAutoIcon fontSize="small" />, label: "System" },
  { value: "light", icon: <LightModeIcon fontSize="small" />, label: "Light" },
  { value: "dark", icon: <DarkModeIcon fontSize="small" />, label: "Dark" },
];

export default function ThemeSwitcher() {
  const { mode, setMode } = useColorScheme();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);

  const handleOpen = (event: MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const select = (next: Mode) => {
    setMode(next);
    handleClose();
  };

  const currentIcon = THEMES.find((t) => t.value === (mode ?? "system"))?.icon ?? <BrightnessAutoIcon fontSize="small" />;

  return (
    <>
      <IconButton
        size="small"
        onClick={handleOpen}
        title={`${mode ?? "system"} theme`}
      >
        {currentIcon}
      </IconButton>

      <Menu
        anchorEl={anchorEl}
        open={open}
        onClose={handleClose}
      >
        {THEMES.map((t) => (
          <MenuItem
            key={t.value}
            selected={t.value === (mode ?? "system")}
            onClick={() => select(t.value)}
          >
            <ListItemIcon>{t.icon}</ListItemIcon>
            <ListItemText>{t.label}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
