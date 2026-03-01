# System Status Icons

This directory contains system monitoring scripts for the Stream Deck interface.

## Available Icons

- **cpu.lua** - Shows CPU usage percentage with color coding:
  - Green: < 60%
  - Orange: 60-80%
  - Red: > 80%

- **memory.lua** - Shows memory usage percentage with color coding:
  - Green: < 75%
  - Orange: 75-90%
  - Red: > 90%

- **disk.lua** - Shows disk usage for C: drive with color coding:
  - Green: < 85%
  - Orange: 85-95%
  - Red: > 95%

- **network.lua** - Shows network connectivity status:
  - Green: Online (can reach 8.8.8.8)
  - Red: Offline

- **uptime.lua** - Shows system uptime:
  - Days and hours if > 1 day
  - Hours and minutes if < 1 day

- **temperature.lua** - Shows CPU temperature (if available):
  - Green: < 65°C
  - Orange: 65-80°C
  - Red: > 80°C

- **shutdown.lua** - Safe system shutdown with confirmation

## Usage

These scripts use passive updates to continuously monitor system status. Place them in the system folder and they will appear as buttons in the Stream Deck interface.

All scripts are designed to work on Windows systems and use standard command-line tools for monitoring.