package web

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/parfenovvs/lazylogcat/internal/model"
	"github.com/parfenovvs/lazylogcat/internal/util"
)

const (
	demoMinInterval = 20 * time.Millisecond
	demoMaxInterval = 200 * time.Millisecond
	demoBurstChance = 0.1 // 10% chance of a burst each tick
	demoBurstMin    = 3   // minimum lines in a burst
	demoBurstMax    = 12  // maximum lines in a burst
	demoMaxPending  = 10000
)

// demoDevice is the fake device returned in demo mode.
var demoDevice = model.Device{Id: "demo-device", Name: "DemoPhone"}

// demoApp represents a fake Android process for demo line generation.
type demoApp struct {
	pkg  string
	pid  string
	tids []string
	tags []string
}

var demoApps = []demoApp{
	{
		pkg:  "com.demo.app",
		pid:  "3950",
		tids: []string{"3950", "3981", "4005", "4012"},
		tags: []string{"MainActivity", "ViewModel", "Repository", "NetworkClient", "DataSync"},
	},
	{
		pkg:  "com.android.systemui",
		pid:  "1842",
		tids: []string{"1842", "1860", "1875"},
		tags: []string{"SystemUI", "StatusBar", "NotificationManager", "NavigationBar"},
	},
	{
		pkg:  "com.android.launcher3",
		pid:  "2105",
		tids: []string{"2105", "2130", "2145"},
		tags: []string{"Launcher", "Workspace", "AllApps", "IconCache"},
	},
	{
		pkg:  "system_server",
		pid:  "1000",
		tids: []string{"1000", "1015", "1032", "1048", "1064"},
		tags: []string{"ActivityManager", "WindowManager", "PackageManager", "PowerManager", "InputDispatcher"},
	},
	{
		pkg:  "com.google.android.gms",
		pid:  "4521",
		tids: []string{"4521", "4550", "4580"},
		tags: []string{"GmsClient", "LocationService", "AuthProxy", "GCM"},
	},
}

// demoMessages are message templates per tag, giving realistic output.
var demoMessages = map[string][]string{
	"MainActivity":        {"onCreate called", "onResume lifecycle", "onPause lifecycle", "Navigating to SettingsFragment", "Back pressed, finishing activity"},
	"ViewModel":           {"Loading user profile", "Data refresh triggered", "Saving state to bundle", "Clearing cached data", "Observer count: %d"},
	"Repository":          {"Cache hit for key=%s", "Fetching from network: /api/v2/users", "Insert %d rows into local DB", "Sync completed in %dms", "Pruning stale entries"},
	"NetworkClient":       {"POST /api/v2/events -> 200 (%dms)", "GET /api/v2/config -> 200 (%dms)", "Connection pool: %d active", "Retry attempt %d for /api/v2/sync", "TLS handshake completed"},
	"DataSync":            {"Starting background sync", "Sync batch: %d items", "Conflict resolved for record %s", "Upload queue depth: %d", "Sync finished, next in %ds"},
	"SystemUI":            {"onConfigurationChanged", "Dark mode toggled", "Battery level: %d%%", "Signal strength changed", "Keyguard state: SHOWN"},
	"StatusBar":           {"Updating clock display", "Notification count: %d", "Expanding panel", "Collapsing panel", "Icon slot updated: wifi"},
	"NotificationManager": {"Posting notification id=%d", "Canceling notification id=%d", "Channel importance: HIGH", "DND mode: OFF", "Enqueued: %s"},
	"NavigationBar":       {"Button pressed: BACK", "Button pressed: HOME", "Button pressed: RECENTS", "Rotation locked: portrait", "Gesture: swipe up"},
	"Launcher":            {"Binding workspace", "Loading widget: %s", "All apps container opened", "Search query: %s", "Workspace page: %d"},
	"Workspace":           {"Adding item at (x=%d, y=%d)", "Removing shortcut: %s", "Rearranging icons", "Page count: %d", "Drag started"},
	"AllApps":             {"Scrolling app list", "Filter applied: %s", "App count: %d", "Opening app info: %s", "Reset scroll position"},
	"IconCache":           {"Cache size: %d entries", "Evicting stale icons", "Loading icon for %s", "Cache hit ratio: %d%%", "Regenerating adaptive icon"},
	"ActivityManager":     {"Start proc %d:%s for activity", "Killing %d:%s (adj %d): empty", "Process %s died", "ANR in %s", "Force stopping %s appId=%d"},
	"WindowManager":       {"addWindow: %s", "removeWindow: %s", "Focus changed: %s", "Relayout window: %s", "Animation: TRANSIT_OPEN"},
	"PackageManager":      {"Scanning package: %s", "Package %s verified", "Install session %d created", "Permission granted: %s", "Reconciling packages"},
	"PowerManager":        {"Wakelock acquired: %s", "Wakelock released: %s", "Going to sleep", "Waking up from sleep", "Battery saver: OFF"},
	"InputDispatcher":     {"Delivering touch event to %s", "Key event: KEYCODE_BACK", "Focus request from %s", "Input channel registered: %s", "Dropped event: no focused window"},
	"GmsClient":           {"Connected to Google Play services", "API call: %s", "Service binding: %s", "Auth token refreshed", "Module sync completed"},
	"LocationService":     {"Location update: lat=%.4f, lon=%.4f", "Provider enabled: GPS", "Geofence triggered: %s", "Accuracy: %.1fm", "Requesting location updates"},
	"AuthProxy":           {"Token refresh for account: %s", "OAuth2 flow initiated", "Credentials cached", "Session validated", "Sign-out requested"},
	"GCM":                 {"Message received from: %s", "Registration token refreshed", "Upstream message sent", "Heartbeat: OK", "Connection state: CONNECTED"},
}

// demoLevels and their weights for random selection.
// Heavily weighted toward D/I with occasional W, rare E, very rare F.
var demoLevelWeights = []struct {
	level  string
	weight int
}{
	{"V", 5},
	{"D", 35},
	{"I", 35},
	{"W", 15},
	{"E", 8},
	{"F", 2},
}

var demoTotalWeight int

func init() {
	for _, lw := range demoLevelWeights {
		demoTotalWeight += lw.weight
	}
}

func pickLevel() string {
	n := rand.IntN(demoTotalWeight)
	for _, lw := range demoLevelWeights {
		n -= lw.weight
		if n < 0 {
			return lw.level
		}
	}
	return "D"
}

// demoReader implements LogReader by generating synthetic logcat lines.
type demoReader struct {
	mu        sync.Mutex
	pending   []model.LogLine
	connected bool
	lastErr   error

	filterMu sync.RWMutex
	filter   model.Filter
	pidSet   map[string]struct{}

	cancel context.CancelFunc
	done   chan struct{}
}

func newDemoReader() *demoReader {
	return &demoReader{
		pending: make([]model.LogLine, 0, 256),
	}
}

func (r *demoReader) Connect(_ string, filter model.Filter) error {
	r.Disconnect()

	r.filterMu.Lock()
	r.filter = filter
	r.pidSet = nil
	r.filterMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	r.mu.Lock()
	r.cancel = cancel
	r.connected = true
	r.lastErr = nil
	r.pending = make([]model.LogLine, 0, 256)
	r.done = make(chan struct{})
	r.mu.Unlock()

	go r.generateLoop(ctx)
	return nil
}

func (r *demoReader) Disconnect() {
	r.mu.Lock()
	cancel := r.cancel
	done := r.done
	wasConnected := r.connected
	r.mu.Unlock()

	if !wasConnected {
		return
	}

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}

	r.mu.Lock()
	r.connected = false
	r.cancel = nil
	r.done = nil
	r.pending = make([]model.LogLine, 0, 256)
	r.mu.Unlock()
}

func (r *demoReader) UpdateFilter(filter model.Filter) {
	r.filterMu.Lock()
	defer r.filterMu.Unlock()
	r.filter = filter
}

func (r *demoReader) UpdatePIDSet(pidSet map[string]struct{}) {
	r.filterMu.Lock()
	defer r.filterMu.Unlock()
	r.pidSet = pidSet
}

func (r *demoReader) Drain() []model.LogLine {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.pending) == 0 {
		return nil
	}
	lines := r.pending
	r.pending = make([]model.LogLine, 0, cap(lines))
	return lines
}

func (r *demoReader) IsConnected() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.connected
}

func (r *demoReader) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastErr
}

func (r *demoReader) WaitForDone() {
	r.mu.Lock()
	done := r.done
	r.mu.Unlock()
	if done != nil {
		<-done
	}
}

// generateLoop produces random log lines until the context is cancelled.
func (r *demoReader) generateLoop(ctx context.Context) {
	defer close(r.done)

	now := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Decide how many lines to emit: usually 1, sometimes a burst
		count := 1
		if rand.Float64() < demoBurstChance {
			count = demoBurstMin + rand.IntN(demoBurstMax-demoBurstMin+1)
		}

		for range count {
			select {
			case <-ctx.Done():
				return
			default:
			}

			now = now.Add(time.Duration(rand.IntN(50)+1) * time.Millisecond)
			raw := r.generateLine(now)
			line := model.ParseLogLine(raw)

			r.filterMu.RLock()
			pass := util.MatchesFilter(raw, line, &r.filter, r.pidSet)
			r.filterMu.RUnlock()

			if !pass {
				continue
			}

			r.mu.Lock()
			if len(r.pending) < demoMaxPending {
				r.pending = append(r.pending, line)
			}
			r.mu.Unlock()
		}

		// Sleep a random interval between lines
		delay := demoMinInterval + time.Duration(rand.Int64N(int64(demoMaxInterval-demoMinInterval)))
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
	}
}

// generateLine creates a single raw logcat threadtime-format line.
func (r *demoReader) generateLine(t time.Time) string {
	app := demoApps[rand.IntN(len(demoApps))]
	tag := app.tags[rand.IntN(len(app.tags))]
	tid := app.tids[rand.IntN(len(app.tids))]
	level := pickLevel()
	msg := pickMessage(tag)

	return fmt.Sprintf("%02d-%02d %02d:%02d:%02d.%03d %5s %5s %s %s: %s",
		t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6,
		app.pid, tid, level, tag, msg,
	)
}

// pickMessage selects a random message template for the given tag and
// fills in any format placeholders with random values.
func pickMessage(tag string) string {
	templates, ok := demoMessages[tag]
	if !ok || len(templates) == 0 {
		return "processing..."
	}
	tmpl := templates[rand.IntN(len(templates))]

	// Simple placeholder filling — count format verbs and fill generically
	return fillTemplate(tmpl)
}

// fillTemplate replaces common fmt verbs with random plausible values.
func fillTemplate(tmpl string) string {
	var args []any
	i := 0
	for i < len(tmpl) {
		if tmpl[i] == '%' && i+1 < len(tmpl) {
			switch tmpl[i+1] {
			case 'd':
				args = append(args, rand.IntN(9999)+1)
				i += 2
			case 's':
				args = append(args, pickRandomString())
				i += 2
			case '.':
				// Handle %.Nf patterns
				args = append(args, rand.Float64()*100)
				// Skip to end of verb
				for i < len(tmpl) && tmpl[i] != 'f' && tmpl[i] != 'g' {
					i++
				}
				if i < len(tmpl) {
					i++
				}
			case '%':
				// Literal %% — no arg needed
				i += 2
			default:
				i++
			}
		} else {
			i++
		}
	}
	if len(args) == 0 {
		return tmpl
	}
	return fmt.Sprintf(tmpl, args...)
}

var randomStrings = []string{
	"com.demo.app", "com.android.settings", "com.google.android.gms",
	"wifi_scan_01", "auth_session", "content_provider",
	"user_prefs", "WorkManager", "JobScheduler", "AlarmManager",
	"main_activity", "splash_screen", "image_loader",
}

func pickRandomString() string {
	return randomStrings[rand.IntN(len(randomStrings))]
}
