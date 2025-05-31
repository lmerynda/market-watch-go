# Quick Start Guide

## 🚀 Get Started in 3 Steps

### Step 1: Get Your Free Polygon.io API Key

1. **Visit** [polygon.io](https://polygon.io/)
2. **Sign up** for a free account
3. **Go to** Dashboard → API Keys
4. **Copy** your API key

### Step 2: Set Up the Application

```bash
# Clone and setup
./setup.sh

# Edit .env file and add your API key
nano .env  # or use your preferred editor
```

Replace `your_polygon_api_key_here` with your actual API key:
```
POLYGON_API_KEY=your_actual_api_key_from_polygon_io
```

### Step 3: Run the Application

The application automatically loads the .env file, so you can run it directly:

```bash
# Start the server (automatically loads .env)
go run cmd/server/main.go

# Or use the built binary
./bin/market-watch
```

**No need to source the .env file** - the application handles it automatically!

🎉 **That's it!** Open http://localhost:8080 in your browser.

## 📊 What You'll See

- **Real-time volume charts** for PLTR, TSLA, BBAI, MSFT, NPWR
- **Interactive dashboard** with multiple time ranges
- **Volume statistics** and market status
- **Auto-refreshing data** every 30 seconds

## 🔧 Optional: Collect Historical Data

```bash
# Collect 7 days of historical data before starting
go run cmd/server/main.go -historical 7
```

## ❓ Having Issues?

1. **API Key Error**: Make sure you've replaced the placeholder in .env
2. **Port in Use**: Change SERVER_PORT in .env to a different port
3. **No Data**: Market might be closed (data only collected 9:30 AM - 4:00 PM ET)

## 📚 Next Steps

- Check out the full [README.md](README.md) for detailed documentation
- Explore the API at http://localhost:8080/api/health
- Monitor collection status in the dashboard

---

**Note**: The free Polygon.io tier includes 5 API calls per minute, which is perfect for this application's 5-minute collection interval.
