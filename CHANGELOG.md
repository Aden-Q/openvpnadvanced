# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2024-03-21

### Added
- Added `trace.go` tool for domain resolution and routing information tracing
- Implemented tabular output for network and routing information
- Added automatic route detection and fixing capabilities
- Added CNAME chain information display

### Changed
- Replaced manual table drawing with `tablewriter` library for professional output
- Converted all prompt text from Chinese to English for better internationalization
- Improved error messages for better clarity and specificity
- Enhanced route status checking logic with more accurate routing suggestions

### Fixed
- Fixed duplicate declaration of `GetRouteInterface` function
- Fixed type conversion issues for proper display of `shouldRoute` as "VPN" or "DIRECT"
- Fixed unused variable warnings
- Fixed conditional logic in route checking

### Technical Details
- Implemented table output using `github.com/olekukonko/tablewriter`
- Added DNS caching mechanism with 10-minute TTL
- Added support for recursive CNAME record processing
- Added detailed network interface and routing information
- Implemented automatic route fixing functionality

### Features
1. Network Information Display
   - Domain resolution results
   - IP address
   - Matched rules
   - CNAME chain

2. Routing Information Display
   - Current routing interface
   - VPN interface
   - Default gateway
   - Default gateway interface

3. Route Status Checking
   - Automatic route problem detection
   - Repair suggestions
   - Automatic route fixing
   - Repair result display

### Improvements
- Clearer output format
- More professional error handling
- More accurate route detection
- Better user experience

## [1.1.0] - 2024-03-20

### Added
- Interactive command-line console (`ovpnctl`) with full command set
- Background `start` and verbose `startv` core logic launch
- Real-time `view-log` filtering with `info`, `err`, `vpn` levels
- VPN interface detection and recovery
- `rtest` and `test` commands for rule match and routing inspection
- Live `status`, `show-iface`, `reload-config`, `clear-log`, `clear` commands
- Auto `logs/` directory creation, multiple log file outputs
- Tab-based command completion, up/down command history

### Fixed
- Corrected route cleanup and VPN routing visibility

## [1.0.0] - 2024-03-19

### Initial Release
- DNS proxy
- VPN route management
