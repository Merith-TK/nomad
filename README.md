# NOMAD Core
**Networked Optimized Mobile Access Device**

> **Note:** This project is developed with assistance from Claude Sonnet 4.5 via GitHub Copilot.

A modular, 3D-printable forearm-mounted wearable computer platform built around the Raspberry Pi Zero 2W.

## What is NOMAD?

NOMAD Core is an open-source wearable computer designed to be worn on your left forearm. It's built from three swappable modules:

- **Compute Module** - Raspberry Pi Zero 2W with UPS and battery power (1× 18650 cells)
- **Display Module** - 4.3" screen angled for comfortable viewing
- **Input Module** - Swappable gamepad or keyboard interface

The entire system mounts to an arm band via a quick-release puck mechanism, making it easy to put on and take off.

## Design Philosophy

> **Core Principle: Modularity**  
> The only fixed standards in NOMAD are the **armband mount**, **M2.5 screws**, and **33×33mm puck interface**. Everything else is designed to be customized and modified to your needs. Want to use a different SBC? Model your own brain mount. Need a bigger battery? Design a custom power module. The baseplate and puck system provide the foundation for whatever you want to build.

- **Modular** - Swap out individual modules without rebuilding the whole system
- **Ergonomic** - Worn comfortably on the forearm, no wrist rotation needed
- **Open Source** - All CAD files, documentation, and designs are freely available
- **DIY-Friendly** - Designed for 3D printing with common hardware components

## Repository Contents

### 📐 CAD Files (`cad/`)
OpenSCAD parametric designs for all 3D-printed components:
- `00_assembly_complete.scad` - Full system assembly view
- `puck_design.scad` - Quick-release mounting interface
- `baseplate_initial.scad` - Raspberry Pi mounting plate
- `brain_mount.scad` - Compute module bracket
- `parts/` - Reusable modules (screw cutouts, support columns)
- `archive/` - Previous design iterations

### 📋 Documentation (`plans/`)
- `hardware/DESIGN.md` - Complete design specifications and system architecture
- `README.md` - Overview of planning documentation

### ⚙️ Other Directories
- `apps/` - Future: User applications for the device
- `firmware/` - Future: System configurations and OS images

## Getting Started

### Prerequisites
- OpenSCAD (or use the included devcontainer)
- 3D printer
- Basic hardware: M2.5 screws, arm band mount

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/Merith-TK/nomad.git
   cd nomad
   ```

2. **Open CAD files**
   ```bash
   openscad cad/00_assembly_complete.scad
   ```

3. **Generate STL files for printing**
   ```bash
   openscad -o puck.stl cad/puck_design.scad
   openscad -o baseplate.stl cad/baseplate_initial.scad
   openscad -o brain_mount.stl cad/brain_mount.scad
   ```

4. **Print and assemble**
   - Slice the STL files in your preferred slicer
   - Print with 20-50% infill
   - Assemble with M2.5 × 16mm screws

## Development Environment

This repository includes a devcontainer with OpenSCAD and Python CAD tools pre-installed:

1. Open in VS Code
2. Install "Dev Containers" extension
3. Press `F1` → "Dev Containers: Reopen in Container"
4. Wait for container build (first time only)

## Hardware Specifications

| Component | Specification |
|-----------|--------------|
| Platform | Raspberry Pi Zero 2W |
| Power | UPS HAT + 1× 18650 battery |
| Display | 4.3" LCD (planned) |
| Mounting | 33×33mm puck interface |
| Print Volume | ~150×80×40mm total |

Full specifications in [`plans/hardware/DESIGN.md`](plans/hardware/DESIGN.md)

## Bill of Materials

### Required Components

| Component | Description | Link |
|-----------|-------------|------|
| **Raspberry Pi Zero 2WH** | Kit includes GPIO header, heatsink, mini-HDMI adapter, and OTG cable | [Amazon](https://www.amazon.com/dp/B0DRRDJKDV/) |
| **Geekworm X306 UPS** | UPS HAT for single 18650 battery | [Amazon](https://www.amazon.com/dp/B0B74NT38D/) |
| **M2.5 Screw Kit** | Assorted M2.5 screws, nuts, lock nuts, and washers | [Amazon](https://www.amazon.com/dp/B0FJ1YMG3J/) |
| **Arm Band Mount** | Phone-style arm/wristband for wearable mounting | [Amazon](https://www.amazon.com/dp/B07Z76ZB8G/) |

### USB Hub (Choose One)

Pick based on your connectivity needs:

| Option | Ports | Features | Link |
|--------|-------|----------|------|
| **Option A** | 3× USB + Ethernet | Includes GPIO header extender (no GPIO loss) | [Amazon](https://www.amazon.com/dp/B07X1BH5FN/) |
| **Option B** | 4× USB + UART | USB-to-UART via micro-USB for PC connection | [Amazon](https://www.amazon.com/dp/B06Y2TSR1D/) |

> **Note:** Both USB hubs have a port positioned to fit a USB dongle inside the housing without modification (housing design pending).

### Additional Items (Not Linked)
- 1× 18650 battery cell
- 3D printer filament (PLA/PETG recommended)

## Current Status

✅ **Completed:**
- Puck mounting system (fit verified)
- Baseplate for Raspberry Pi 3B+ (test platform)
- Brain mount bracket with screw holes
- Modular assembly system with adjustable positioning

🚧 **In Progress:**
- Display housing design
- Cable management system

📋 **Planned:**
- Input module dock
- Battery retention system
- Assembly instructions and build guide

## Contributing

This is an open-source hardware project. Contributions, suggestions, and improvements are welcome!

## License

Open source hardware - details TBD

---

**Project Status:** Early development - functional components, active iteration
