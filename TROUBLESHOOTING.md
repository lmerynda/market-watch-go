# Troubleshooting Guide

## Common Issues and Solutions

### 🔄 Loading Issues

**Symptoms**: Data doesn't load or refresh button gets stuck.

**Solutions**:
1. **Use Cancel button** - Click the "Cancel" button next to the refresh button if loading gets stuck
2. **Hard refresh** the browser (Ctrl+F5 or Cmd+Shift+R)
3. **Check browser console** (F12) for JavaScript errors
4. **Try individual refresh** - The app remains usable during loading

**New Feature**: Non-blocking loading with spinner in refresh button and cancel option - app stays usable while loading!

### 📊 No Chart Data Showing

**Symptoms**: Charts are empty or show "No data available"

**Possible Causes**:
1. **Market is closed** - Data is only collected during market hours (9:30 AM - 4:00 PM ET)
2. **No historical data** - App just started and hasn't collected enough data yet
3. **API rate limiting** - Free tier Polygon.io has limited requests

**Solutions**:
1. **Force collection**: Click the "Force Update" button in the dashboard
2. **Wait for market hours**: Data collection is automatic during trading hours
3. **Check collection status**: Look at the collection status panel on the dashboard
4. **Collect historical data**: Start the app with `-historical 7` flag

### 🔑 API Key Issues

**Symptoms**: "API key is required" error on startup

**Solutions**:
1. **Edit .env file**: Replace `your_polygon_api_key_here` with your actual API key
2. **Get API key**: Visit [polygon.io](https://polygon.io/), sign up, go to Dashboard → API Keys
3. **Verify API key**: Make sure there are no extra spaces or quotes around the key

### 🌐 Port Already in Use

**Symptoms**: "bind: address already in use" error

**Solutions**:
1. **Change port**: Edit `SERVER_PORT=8081` in .env file
2. **Kill existing process**: `lsof -ti:8080 | xargs kill -9`
3. **Use different port**: `SERVER_PORT=3000 go run cmd/server/main.go`

### 📡 DELAYED API Responses

**Symptoms**: Seeing "DELAYED" in logs

**This is normal!** Free tier Polygon.io accounts receive delayed data. The app handles this gracefully.

### 🗄️ Database Issues

**Symptoms**: Database connection errors

**Solutions**:
1. **Check permissions**: Ensure `data/` directory is writable
2. **Create directory**: `mkdir -p data`
3. **Check disk space**: Ensure sufficient disk space for SQLite database

### 🚀 Performance Issues

**Symptoms**: Slow loading, high CPU usage

**Solutions**:
1. **Reduce symbols**: Edit `TRACKED_SYMBOLS` in .env to fewer symbols
2. **Increase interval**: Change `COLLECTION_INTERVAL=10m` for less frequent updates
3. **Clear old data**: Restart the app to trigger cleanup

## Debug Endpoints

Use these endpoints to diagnose issues:

- **Health Check**: `GET /api/health`
- **Data Count**: `GET /api/debug/count`
- **Collection Status**: `GET /api/collection/status`
- **Force Collection**: `POST /api/collection/force`

## Browser Console Debugging

1. **Open Developer Tools** (F12)
2. **Go to Console tab**
3. **Look for error messages** in red
4. **Check Network tab** for failed API requests

## Log Analysis

The application logs show:
- ✅ Successful API calls and data collection
- ⚠️ DELAYED responses (normal for free tier)
- ❌ Errors and failures
- 📊 Data collection statistics

## Getting Help

If issues persist:

1. **Check the logs** for specific error messages
2. **Verify API key** is valid and has quota remaining
3. **Test individual endpoints** using curl or browser
4. **Try with historical data**: `go run cmd/server/main.go -historical 1`

## Reset Everything

If all else fails:

```bash
# Stop the application
# Delete database
rm data/market-watch.db

# Restart
go run cmd/server/main.go -historical 1
```

This will start fresh with 1 day of historical data.
