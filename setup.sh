#!/bin/bash

# Market Watch Go - Setup Script

echo "🚀 Setting up Market Watch Go..."

# Create necessary directories
echo "📁 Creating directories..."
mkdir -p bin data web/static/lib

# Build the application
echo "🔨 Building application..."
go mod tidy
go build -o bin/market-watch cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed!"
    exit 1
fi

# Check if API key is set
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
fi

# Check if API key needs to be set
if grep -q "your_polygon_api_key_here" .env; then
    echo ""
    echo "🔑 SETUP REQUIRED:"
    echo "Please edit .env file and replace 'your_polygon_api_key_here' with your actual Polygon.io API key"
    echo ""
    echo "💡 Get a free API key at: https://polygon.io/"
    echo "   1. Sign up for a free account"
    echo "   2. Go to Dashboard -> API Keys"
    echo "   3. Copy your API key and paste it in .env file"
    echo ""
    echo "Then run the application with:"
    echo "   go run cmd/server/main.go"
    echo ""
    exit 0
fi

# Set executable permissions
chmod +x bin/market-watch

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file and add your Polygon.io API key"
echo "2. Run the application:"
echo "   ./bin/market-watch"
echo ""
echo "Or run with Go:"
echo "   go run cmd/server/main.go"
echo ""
echo "Dashboard will be available at: http://localhost:8080"
echo "API will be available at: http://localhost:8080/api"
echo ""
echo "For historical data collection, run:"
echo "   ./bin/market-watch -historical 7"
echo ""
