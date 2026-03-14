import { useState, useCallback, type KeyboardEvent } from "react";

import Box from "@mui/material/Box";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import ToggleButton from "@mui/material/ToggleButton";
import Divider from "@mui/material/Divider";
import CloseIcon from "@mui/icons-material/Close";

import type { Filter } from "../lib/types";

const LEVELS = ["V", "D", "I", "W", "E", "F"] as const;

interface FilterBarProps {
  onFilterChange: (filter: Filter) => void;
  disabled?: boolean;
}

export default function FilterBar({ onFilterChange, disabled }: FilterBarProps) {
  const [tag, setTag] = useState("");
  const [text, setText] = useState("");
  const [pkg, setPkg] = useState("");
  const [level, setLevel] = useState("V");

  const buildFilter = useCallback(
    (overrides?: { level?: string }) => {
      const filter: Filter = {};
      if (tag) filter.tag = tag;
      if (text) filter.text = text;
      if (pkg) filter.packageName = pkg;
      const lv = overrides?.level ?? level;
      if (lv) filter.level = lv;
      return filter;
    },
    [tag, text, pkg, level],
  );

  const emit = () => {
    onFilterChange(buildFilter());
  };

  const handleKey = (e: KeyboardEvent) => {
    if (e.key === "Enter") emit();
  };

  const selectLevel = (_: React.MouseEvent<HTMLElement>, newLevel: string | null) => {
    if (newLevel === null) return;
    setLevel(newLevel);
    onFilterChange(buildFilter({ level: newLevel }));
  };

  const clearAll = () => {
    setTag("");
    setText("");
    setPkg("");
    setLevel("V");
    onFilterChange({});
  };

  const hasFilter = tag || text || pkg || level !== "V";

  return (
    <Box display="flex" alignItems="center" gap={1} px={1} py={0.5} borderBottom={1} borderColor="divider">
      <ToggleButtonGroup
        value={level}
        exclusive
        onChange={selectLevel}
        disabled={disabled}
        size="small"
      >
        {LEVELS.map((l) => (
          <ToggleButton key={l} value={l}>
            {l}
          </ToggleButton>
        ))}
      </ToggleButtonGroup>

      <Divider orientation="vertical" flexItem />

      <TextField
        label="tag"
        size="small"
        variant="outlined"
        value={tag}
        onChange={(e) => setTag(e.target.value)}
        onKeyDown={handleKey}
        disabled={disabled}
        slotProps={{ htmlInput: { sx: { fontFamily: "monospace" } } }}
      />

      <TextField
        label="package"
        size="small"
        variant="outlined"
        value={pkg}
        onChange={(e) => setPkg(e.target.value)}
        onKeyDown={handleKey}
        disabled={disabled}
        slotProps={{ htmlInput: { sx: { fontFamily: "monospace" } } }}
      />

      <TextField
        label="message"
        size="small"
        variant="outlined"
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKey}
        disabled={disabled}
        sx={{ flex: 1 }}
        slotProps={{ htmlInput: { sx: { fontFamily: "monospace" } } }}
      />

      <Button
        variant="contained"
        size="small"
        onClick={emit}
        disabled={disabled}
      >
        Apply
      </Button>

      {hasFilter && (
        <IconButton size="small" onClick={clearAll} title="Clear filters">
          <CloseIcon fontSize="small" />
        </IconButton>
      )}
    </Box>
  );
}
