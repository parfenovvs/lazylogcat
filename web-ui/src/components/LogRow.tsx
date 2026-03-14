import { memo } from "react";

import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

import type { LogLine } from "../lib/types";

const LEVEL_COLOR: Record<string, string> = {
  V: "text.disabled",
  D: "info.main",
  I: "text.secondary",
  W: "warning.main",
  E: "error.main",
  F: "error.dark",
  S: "text.disabled",
};

const LEVEL_BG: Record<string, string> = {
  E: "error.main",
  F: "error.main",
  W: "warning.main",
};

function getLevelColor(level: string | undefined): string {
  if (!level) return "text.primary";
  return LEVEL_COLOR[level.toUpperCase()] ?? "text.primary";
}

function getLevelBg(level: string | undefined): string | undefined {
  if (!level) return undefined;
  const bg = LEVEL_BG[level.toUpperCase()];
  return bg;
}

interface LogRowProps {
  line: LogLine;
  highlighted?: boolean;
}

export default memo(function LogRow({ line, highlighted }: LogRowProps) {
  const levelColor = getLevelColor(line.level);
  const levelBg = getLevelBg(line.level);
  const upperLevel = line.level?.toUpperCase() ?? "?";

  return (
    <Box
      display="flex"
      alignItems="baseline"
      px={0.5}
      pr={1.5}
      py={0.25}
      minHeight="1.6rem"
      color={levelColor}
      sx={{
        borderLeft: 2,
        borderLeftColor: levelBg ?? "transparent",
        bgcolor: highlighted
          ? "action.selected"
          : levelBg
            ? "action.hover"
            : undefined,
      }}
    >
      <Typography
        variant="body2"
        component="span"
        color="text.secondary"
        noWrap
        sx={{ width: "7.5rem", flexShrink: 0, pr: 1, userSelect: "none" }}
      >
        {line.time}
      </Typography>

      <Typography
        variant="caption"
        component="span"
        color="text.disabled"
        noWrap
        sx={{ width: "4.5rem", flexShrink: 0, pr: 0.5, textAlign: "right", userSelect: "none" }}
      >
        {line.pid}
      </Typography>

      <Typography
        variant="body2"
        component="span"
        fontWeight="bold"
        noWrap
        sx={{ width: "1.6rem", flexShrink: 0, pr: 0.5, textAlign: "center", userSelect: "none" }}
      >
        {upperLevel}
      </Typography>

      <Typography
        variant="body2"
        component="span"
        color="text.secondary"
        noWrap
        sx={{ width: "12rem", flexShrink: 0, pr: 1, overflow: "hidden", textOverflow: "ellipsis" }}
      >
        {line.tag}
      </Typography>

      <Typography
        variant="body2"
        component="span"
        sx={{ flex: 1, wordBreak: "break-all" }}
      >
        {line.message || line.raw}
      </Typography>
    </Box>
  );
});
