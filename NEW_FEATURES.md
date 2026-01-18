# What's New in LazyHog TUI

## New Features

### 🎨 Refreshed Design
- **PostHog Orange** brand color throughout the UI
- **Enhanced JSON syntax highlighting**: Blue keys, green values, yellow strings, purple numbers
- **Nerd Font icons** with automatic fallback for better visual hierarchy

### ⌨️ Enhanced Navigation
- **'b' key**: Collapse sidebar to icon-only mode for more screen space
- **'1'/'2'/'3' keys**: Jump directly to Pane 1, 2, or 3 (context-aware)
- **Improved help system**: Footer now shows context-relevant keybindings

### 🛡️ Safety Features
- **Environment warnings**: Confirmation dialogs for production flag toggles
- **Visual indicators**: DEV environments show green, PROD shows red warnings
- **Explicit confirmation**: "Yes, proceed" / "No, cancel" for critical actions

### 🚀 Performance Improvements
- **Smoother scrolling**: Using Bubble Tea's viewport component
- **Better table rendering**: Native table component for events/persons/flags
- **Optimized list navigation**: Built-in keyboard handling

---

## Keyboard Shortcuts

### Global
- `?` - Toggle help overlay
- `q` - Quit app
- `Ctrl+C` - Force quit
- `Tab` / `→` / `l` - Move focus right
- `Shift+Tab` / `←` / `h` / `Esc` - Move focus left / Go back
- **NEW**: `b` - Toggle sidebar collapse
- **NEW**: `1` / `2` / `3` - Jump to Pane 1 / 2 / 3

### Resource Selector (Pane 1)
- `↑`/`↓` or `j`/`k` - Navigate resources
- `Enter` - Cycle to next project
- `1` / `2` / `3` - Quick select Events / Persons / Flags

### List View (Pane 2)
- `↑`/`↓` or `j`/`k` - Navigate list
- `G` - Jump to bottom (resume auto-scroll)
- `/` - Search/filter
- `r` - Refresh current resource
- `p` - Pivot to person (Events only)

### Inspector (Pane 3)
- `j`/`k` or `↑`/`↓` - Scroll content
- `Space` - Fold/expand JSON object
- `Shift+Z` - Fold/expand all top-level keys
- `y` - Copy full JSON to clipboard
- `c` - Copy ID to clipboard
- `p` - Pivot to person (Events only)

---

## Configuration

Optional settings in `~/.config/ph-tui.yaml`:

```yaml
# Environment (auto-detected from instance URL if not set)
environment: "prod"  # or "dev"

# Visual theme
theme: "orange"  # New default PostHog orange theme

# Icon support (auto-detected based on terminal)
use_nerd_fonts: true  # Set to false to force Unicode fallback
```

---

## What Stayed the Same

All your favorite features are still here:
- ✅ Live event streaming with auto-scroll
- ✅ Person lookup and inspection
- ✅ Feature flag management
- ✅ Search and filtering
- ✅ Pivot from events to persons
- ✅ JSON viewing with folding
- ✅ Clipboard operations
- ✅ Project switching
- ✅ Smart polling (pauses when inspecting)

---

## Technical Improvements

Under the hood, LazyHog has been modernized:
- **Bubble Tea v1.1.0**: Latest version with improved performance
- **Bubbles components**: Using official list, table, viewport, and help components
- **Huh forms**: Beautiful confirmation dialogs
- **Better architecture**: More maintainable and extensible code

---

## Getting Started

1. **Login** (if first time):
   ```bash
   lazyhog login
   ```

2. **Run**:
   ```bash
   lazyhog
   ```

3. **Explore**:
   - Press `?` for full help
   - Try collapsing the sidebar with `b`
   - Jump between panes with `1`, `2`, `3`
   - Toggle a feature flag to see environment warnings

---

## Feedback

Found a bug or have a suggestion? Please open an issue on GitHub!

---

**Version**: Latest
**Updated**: 2026-01-18
