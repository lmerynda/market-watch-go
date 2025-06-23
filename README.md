# Market Watch Go - Volume Tracker

A real-time stock volume tracking web application similar to unusualwhales.com functionality. This application fetches trading volume data from Polygon.io API and displays it in interactive charts.

---

## ⚡️ Moving Averages: Now Using EMA (Exponential Moving Average)

- **SMA (Simple Moving Average) is deprecated.**
- All moving average calculations and watchlist fields now use EMA (9, 50, 200) instead of SMA.
- The database schema and API have been updated to reflect this change.
- New endpoints for querying EMA values directly from Polygon.io:
  - `GET /api/polygon/ema?symbol=TSLA&window=9` (single window)
  - `GET /api/polygon/ema/batch?symbol=TSLA&windows=9,50,200` (multiple windows)

---

## Features

- **Real-time Volume Tracking**: Monitors volume data for selected stocks
- **Interactive Charts**: Chart.js-powered volume visualization with multiple time ranges
- **Automated Data Collection**: Scheduled data fetching every 5 minutes during market hours
- **Local Data Storage**: SQLite database for reliable data persistence
- **Responsive Dashboard**: Clean, mobile-friendly web interface
- **REST API**: Complete API for programmatic access to volume data
- **Health Monitoring**: Built-in health checks and collection status monitoring

## Architecture

- **Backend**: Go with Gin web framework
- **Database**: SQLite for local data storage
- **Frontend**: HTML5, Bootstrap 5, Chart.js
- **Data Source**: Polygon.io REST API
- **Scheduling**: Cron-based automated data collection

## Prerequisites

- Go 1.19 or later
- Polygon.io API key (free tier available)
- Modern web browser

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd market-watch-go
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and add your Polygon.io API key:
   ```
   POLYGON_API_KEY=your_polygon_api_key_here
   ```
   
   **Note**: The application automatically loads the `.env` file - no need to source it manually!

4. **Create data directory**
   ```bash
   mkdir -p data
   ```

## Configuration

The application uses a YAML configuration file at `configs/config.yaml`. Key settings include:

- **Server**: Port, host, timeouts
- **Database**: SQLite path and connection settings
- **Polygon API**: Base URL, timeout, retry settings
- **Collection**: Data fetching interval, tracked symbols, market hours
- **Data Retention**: How long to keep historical data

Environment variables override configuration file settings.

## Usage

### Basic Startup

```bash
go run cmd/server/main.go
```

### With Configuration File

```bash
go run cmd/server/main.go -config configs/config.yaml
```

### Collect Historical Data

```bash
go run cmd/server/main.go -historical 7
```

This will collect 7 days of historical data before starting the server.

### Command Line Options

- `-config`: Path to configuration file (default: `configs/config.yaml`)
- `-env`: Path to environment file
- `-historical`: Number of days of historical data to collect (0 = disabled)

## API Endpoints

### Volume Data
- `GET /api/volume/{symbol}` - Get volume data for a symbol
- `GET /api/volume/{symbol}/latest` - Get latest volume data
- `GET /api/volume/{symbol}/chart?range=1D` - Get chart data

### Dashboard
- `GET /api/dashboard/summary` - Get dashboard summary with all symbols
- `GET /` - Main dashboard interface

### Collection Management
- `GET /api/collection/status` - Get collection service status
- `POST /api/collection/force` - Force immediate data collection

### Health Check
- `GET /api/health` - Application health status

### Query Parameters

**Volume Data Endpoints:**
- `from`: Start date (YYYY-MM-DD or RFC3339)
- `to`: End date (YYYY-MM-DD or RFC3339)
- `limit`: Maximum number of records (default: 1000, max: 10000)
- `offset`: Number of records to skip
- `interval`: Data interval (5m, 15m, 1h, 4h, 1d)

**Chart Endpoints:**
- `range`: Time range (1H, 4H, 1D, 1W, 1M)

## Web Dashboard

Access the dashboard at `http://localhost:8080` (or your configured address).

### Dashboard Features

- **Volume Charts**: Interactive time-series charts for each tracked symbol
- **Time Range Selection**: 1H, 4H, 1D, 1W views
- **Volume Statistics**: Current volume, average volume, volume ratios
- **Market Status**: Real-time market open/closed indicator
- **Collection Status**: Monitor data collection health and statistics
- **Auto-refresh**: Automatic data updates every 30 seconds

### Dashboard Controls

- **Time Range Buttons**: Switch between different time periods
- **Refresh Button**: Manually refresh all data
- **Force Update Button**: Trigger immediate data collection

## Data Collection

The application automatically:

1. **Fetches Data**: Every 5 minutes during market hours (9:30 AM - 4:00 PM ET)
2. **Stores Locally**: Saves all data to SQLite database
3. **Handles Duplicates**: Prevents duplicate data insertion
4. **Respects Rate Limits**: Includes delays between API calls
5. **Cleans Up**: Removes old data based on retention policy (default: 30 days)

### Market Hours

- **Trading Days**: Monday - Friday
- **Trading Hours**: 9:30 AM - 4:00 PM Eastern Time
- **Timezone Handling**: Automatic timezone conversion

## Database Schema

### volume_data table
```sql
CREATE TABLE volume_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    volume INTEGER NOT NULL,
    price DECIMAL(10,2),
    open_price DECIMAL(10,2),
    high_price DECIMAL(10,2),
    low_price DECIMAL(10,2),
    close_price DECIMAL(10,2),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, timestamp)
);
```

## Error Handling

- **API Failures**: Graceful degradation when Polygon API is unavailable
- **Rate Limiting**: Automatic delays to respect API limits
- **Database Errors**: Comprehensive error logging and recovery
- **Network Issues**: Retry logic with exponential backoff

## Monitoring

### Health Checks

The `/api/health` endpoint provides comprehensive health information:

```json
{
  "status": "ok",
  "timestamp": "2025-05-31T10:00:00Z",
  "services": {
    "database": {"status": "ok"},
    "polygon": {"status": "ok"},
    "collector": {"status": "ok"}
  }
}
```

### Collection Statistics

Monitor data collection via `/api/collection/status`:

```json
{
  "last_run": "2025-05-31T10:00:00Z",
  "next_run": "2025-05-31T10:05:00Z",
  "successful_runs": 120,
  "failed_runs": 2,
  "collected_today": 1440,
  "is_running": false
}
```

## Development

### Project Structure

```
market-watch-go/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── config/config.go           # Configuration management
│   ├── database/sqlite.go         # Database operations
│   ├── handlers/                  # HTTP handlers
│   ├── models/                    # Data models
│   └── services/                  # Business logic
├── web/
│   ├── static/                    # CSS, JS assets
│   └── templates/                 # HTML templates
├── configs/config.yaml            # Configuration file
└── data/                         # SQLite database
```

### Adding New Symbols

1. Update `configs/config.yaml`:
   ```yaml
   collection:
     symbols:
       - "PLTR"
       - "TSLA"
       - "NEW_SYMBOL"
   ```

2. Update frontend template to include new symbol
3. Add color scheme in `models/volume.go`

### Building for Production

```bash
# Build binary
go build -o bin/market-watch cmd/server/main.go

# Run binary
./bin/market-watch -config configs/config.yaml
```

## Troubleshooting

### Common Issues

1. **API Key Invalid**
   - Verify your Polygon.io API key is correct
   - Check if you have sufficient API quota

2. **Database Permission Issues**
   - Ensure the `data/` directory is writable
   - Check file permissions on the SQLite database

3. **Port Already in Use**
   - Change the port in `configs/config.yaml` or set `SERVER_PORT` environment variable

4. **No Data Appearing**
   - Check if market is open (data only collected during market hours)
   - Verify API key has access to the symbols
   - Check collection status via `/api/collection/status`

### Debug Mode

Run with debug logging:
```bash
LOG_LEVEL=debug go run cmd/server/main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Future Enhancements

- **Volume Deviation Alerts**: Email/SMS notifications for unusual volume
- **More Symbols**: Dynamic symbol management
- **Advanced Analytics**: Volume profile analysis, correlation studies
- **WebSocket Updates**: Real-time streaming data
- **User Authentication**: Multi-user support
- **Export Features**: CSV/Excel data export
- **Technical Indicators**: Moving averages, RSI, etc.

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review application logs
3. Check API endpoint health
4. Create an issue in the repository

---

**Note**: This application is for educational and research purposes. Always verify trading data from official sources before making investment decisions.
