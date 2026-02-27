# Nomad Core - Wearable Computer Platform

A modular forearm-mounted wearable computer system built around the Raspberry Pi Zero 2W.

## 📐 Design Philosophy

**Nomad Core** is designed with three key principles:
1. **Modular** - Three separate modules (Compute, Display, Input) that can be individually upgraded
2. **Ergonomic** - Worn on the left forearm without requiring wrist rotation to view
3. **Accessible** - Open-source designs, 3D printable, uses common off-the-shelf components

## 🧩 System Architecture

```
┌──────────────────────────────────────────────────┐
│           MOUNTING PLATE (Main Platform)         │
│  ┌────────────┐                                  │
│  │   PUCK     │  ← Attaches to arm band          │
│  │ (Release)  │                                  │
│  └────────────┘                                  │
│                                                   │
│  [Compute Module]  [Display Module]  [Input]     │
│  - Pi Zero 2W      - 4.3" screen    - Gamepad    │
│  - X306 UPS        - 30° tilt       - Optional   │
│  - 2× 18650        - HDMI                        │
└──────────────────────────────────────────────────┘
```

### Module Layout on Forearm

When worn on **left arm**, hand pointing forward (+X direction):

- **PUCK**: Forward/hand side - quick-release mount (needs 81mm forward + 38mm side clearance)
- **COMPUTE**: Back/elbow side (outer forearm) - Pi + UPS + batteries
- **DISPLAY**: Top - angled toward wearer (30-35° tilt, 5-10° cant)
- **INPUT**: Inner forearm - gamepad dock or buttons

## 📦 Hardware Components

### Core Compute
- Raspberry Pi Zero 2W (with header)
- X306 UPS HAT (18650-based power)
- 2× 18650 batteries (parallel configuration)
- USB expansion board (4-port hub with UART)

### Display
- 4.3" 16:9 LCD (1280×720 or similar)
- HDMI interface (micro-HDMI from Pi)
- Right-angle connector + 50mm flexible ribbon cable
- USB-powered from Pi's hub

### Input
- Primary: Small USB/Bluetooth gamepad (e.g., 8BitDo Micro)
- Secondary: External Bluetooth keyboard (optional)
- Hot-swappable dock on inner forearm

### Mounting
- Modified arm band puck (phone mount style)
- 33mm × 33mm puck interface
- Release mechanism requires 81mm forward + 38mm side clearance

## 📏 Key Dimensions

| Component | Dimensions (mm) | Notes |
|-----------|-----------------|-------|
| Mounting Plate | 150 × 80 × 3 | 6" × 3.15" platform |
| Puck | 33 × 33 × 3.5 | Square mount with corner insets |
| Puck Clearance | 81 forward / 38 sides | For release mechanism |
| Baseplate | 81 × 56 × 4.8 | Pi + UPS mount |
| Display | ~101 × 60 × 5 | 4.3" 16:9 panel |
| Body Gap | 6mm | Clearance from skin (WiFi) |

See `src/nomad_core/design_parameters.py` for complete specifications.

## 🔧 Project Status

### ✅ Completed
- [x] Analyzed existing STL models (baseplate, puck)
- [x] Extracted accurate dimensions from models
- [x] Created centralized design parameters
- [x] Parametric mounting plate framework

### 🚧 In Progress
- [ ] Parametric baseplate with hex nut recesses
- [ ] Mounting plate prototype for testing
- [ ] Clearance zone visualization

### 📋 Planned
- [ ] Display housing with angle adjustment
- [ ] Input module dock (gamepad retention)
- [ ] Cable management system
- [ ] Ventilation patterns
- [ ] Battery retention system
- [ ] Assembly guide
- [ ] Print settings optimization

## 🛠️ Development Setup

### Requirements
- Python 3.11+
- Build123d (parametric CAD)
- Trimesh (STL analysis)

### Quick Start

```bash
# Create and activate virtual environment
python3 -m venv venv
source venv/bin/activate  # or `venv\Scripts\activate` on Windows

# Install dependencies
pip install build123d trimesh numpy

# View design parameters
python src/nomad_core/design_parameters.py

# Analyze existing models
python tools/analyze_both_pucks.py

# Generate mounting plate (when Build123d is fully configured)
python src/nomad_core/mounting_plate.py
```

## 📂 Project Structure

```
nomad-core/
├── designs/
│   ├── original/              # Original STL files
│   │   ├── puck.stl          # Basic puck (33×33mm)
│   │   ├── puck-with-surface-clearence.stl
│   │   ├── Nomad Baseplate.stl  # Current Pi mount (81×56mm)
│   │   └── ChatGPT-Nomad Wearable Parts List.md
│   ├── parametric/            # Generated parametric models
│   └── reference/             # Specs, measurements, datasheets
├── src/
│   └── nomad_core/
│       ├── design_parameters.py   # All measurements/specs
│       ├── mounting_plate.py      # Main platform design
│       ├── baseplate.py           # Pi + UPS mount
│       └── utils.py
├── tools/                     # Analysis and utility scripts
│   ├── analyze_both_pucks.py # STL analysis
│   └── recreate_*.py         # Model recreation tools
├── prototypes/                # Test prints and iterations
├── output/                    # Generated STL/STEP files (gitignored)
├── examples/                  # Example 3D modeling scripts
├── docs/
│   └── build-guides/
└── ...config files
```

## 🎯 Design Goals

### Ergonomics
- **No wrist rotation required** - display angled for flat-arm viewing
- **Natural typing position** - when using external keyboard
- **Distributed weight** - <500g total, wide strap contact
- **Thermal management** - heat directed away from body

### Modularity
- **Swappable display** - HDMI interface, independent housing
- **Swappable input** - USB/Bluetooth, hot-plug capable
- **Updatable compute** - Pi module can be replaced without reprinting entire assembly
- **Standard connectors** - no custom PCBs required

### Practicality
- **3D printable** - PLA/PETG, standard FDM printer
- **Off-the-shelf parts** - Amazon/AliExpress availability
- **Field serviceable** - M2.5/M3 screws, replaceable components
- **Cable management** - Short runs (50-100mm), strain relief

## 🔌 Cable Routing

- **Display**: 50mm right-angle micro-HDMI ribbon (Pi → Display, top route)
- **Input**: 100mm flexible USB cable (Pi hub → Input dock, inner route)
- **Power**: Internal UPS wiring (within compute module)
- **All cables**: Flexible/silicone jacket to prevent port damage

## ⚡ Power Budget

| Component | Power Draw | Notes |
|-----------|------------|-------|
| Pi Zero 2W | 2-3W | Varies with load |
| Display | 2-4W | Backlight dependent |
| USB devices | 0.5-2W | Gamepad, peripherals |
| **Total** | **~7-9W** | Moderate use |

**Battery Life**: 2× 3000mAh 18650 (parallel) @ 3.7V  
≈ 22Wh usable → **2.5-3 hours heavy use** / **4-5 hours light use**

UPS supports pass-through charging (can charge while powered).

## 📐 3D Printing Notes

- **Material**: PLA (prototyping) or PETG (final, better strength)
- **Layer Height**: 0.2mm recommended
- **Infill**: 30% (structural parts)
- **Walls**: 3-4 perimeters (1.2-1.6mm at 0.4mm nozzle)
- **Supports**: Required for overhang angles, display housing
- **Print orientation**: Flattest major surface down for strength

## 🤝 Contributing

This is an open-source project! Contributions welcome:
- Design improvements
- Alternative module designs (different Pi boards, displays, etc.)
- Documentation
- Build guides
- Photos of your builds

## 📝 License

TO BE DETERMINED (likely MIT or Creative Commons)

## 🙏 Credits

- Original concept and design: Merith-TK
- Community feedback and iteration
- Build123d, CadQuery, and OpenSCAD communities

---

**Status**: Early development - expect frequent changes!  
**Last Updated**: February 20, 2026
