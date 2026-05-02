package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// ANSI-16 base colors (0-15).
// These map to the terminal emulator's configured color scheme, ensuring
// compatibility across themes without hardcoding hex values.
const (
	// Standard colors (0-7)
	black   string = "0"
	red     string = "1"
	green   string = "2"
	yellow  string = "3"
	blue    string = "4"
	magenta string = "5"
	cyan    string = "6"
	white   string = "7"

	// Bright/Bold colors (8-15)
	brightBlack   string = "8"
	brightRed     string = "9"
	brightGreen   string = "10"
	brightYellow  string = "11"
	brightBlue    string = "12"
	brightMagenta string = "13"
	brightCyan    string = "14"
	brightWhite   string = "15"
)

// Semantic color palette.
//
// All colors use ANSI-16 indices so they inherit the terminal emulator's
// configured theme. The terminal's color scheme handles light/dark adaptation.
var (
	// ColorPrimary is the main accent color for active borders, dialog titles,
	// and other prominent interactive elements.
	ColorPrimary = lipgloss.Color(brightCyan)

	// ColorRegular is the default foreground color for normal text and log messages.
	ColorRegular = lipgloss.Color("")

	// ColorMuted is for de-emphasized text such as help hints, inactive
	// labels, and secondary information.
	ColorMuted = lipgloss.Color(brightBlack)

	// ColorWarning is for warning messages and cautionary notifications.
	ColorWarning = lipgloss.Color(yellow)

	// ColorDanger is for error messages and validation warnings.
	ColorDanger = lipgloss.Color(brightRed)

	// ColorBorder is for inactive panel borders, dividers, and outlines.
	ColorBorder = lipgloss.Color(brightBlack)

	// ColorVisualBG is the background for visual-mode selected lines.
	ColorVisualBG = lipgloss.Color(yellow)

	// ColorVisualFG is the foreground for visual-mode selected lines.
	ColorVisualFG = lipgloss.Color(black)

	// ColorSelectedFG is the foreground for selected table rows.
	ColorSelectedFG = lipgloss.Color(black)

	// ColorSelectedBG is the background for selected table rows.
	ColorSelectedBG = lipgloss.Color(brightCyan)

	// ColorBrightMagenta highlights shortcut keys in dialog help (ANSI bright magenta).
	ColorBrightMagenta = lipgloss.Color(brightMagenta)
)

// GetLogColor returns a foreground color for the given logcat severity level.
// Colors use plain ANSI indices (identical for light/dark) because the
// terminal theme already provides appropriate shades for each index.
// IsDefaultForeground reports whether c is the terminal default foreground
// (no ANSI color sequence), matching [ColorRegular].
func IsDefaultForeground(c color.Color) bool {
	ar, ag, ab, aa := c.RGBA()
	br, bg, bb, ba := ColorRegular.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func GetLogColor(level string) color.Color {
	switch level {
	case "V":
		return ColorRegular
	case "D":
		return lipgloss.Color(blue)
	case "I":
		return lipgloss.Color(green)
	case "W":
		return lipgloss.Color(yellow)
	case "E":
		return lipgloss.Color(red)
	case "F":
		return lipgloss.Color(magenta)
	default:
		return ColorRegular
	}
}

// Panel returns a base panel style with an inactive border.
func Panel() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)
}

// Dialog returns a style for dialog overlays.
func Dialog() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 1)
}

// DialogTitle returns a style for dialog title text.
func DialogTitle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Padding(0, 1)
}

// DialogHelp returns a style for dialog help/footer text.
func DialogHelp() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1)
}

// DialogError returns a style for dialog error text.
func DialogError() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(ColorDanger).Padding(0, 1)
}

// DialogSearch returns a style for dialog search/input wrappers.
func DialogSearch() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

// TableHeader returns a style for table header rows.
func TableHeader() lipgloss.Style {
	return lipgloss.NewStyle()
}

// TableCell returns a style for table cell content.
func TableCell() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1)
}

// TableSelected returns a style for the selected table row.
func TableSelected() lipgloss.Style {
	return lipgloss.NewStyle().Bold(false).Foreground(ColorSelectedFG).Background(ColorSelectedBG)
}

// FilterBadge returns a style for active filter labels in the header.
func FilterBadge() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(ColorSelectedFG).
		Background(ColorSelectedBG).
		Padding(0, 1)
}
