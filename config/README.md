# Config Package

This package provides configuration management for the CentralHub application.

## Features

- JSON-based configuration
- Command-line flag support for config file path
- Configuration validation
- Default fallback values
- Global config access

## Usage

### 1. Create Configuration File

Copy the example configuration file:

```bash
cp config.json.example config.json
```

Then edit `config.json` with your settings.

### 2. Configuration Structure

The configuration file supports the following sections:

- **server**: Server settings (port, mode, timeout)
- **database**: Database configuration (MongoDB connection)
- **logger**: Logging configuration (level, output, file settings)
- **external**: External service configurations (Volcengine credentials)

### 3. Running the Application

#### Using default config file (config.json)
```bash
./centralhub
```

#### Using custom config file
```bash
./centralhub -config=/path/to/your/config.json
```

### 4. Configuration Priority

1. If a config file is specified and exists, it will be loaded
2. If the config file is not found or invalid, default values will be used
3. The application will log whether config was loaded successfully or defaults were used

### 5. Accessing Configuration

The loaded configuration is available globally:

```go
import "centralHub/config"

// Access the global config
port := config.GlobalConfig.Server.Port
mongoURI := config.GlobalConfig.Database.MongoDB.URI
```

### 6. Environment Modes

The server supports three modes:
- `debug`: Development mode with verbose logging
- `release`: Production mode with optimized performance
- `test`: Testing mode

## Example Configuration

Below is a sample `config.json`:

```json
{
  "server": {
    "port": "8080",
    "mode": "debug",
    "timeout": 30
  },
  "database": {
    "mongodb": {
      "uri": "mongodb://localhost:27017",
      "database": "centralhub",
      "timeout": 10
    }
  },
  "logger": {
    "level": "info",
    "output": "stdout",
    "file_path": "logs/app.log",
    "max_size": 100,
    "max_backups": 3,
    "max_age": 7
  },
  "external": {
    "volcengine": {
      "access_key": "your-access-key",
      "secret_key": "your-secret-key",
      "region": "cn-beijing"
    }
  }
}
```
