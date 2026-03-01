# NOMAD Lua Standard Library

This document describes the additional Lua modules available in the NOMAD scripting environment beyond the standard Lua libraries.

## Table of Contents

- [Available Modules](#available-modules)
  - [json - JSON Encoding/Decoding](#json---json-encodingdecoding)
  - [file - File System Operations](#file---file-system-operations)
  - [time - Time and Date Utilities](#time---time-and-date-utilities)
  - [log - Structured Logging](#log---structured-logging)
- [Configuration File](#configuration-file)
- [Migration Guide](#migration-guide)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Available Modules

### `json` - JSON Encoding/Decoding

Provides JSON serialization and parsing capabilities for data persistence and API communication.

#### Functions

- `json.encode(table)` - Converts a Lua table to a JSON string
- `json.decode(string)` - Parses a JSON string into a Lua table

#### Example

```lua
local json = require("json")

-- Encode Lua table to JSON string
local data = {name = "John", age = 30, items = {"apple", "banana"}}
local jsonStr, err = json.encode(data)
if not jsonStr then
    print("Encode error:", err)
end

-- Decode JSON string to Lua table
local decoded, err = json.decode('{"name":"John","age":30}')
if not decoded then
    print("Decode error:", err)
else
    print(decoded.name) -- "John"
end
```

### `file` - File System Operations

Provides safe file system operations restricted to the config directory for security.

#### Functions

- `file.read(filename)` - Reads entire file content as string
- `file.write(filename, content)` - Writes content to file (overwrites)
- `file.append(filename, content)` - Appends content to file
- `file.exists(filename)` - Returns true if file exists
- `file.list(directory)` - Returns array of file/directory info
- `file.size(filename)` - Returns file size in bytes
- `file.is_dir(path)` - Returns true if path is a directory

#### Example

```lua
local file = require("file")

-- Read a file
local content, err = file.read("mydata.txt")
if not content then
    print("Read error:", err)
end

-- Write to a file
local success, err = file.write("output.txt", "Hello World")
if not success then
    print("Write error:", err)
end

-- Check if file exists and get info
if file.exists("config.txt") then
    print("Config file exists, size:", file.size("config.txt"))
end

-- List directory contents
local items, err = file.list(".")
if items then
    for i, item in ipairs(items) do
        print(item.name, item.is_dir and "DIR" or "FILE", item.size)
    end
end
```

**Security Note:** All file operations are restricted to the config directory and its subdirectories.

### `time` - Time and Date Utilities

Provides enhanced time and date manipulation functions for scheduling and timestamps.

#### Functions

- `time.now()` - Returns current Unix timestamp
- `time.format(timestamp, format)` - Formats timestamp using strftime format
- `time.parse(format, string)` - Parses time string into timestamp
- `time.date([timestamp])` - Returns date components table
- `time.sleep(milliseconds)` - Pauses execution for specified milliseconds

#### Example

```lua
local time = require("time")

-- Get current timestamp
local now = time.now()
print("Current timestamp:", now)

-- Format timestamp
local formatted = time.format(now, "%Y-%m-%d %H:%M:%S")
print("Formatted:", formatted)

-- Parse time string
local timestamp, err = time.parse("%Y-%m-%d", "2024-01-15")
if timestamp then
    print("Parsed timestamp:", timestamp)
end

-- Get date components
local dateInfo = time.date() -- Current time
-- or
local dateInfo = time.date(timestamp) -- Specific timestamp

print("Year:", dateInfo.year)
print("Month:", dateInfo.month)
print("Day:", dateInfo.day)
print("Hour:", dateInfo.hour)
print("Minute:", dateInfo.minute)
print("Second:", dateInfo.second)

-- Sleep for 1 second
time.sleep(1000)
```

### `log` - Structured Logging

Provides structured logging with different levels for debugging and monitoring.

#### Functions

- `log.info(message)` - General information
- `log.warn(message)` - Warning messages
- `log.error(message)` - Error messages
- `log.debug(message)` - Debug information
- `log.printf(format, ...)` - Formatted logging

#### Example

```lua
local log = require("log")

-- Different log levels
log.info("Application started")
log.warn("Warning: low disk space")
log.error("Failed to connect to database")
log.debug("Processing item ID: 12345")

-- Formatted logging
log.printf("User %s logged in from %s", username, ip_address)
```

## Configuration File

The application now supports configuration via `config.yml` in the config directory. This file allows you to customize various aspects of the application:

```yaml
# Application settings
application:
  brightness: 75          # Screen brightness (0-100)
  passive_fps: 2          # Passive update frequency
  debug: false            # Enable debug logging

# Device settings
device:
  auto_detect: true       # Auto-detect Stream Deck
  path: ""               # Manual device path
  model: ""              # Force specific model

# Performance settings
performance:
  image_cache_size: 50    # MB
  compress_images: true
  jpeg_quality: 90

# And many more options...
```

## Migration Guide

If you have existing scripts, they will continue to work unchanged. The new modules are additions that provide more capabilities:

- **Use `json`** for data persistence instead of manual string formatting
- **Use `file`** for reading/writing configuration files safely
- **Use `time`** for better date/time handling and scheduling
- **Use `log`** for consistent logging output instead of `print()`

## Best Practices

1. **Error Handling**: Always check return values for errors, especially with file and JSON operations
2. **Security**: File operations are sandboxed to the config directory - this is intentional for safety
3. **Performance**: Be mindful of file I/O and network operations in passive functions that run frequently
4. **Configuration**: Use the config file for user-customizable settings rather than hardcoding values in scripts

## Examples

### Persistent Counter

```lua
local file = require("file")
local json = require("json")

-- Load counter from file
local counter = 0
local data, err = file.read("counter.json")
if data then
    local decoded, err = json.decode(data)
    if decoded then
        counter = decoded.count or 0
    end
end

function trigger(state)
    counter = counter + 1
    print("Counter:", counter)

    -- Save to file
    local data = json.encode({count = counter})
    file.write("counter.json", data)
end
```

### Scheduled Tasks

```lua
local time = require("time")
local log = require("log")

function background(state)
    while true do
        local now = time.date()
        -- Do something every hour at :00
        if now.minute == 0 then
            log.info("Hourly task executed")
            -- Perform maintenance task
        end
        time.sleep(60000) -- Check every minute
    end
end
```