// TradingView-Style Market Watch Dashboard
class MarketWatchDashboard {
    constructor() {
        this.charts = {};
        this.symbols = [];
        this.currentTimeRange = '1D';
        this.refreshInterval = null;
        this.updateInterval = 30000; // 30 seconds
        this.isLoading = false;
        this.loadingController = null;
        this.retryCount = {
            technicalAnalysis: 0,
            tradingSetups: 0,
            supportResistance: 0
        };
        
        // TradingView color palette
        this.colors = {
            background: '#131722',
            surface: '#1e222d',
            surfaceLight: '#2a2e39',
            border: '#363a45',
            textPrimary: '#d1d4dc',
            textSecondary: '#787b86',
            textMuted: '#5d606b',
            
            // Chart colors
            candleUp: '#26a69a',
            candleDown: '#ef5350',
            volumeUp: 'rgba(38, 166, 154, 0.6)',
            volumeDown: 'rgba(239, 83, 80, 0.6)',
            
            // Technical indicators
            sma20: '#ff9800',
            sma50: '#2196f3',
            sma200: '#9c27b0',
            
            // Grid and axis
            grid: 'rgba(120, 123, 134, 0.1)',
            axis: '#363a45'
        };
        
        this.init();
    }

    async init() {
        console.log('Dashboard init() started');
        
        console.log('Loading symbols...');
        await this.loadSymbols();
        console.log('Symbols loaded:', this.symbols);
        
        console.log('Setting up event listeners...');
        this.setupEventListeners();
        
        console.log('Creating charts container...');
        this.createChartsContainer();
        
        console.log('Initializing charts...');
        this.initializeCharts();
        
        await new Promise(resolve => setTimeout(resolve, 100));
        
        this.syncTimeRangeFromHTML();
        console.log('Dashboard initialized with time range:', this.currentTimeRange);
        
        console.log('Loading dashboard data...');
        this.loadDashboardData();
        
        console.log('Starting auto refresh...');
        this.startAutoRefresh();
        
        console.log('Dashboard initialization complete');
    }

    async loadSymbols() {
        try {
            // First, quick health check to see if backend is running
            console.log('🔍 Testing backend connectivity...');
            try {
                const healthResponse = await fetch('/health', {
                    signal: AbortSignal.timeout(3000)
                });
                console.log('✅ Backend health check:', healthResponse.status);
            } catch (healthError) {
                console.error('❌ Backend appears to be down:', healthError.message);
                this.symbols = [];
                this.showNoSymbolsMessage();
                this.loadEmptyDashboardSections();
                return;
            }
            
            console.log('📋 Fetching symbols from /api/symbols...');
            
            // Add timeout to detect hanging requests
            const controller = new AbortController();
            const timeoutId = setTimeout(() => {
                controller.abort();
                console.error('❌ /api/symbols request timed out after 10 seconds');
            }, 10000);
            
            const response = await fetch('/api/symbols', {
                signal: controller.signal
            });
            
            clearTimeout(timeoutId);
            console.log('Symbols API response status:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                console.log('Raw symbols API response:', data);
                console.log('Response type:', typeof data);
                console.log('Is array?', Array.isArray(data));
                
                // Handle the API response structure: {symbols: [...], count: N}
                if (data && data.symbols && Array.isArray(data.symbols) && data.symbols.length > 0) {
                    this.symbols = data.symbols.map(symbolObj => symbolObj.symbol);
                    console.log('✅ Loaded symbols from database (structure):', this.symbols);
                } else if (data && Array.isArray(data) && data.length > 0) {
                    // Fallback for direct array response
                    this.symbols = data.map(symbol => symbol.symbol || symbol);
                    console.log('✅ Loaded symbols from database (direct array):', this.symbols);
                } else {
                    console.warn('❌ No symbols returned from API - database may be empty');
                    console.log('Data structure received:', data);
                    this.symbols = [];
                    this.showNoSymbolsMessage();
                    // Still try to load other dashboard sections even without symbols
                    this.loadEmptyDashboardSections();
                }
            } else {
                console.error('❌ Failed to load symbols from API. Status:', response.status);
                const errorText = await response.text();
                console.error('Error response:', errorText);
                this.symbols = [];
                this.showNoSymbolsMessage();
                // Still try to load other dashboard sections even without symbols
                this.loadEmptyDashboardSections();
            }
        } catch (error) {
            console.error('❌ Error loading symbols:', error);
            this.symbols = [];
            this.showNoSymbolsMessage();
            // Still try to load other dashboard sections even without symbols
            this.loadEmptyDashboardSections();
        }
        
        console.log('Final symbols array:', this.symbols);
        console.log('Symbols count:', this.symbols.length);
    }

    showNoSymbolsMessage() {
        console.log('📢 showNoSymbolsMessage() called');
        const chartsGrid = document.getElementById('charts-grid');
        console.log('Charts grid element for no symbols message:', !!chartsGrid);
        
        if (chartsGrid) {
            const noSymbolsHtml = `
                <div class="col-12">
                    <div class="alert alert-warning">
                        <h5><i class="bi bi-exclamation-triangle me-2"></i>No Symbols Configured</h5>
                        <p>No symbols are currently being watched. Please add symbols using:</p>
                        <ul>
                            <li>The API endpoint: <code>POST /api/symbols</code></li>
                            <li>Or add symbols to the <code>configs/config.yaml</code> file and restart</li>
                        </ul>
                    </div>
                </div>
            `;
            chartsGrid.innerHTML = noSymbolsHtml;
            console.log('✅ No symbols message displayed in charts grid');
            console.log('Charts grid HTML after setting message:', chartsGrid.innerHTML.substring(0, 100) + '...');
        } else {
            console.error('❌ Could not display no symbols message - charts grid not found');
        }
    }

    loadEmptyDashboardSections() {
        console.log('Loading empty dashboard sections...');
        
        // Initialize tab containers with empty states
        setTimeout(() => {
            this.initializeEmptyTechnicalAnalysis();
            this.initializeEmptyTradingSetups();
            this.initializeEmptySupportResistance();
        }, 100);
    }

    initializeEmptyTechnicalAnalysis() {
        const container = document.getElementById('technical-indicators');
        if (container) {
            container.innerHTML = '<div class="text-center text-tv-muted">No symbols configured for technical analysis</div>';
            console.log('Technical analysis container initialized with empty state');
        } else {
            console.warn('Technical indicators container still not found');
        }
    }

    initializeEmptyTradingSetups() {
        const container = document.getElementById('trading-setups');
        if (container) {
            container.innerHTML = '<div class="text-center text-tv-muted">No symbols configured for trading setups</div>';
            console.log('Trading setups container initialized with empty state');
        } else {
            console.warn('Trading setups container still not found');
        }
    }

    initializeEmptySupportResistance() {
        const container = document.getElementById('support-resistance');
        if (container) {
            container.innerHTML = '<div class="text-center text-tv-muted">No symbols configured for support/resistance analysis</div>';
            console.log('Support/resistance container initialized with empty state');
        } else {
            console.warn('Support/resistance container still not found');
        }
    }

    setupEventListeners() {
        // Time range selector
        document.querySelectorAll('input[name="timeRange"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.currentTimeRange = e.target.value;
                this.refreshCharts();
            });
        });

        // Control buttons
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadDashboardData();
        });

        document.getElementById('cancel-refresh-btn').addEventListener('click', () => {
            this.cancelLoading();
        });

        document.getElementById('force-collection-btn').addEventListener('click', () => {
            this.forceCollection();
        });

        document.getElementById('manage-symbols-btn').addEventListener('click', () => {
            this.toggleSymbolManagement();
        });

        // Window resize handler
        window.addEventListener('resize', () => {
            this.resizeCharts();
        });
    }

    syncTimeRangeFromHTML() {
        const checkedRadio = document.querySelector('input[name="timeRange"]:checked');
        if (checkedRadio) {
            this.currentTimeRange = checkedRadio.value;
        }
    }

    createChartsContainer() {
        console.log('📊 createChartsContainer() called');
        const chartsGrid = document.getElementById('charts-grid');
        console.log('Charts grid element found:', !!chartsGrid);
        
        if (!chartsGrid) {
            console.error('❌ charts-grid element not found!');
            return;
        }
        
        console.log('📋 Creating chart containers for symbols:', this.symbols);
        console.log('📊 Symbols count:', this.symbols.length);
        
        // Clear existing content first
        chartsGrid.innerHTML = '';
        console.log('🗑️ Cleared existing chart grid content');
        
        if (this.symbols.length === 0) {
            console.warn('⚠️ No symbols available to create charts for');
            console.log('Showing no symbols message...');
            return;
        }
        
        console.log('🔨 Building chart containers...');
        
        // Create chart containers for each symbol
        this.symbols.forEach((symbol, index) => {
            console.log(`📈 Creating chart container ${index + 1}/${this.symbols.length} for symbol: ${symbol}`);
            const chartContainer = document.createElement('div');
            chartContainer.className = 'col-lg-6 col-xl-4 mb-4';
            chartContainer.innerHTML = `
                <div class="card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h6 class="mb-0">${symbol}</h6>
                        <div class="d-flex align-items-center">
                            <span id="${symbol}-current-price" class="badge bg-primary me-2">--</span>
                            <span id="${symbol}-price-change" class="badge bg-secondary">--</span>
                        </div>
                    </div>
                    <div class="card-body p-2">
                        <div class="financial-chart-container">
                            <svg id="chart-${symbol}" width="100%" height="350"></svg>
                        </div>
                        <div class="mt-2 px-2">
                            <small class="text-muted">
                                <span class="price-display">
                                    O: <span id="${symbol}-open" class="text-tv-secondary">--</span>
                                    H: <span id="${symbol}-high" class="text-success">--</span>
                                    L: <span id="${symbol}-low" class="text-danger">--</span>
                                    C: <span id="${symbol}-close" class="text-tv-primary">--</span>
                                </span>
                                <span class="ms-3">
                                    Vol: <span id="${symbol}-volume" class="text-tv-secondary">--</span>
                                </span>
                            </small>
                        </div>
                    </div>
                </div>
            `;
            chartsGrid.appendChild(chartContainer);
            console.log(`✅ Added chart container for ${symbol} to DOM`);
        });
        
        console.log(`🎉 Finished creating ${this.symbols.length} chart containers`);
        console.log('Final charts grid HTML length:', chartsGrid.innerHTML.length);
    }

    initializeCharts() {
        this.symbols.forEach(symbol => {
            this.createTradingViewChart(symbol);
        });
    }

    createTradingViewChart(symbol) {
        const svgElement = document.getElementById(`chart-${symbol}`);
        if (!svgElement) return;

        // Clear any existing content
        d3.select(svgElement).selectAll("*").remove();

        const container = svgElement.parentElement;
        const margin = { top: 20, right: 50, bottom: 50, left: 60 };
        const width = container.clientWidth - margin.left - margin.right;
        const height = 350 - margin.top - margin.bottom;

        const svg = d3.select(svgElement)
            .attr("width", width + margin.left + margin.right)
            .attr("height", height + margin.top + margin.bottom)
            .style("background", this.colors.surface);

        const g = svg.append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`);

        // Add chart background
        g.append("rect")
            .attr("width", width)
            .attr("height", height)
            .attr("fill", this.colors.surface)
            .attr("stroke", this.colors.border)
            .attr("stroke-width", 1);

        // Store chart components
        this.charts[symbol] = {
            svg,
            g,
            width,
            height,
            margin
        };
    }

    async loadDashboardData() {
        if (this.isLoading) return;

        console.log('Loading dashboard data...');
        this.showLoading();
        this.loadingController = new AbortController();
        
        try {
            // Load collection status
            await this.loadCollectionStatus();
            
            // Load chart data for all symbols
            await this.loadAllChartData();
            
            // Load technical analysis
            await this.loadTechnicalAnalysis();
            
            // Load trading setups
            await this.loadTradingSetups();
            
            // Load support/resistance
            await this.loadSupportResistance();
            
            this.updateLastUpdateTime();
            console.log('Dashboard data loaded successfully');
        } catch (error) {
            if (error.name === 'AbortError') {
                console.log('Dashboard loading cancelled');
                return;
            }
            console.error('Failed to load dashboard data:', error);
            this.showError('Failed to load dashboard data: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    async loadAllChartData() {
        const promises = this.symbols.map(symbol => this.loadChartData(symbol));
        await Promise.all(promises);
    }

    async loadChartData(symbol) {
        try {
            const url = `/api/price/${symbol}/chart?range=${this.currentTimeRange}`;
            console.log(`Loading chart data for ${symbol}: ${url}`);
            
            const response = await fetch(url, {
                signal: this.loadingController?.signal
            });
            const data = await response.json();
            
            if (response.ok && data.data && data.data.length > 0) {
                this.updateTradingViewChart(symbol, data.data);
                this.updateSymbolInfo(symbol, data.data);
            } else {
                console.warn(`No chart data for ${symbol}`);
                this.updateTradingViewChart(symbol, []);
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error;
            console.error(`Failed to load chart data for ${symbol}:`, error);
        }
    }

    updateTradingViewChart(symbol, data) {
        const chart = this.charts[symbol];
        if (!chart || !data) return;

        const { g, width, height } = chart;
        
        // Clear previous content
        g.selectAll("*").remove();

        if (data.length === 0) {
            g.append("text")
                .attr("x", width / 2)
                .attr("y", height / 2)
                .attr("text-anchor", "middle")
                .style("fill", this.colors.textMuted)
                .style("font-size", "12px")
                .text("No data available");
            return;
        }

        // Parse data
        const parsedData = data.map(d => ({
            timestamp: new Date(d.timestamp),
            open: d.open_price || d.open,
            high: d.high_price || d.high,
            low: d.low_price || d.low,
            close: d.close_price || d.close,
            volume: d.volume
        }));

        // Create scales
        const xScale = d3.scaleTime()
            .domain(d3.extent(parsedData, d => d.timestamp))
            .range([0, width]);

        const priceExtent = d3.extent(parsedData, d => Math.max(d.high, d.low));
        const yScale = d3.scaleLinear()
            .domain([priceExtent[0] * 0.99, priceExtent[1] * 1.01])
            .range([height - 60, 0]); // Leave space for volume

        const maxVolume = d3.max(parsedData, d => d.volume) || 1;
        const volumeScale = d3.scaleLinear()
            .domain([0, maxVolume])
            .range([height - 50, height - 10])
            .clamp(true); // Prevent values outside the range

        // Add grid
        this.addGrid(g, width, height, xScale, yScale);

        // Add candlesticks
        this.addCandlesticks(g, parsedData, xScale, yScale);

        // Add volume bars
        this.addVolumeBar(g, parsedData, xScale, volumeScale);

        // Add moving averages
        this.addMovingAverages(g, parsedData, xScale, yScale);

        // Add axes
        this.addAxes(g, width, height, xScale, yScale);

        // Add crosshair
        this.addCrosshair(g, width, height, parsedData, xScale, yScale, symbol);
    }

    addGrid(g, width, height, xScale, yScale) {
        // Vertical grid lines
        g.selectAll(".grid-x")
            .data(xScale.ticks(6))
            .enter().append("line")
            .attr("class", "grid-x")
            .attr("x1", d => xScale(d))
            .attr("x2", d => xScale(d))
            .attr("y1", 0)
            .attr("y2", height - 60)
            .attr("stroke", this.colors.grid)
            .attr("stroke-width", 1);

        // Horizontal grid lines
        g.selectAll(".grid-y")
            .data(yScale.ticks(6))
            .enter().append("line")
            .attr("class", "grid-y")
            .attr("x1", 0)
            .attr("x2", width)
            .attr("y1", d => yScale(d))
            .attr("y2", d => yScale(d))
            .attr("stroke", this.colors.grid)
            .attr("stroke-width", 1);
    }

    addCandlesticks(g, data, xScale, yScale) {
        const candleWidth = Math.max(2, (xScale.range()[1] - xScale.range()[0]) / data.length * 0.8);

        // Add wicks
        g.selectAll(".wick")
            .data(data)
            .enter().append("line")
            .attr("class", "wick")
            .attr("x1", d => xScale(d.timestamp))
            .attr("x2", d => xScale(d.timestamp))
            .attr("y1", d => yScale(d.high))
            .attr("y2", d => yScale(d.low))
            .attr("stroke", d => d.close >= d.open ? this.colors.candleUp : this.colors.candleDown)
            .attr("stroke-width", 1);

        // Add candle bodies
        g.selectAll(".candle")
            .data(data)
            .enter().append("rect")
            .attr("class", "candle")
            .attr("x", d => xScale(d.timestamp) - candleWidth / 2)
            .attr("y", d => yScale(Math.max(d.open, d.close)))
            .attr("width", candleWidth)
            .attr("height", d => Math.abs(yScale(d.open) - yScale(d.close)) || 1)
            .attr("fill", d => d.close >= d.open ? this.colors.candleUp : this.colors.candleDown)
            .attr("stroke", d => d.close >= d.open ? this.colors.candleUp : this.colors.candleDown)
            .attr("stroke-width", 1);
    }

    addVolumeBar(g, data, xScale, volumeScale) {
        const barWidth = Math.max(1, (xScale.range()[1] - xScale.range()[0]) / data.length * 0.6);

        g.selectAll(".volume")
            .data(data)
            .enter().append("rect")
            .attr("class", "volume")
            .attr("x", d => xScale(d.timestamp) - barWidth / 2)
            .attr("y", d => {
                // Ensure volume is a valid number
                const volume = d.volume || 0;
                return Math.min(volumeScale(volume), volumeScale(0));
            })
            .attr("width", barWidth)
            .attr("height", d => {
                // Ensure volume is a valid number and height is never negative
                const volume = d.volume || 0;
                const height = Math.abs(volumeScale(0) - volumeScale(volume));
                return Math.max(0, height);
            })
            .attr("fill", d => d.close >= d.open ? this.colors.volumeUp : this.colors.volumeDown)
            .attr("opacity", 0.7);
    }

    addMovingAverages(g, data, xScale, yScale) {
        if (data.length < 20) return;

        // Calculate SMA20
        const sma20Data = this.calculateSMA(data, 20);
        if (sma20Data.length > 0) {
            const line20 = d3.line()
                .x(d => xScale(d.timestamp))
                .y(d => yScale(d.sma))
                .curve(d3.curveMonotoneX);

            g.append("path")
                .datum(sma20Data)
                .attr("class", "ma-line sma-20")
                .attr("d", line20)
                .attr("stroke", this.colors.sma20)
                .attr("stroke-width", 1.5)
                .attr("fill", "none")
                .attr("opacity", 0.8);
        }

        // Calculate SMA50 if enough data
        if (data.length >= 50) {
            const sma50Data = this.calculateSMA(data, 50);
            if (sma50Data.length > 0) {
                const line50 = d3.line()
                    .x(d => xScale(d.timestamp))
                    .y(d => yScale(d.sma))
                    .curve(d3.curveMonotoneX);

                g.append("path")
                    .datum(sma50Data)
                    .attr("class", "ma-line sma-50")
                    .attr("d", line50)
                    .attr("stroke", this.colors.sma50)
                    .attr("stroke-width", 1.5)
                    .attr("fill", "none")
                    .attr("opacity", 0.8);
            }
        }
    }

    calculateSMA(data, period) {
        const result = [];
        for (let i = period - 1; i < data.length; i++) {
            const sum = data.slice(i - period + 1, i + 1)
                .reduce((acc, d) => acc + d.close, 0);
            result.push({
                timestamp: data[i].timestamp,
                sma: sum / period
            });
        }
        return result;
    }

    addAxes(g, width, height, xScale, yScale) {
        // X-axis
        g.append("g")
            .attr("class", "axis")
            .attr("transform", `translate(0,${height - 60})`)
            .call(d3.axisBottom(xScale)
                .tickFormat(d3.timeFormat("%H:%M"))
                .tickSize(5))
            .selectAll("text")
            .style("fill", this.colors.textSecondary)
            .style("font-size", "11px");

        // Y-axis (price)
        g.append("g")
            .attr("class", "axis")
            .attr("transform", `translate(${width}, 0)`)
            .call(d3.axisRight(yScale)
                .tickFormat(d => `$${d.toFixed(2)}`)
                .tickSize(5))
            .selectAll("text")
            .style("fill", this.colors.textSecondary)
            .style("font-size", "11px");

        // Style axis paths
        g.selectAll(".axis path, .axis line")
            .style("stroke", this.colors.axis);
    }

    addCrosshair(g, width, height, data, xScale, yScale, symbol) {
        const tooltip = d3.select("body").append("div")
            .attr("class", "chart-tooltip")
            .style("opacity", 0);

        const crosshairX = g.append("line")
            .attr("class", "crosshair-x")
            .attr("stroke", this.colors.textMuted)
            .attr("stroke-width", 1)
            .attr("stroke-dasharray", "3,3")
            .style("opacity", 0);

        const crosshairY = g.append("line")
            .attr("class", "crosshair-y")
            .attr("stroke", this.colors.textMuted)
            .attr("stroke-width", 1)
            .attr("stroke-dasharray", "3,3")
            .style("opacity", 0);

        g.append("rect")
            .attr("width", width)
            .attr("height", height - 60)
            .attr("fill", "transparent")
            .on("mousemove", (event) => {
                const [mouseX, mouseY] = d3.pointer(event);
                const date = xScale.invert(mouseX);
                const price = yScale.invert(mouseY);

                // Find closest data point
                const bisect = d3.bisector(d => d.timestamp).left;
                const index = bisect(data, date, 1);
                const dataPoint = data[index - 1] || data[index];

                if (dataPoint) {
                    crosshairX
                        .attr("x1", mouseX)
                        .attr("x2", mouseX)
                        .attr("y1", 0)
                        .attr("y2", height - 60)
                        .style("opacity", 1);

                    crosshairY
                        .attr("x1", 0)
                        .attr("x2", width)
                        .attr("y1", mouseY)
                        .attr("y2", mouseY)
                        .style("opacity", 1);

                    tooltip.transition().duration(200).style("opacity", 0.9);
                    tooltip.html(`
                        <div class="tooltip-row">
                            <span class="tooltip-label">${symbol}</span>
                            <span class="tooltip-value">${dataPoint.timestamp.toLocaleTimeString()}</span>
                        </div>
                        <div class="tooltip-row">
                            <span class="tooltip-label">O:</span>
                            <span class="tooltip-value">$${dataPoint.open.toFixed(2)}</span>
                        </div>
                        <div class="tooltip-row">
                            <span class="tooltip-label">H:</span>
                            <span class="tooltip-value">$${dataPoint.high.toFixed(2)}</span>
                        </div>
                        <div class="tooltip-row">
                            <span class="tooltip-label">L:</span>
                            <span class="tooltip-value">$${dataPoint.low.toFixed(2)}</span>
                        </div>
                        <div class="tooltip-row">
                            <span class="tooltip-label">C:</span>
                            <span class="tooltip-value">$${dataPoint.close.toFixed(2)}</span>
                        </div>
                        <div class="tooltip-row">
                            <span class="tooltip-label">Vol:</span>
                            <span class="tooltip-value">${this.formatVolume(dataPoint.volume)}</span>
                        </div>
                    `)
                    .style("left", (event.pageX + 10) + "px")
                    .style("top", (event.pageY - 10) + "px");
                }
            })
            .on("mouseout", () => {
                crosshairX.style("opacity", 0);
                crosshairY.style("opacity", 0);
                tooltip.transition().duration(500).style("opacity", 0);
            });
    }

    updateSymbolInfo(symbol, data) {
        if (!data || data.length === 0) return;
        
        const latest = data[data.length - 1];
        const previous = data.length > 1 ? data[data.length - 2] : null;
        
        // Update current price
        const currentPriceEl = document.getElementById(`${symbol}-current-price`);
        if (currentPriceEl) {
            const price = latest.close_price || latest.close;
            currentPriceEl.textContent = `$${price.toFixed(2)}`;
        }
        
        // Update price change
        const priceChangeEl = document.getElementById(`${symbol}-price-change`);
        if (priceChangeEl && previous) {
            const currentPrice = latest.close_price || latest.close;
            const previousPrice = previous.close_price || previous.close;
            const change = ((currentPrice - previousPrice) / previousPrice) * 100;
            const changeText = change > 0 ? `+${change.toFixed(2)}%` : `${change.toFixed(2)}%`;
            priceChangeEl.textContent = changeText;
            priceChangeEl.className = `badge ${change > 0 ? 'bg-success' : change < 0 ? 'bg-danger' : 'bg-secondary'}`;
        }
        
        // Update OHLC data
        const openEl = document.getElementById(`${symbol}-open`);
        const highEl = document.getElementById(`${symbol}-high`);
        const lowEl = document.getElementById(`${symbol}-low`);
        const closeEl = document.getElementById(`${symbol}-close`);
        const volumeEl = document.getElementById(`${symbol}-volume`);
        
        if (openEl) openEl.textContent = `$${(latest.open_price || latest.open).toFixed(2)}`;
        if (highEl) highEl.textContent = `$${(latest.high_price || latest.high).toFixed(2)}`;
        if (lowEl) lowEl.textContent = `$${(latest.low_price || latest.low).toFixed(2)}`;
        if (closeEl) closeEl.textContent = `$${(latest.close_price || latest.close).toFixed(2)}`;
        if (volumeEl) volumeEl.textContent = this.formatVolume(latest.volume);
    }

    async loadCollectionStatus() {
        try {
            const response = await fetch('/api/collection/status', {
                signal: this.loadingController?.signal
            });
            const data = await response.json();
            
            if (response.ok) {
                this.updateCollectionStatus(data);
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error;
            console.error('Failed to load collection status:', error);
        }
    }

    async loadTechnicalAnalysis() {
        try {
            const container = document.getElementById('technical-indicators');
            if (!container) {
                if (this.retryCount.technicalAnalysis < 3) {
                    this.retryCount.technicalAnalysis++;
                    console.warn(`Technical indicators container not found, retrying in 1 second... (${this.retryCount.technicalAnalysis}/3)`);
                    setTimeout(() => this.loadTechnicalAnalysis(), 1000);
                    return;
                } else {
                    console.error('Technical indicators container not found after 3 retries, giving up');
                    return;
                }
            }
            
            // Reset retry count on success
            this.retryCount.technicalAnalysis = 0;
            container.innerHTML = '';
            
            if (this.symbols.length === 0) {
                container.innerHTML = '<div class="text-center text-tv-muted">No symbols configured for technical analysis</div>';
                console.log('Technical analysis: No symbols available');
                return;
            }
            
            console.log('Loading technical analysis for symbols:', this.symbols);
            
            for (const symbol of this.symbols) {
                console.log(`Fetching technical indicators for ${symbol}`);
                const response = await fetch(`/api/indicators/${symbol}`, {
                    signal: this.loadingController?.signal
                });
                
                console.log(`Technical indicators response for ${symbol}:`, response.status);
                
                if (response.ok) {
                    const data = await response.json();
                    console.log(`Technical indicators data for ${symbol}:`, data);
                    this.displayTechnicalIndicators(container, symbol, data);
                } else {
                    console.warn(`Failed to load technical indicators for ${symbol}: ${response.status}`);
                }
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error;
            console.error('Failed to load technical analysis:', error);
        }
    }

    async loadTradingSetups() {
        try {
            const container = document.getElementById('trading-setups');
            if (!container) {
                if (this.retryCount.tradingSetups < 3) {
                    this.retryCount.tradingSetups++;
                    console.warn(`Trading setups container not found, retrying in 1 second... (${this.retryCount.tradingSetups}/3)`);
                    setTimeout(() => this.loadTradingSetups(), 1000);
                    return;
                } else {
                    console.error('Trading setups container not found after 3 retries, giving up');
                    return;
                }
            }
            
            // Reset retry count on success
            this.retryCount.tradingSetups = 0;
            container.innerHTML = '';
            
            console.log('Loading trading setups...');
            
            const response = await fetch('/api/setups/high-quality', {
                signal: this.loadingController?.signal
            });
            
            console.log('Trading setups response:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                console.log('Trading setups data:', data);
                this.displayTradingSetups(container, data);
            } else {
                console.warn(`Failed to load trading setups: ${response.status}`);
                container.innerHTML = '<div class="text-center text-tv-muted">Failed to load trading setups</div>';
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error;
            console.error('Failed to load trading setups:', error);
        }
    }

    async loadSupportResistance() {
        try {
            const container = document.getElementById('support-resistance');
            if (!container) {
                if (this.retryCount.supportResistance < 3) {
                    this.retryCount.supportResistance++;
                    console.warn(`Support/resistance container not found, retrying in 1 second... (${this.retryCount.supportResistance}/3)`);
                    setTimeout(() => this.loadSupportResistance(), 1000);
                    return;
                } else {
                    console.error('Support/resistance container not found after 3 retries, giving up');
                    return;
                }
            }
            
            // Reset retry count on success
            this.retryCount.supportResistance = 0;
            container.innerHTML = '';
            
            if (this.symbols.length === 0) {
                container.innerHTML = '<div class="text-center text-tv-muted">No symbols configured for support/resistance analysis</div>';
                console.log('Support/resistance: No symbols available');
                return;
            }
            
            console.log('Loading support/resistance for symbols:', this.symbols);
            
            for (const symbol of this.symbols) {
                console.log(`Fetching support/resistance for ${symbol}`);
                const response = await fetch(`/api/support-resistance/${symbol}/levels`, {
                    signal: this.loadingController?.signal
                });
                
                console.log(`Support/resistance response for ${symbol}:`, response.status);
                
                if (response.ok) {
                    const data = await response.json();
                    console.log(`Support/resistance data for ${symbol}:`, data);
                    this.displaySupportResistance(container, symbol, data);
                } else {
                    console.warn(`Failed to load support/resistance for ${symbol}: ${response.status}`);
                }
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error;
            console.error('Failed to load support/resistance:', error);
        }
    }

    displayTechnicalIndicators(container, symbol, data) {
        const indicators = data.indicators || {};
        const div = document.createElement('div');
        div.className = 'mb-3 p-3 border-tv rounded';
        div.innerHTML = `
            <h6 class="text-tv-primary mb-3">${symbol}</h6>
            <div class="row">
                <div class="col-md-6">
                    <div class="indicator-label">RSI (14)</div>
                    <div class="indicator-value ${this.getRSIClass(indicators.rsi_14)}">${indicators.rsi_14?.toFixed(2) || '--'}</div>
                </div>
                <div class="col-md-6">
                    <div class="indicator-label">SMA (20)</div>
                    <div class="indicator-value text-tv-primary">$${indicators.sma_20?.toFixed(2) || '--'}</div>
                </div>
            </div>
            <div class="row mt-2">
                <div class="col-md-6">
                    <div class="indicator-label">MACD</div>
                    <div class="indicator-value ${this.getMACDClass(indicators.macd_line, indicators.macd_signal)}">${indicators.macd_line?.toFixed(3) || '--'}</div>
                </div>
                <div class="col-md-6">
                    <div class="indicator-label">Volume Ratio</div>
                    <div class="indicator-value ${indicators.volume_ratio > 1.5 ? 'text-warning' : 'text-tv-secondary'}">${indicators.volume_ratio?.toFixed(2) || '--'}</div>
                </div>
            </div>
        `;
        container.appendChild(div);
    }

    getRSIClass(rsi) {
        if (!rsi) return 'text-tv-secondary';
        if (rsi > 70) return 'rsi-overbought';
        if (rsi < 30) return 'rsi-oversold';
        return 'rsi-neutral';
    }

    getMACDClass(macdLine, macdSignal) {
        if (!macdLine || !macdSignal) return 'text-tv-secondary';
        return macdLine > macdSignal ? 'macd-bullish' : 'macd-bearish';
    }

    displayTradingSetups(container, data) {
        if (!data.setups || data.setups.length === 0) {
            container.innerHTML = '<div class="text-tv-muted">No high-quality setups found</div>';
            return;
        }

        data.setups.forEach(setup => {
            const div = document.createElement('div');
            div.className = 'mb-3 p-3 setup-card rounded';
            div.innerHTML = `
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <h6 class="text-tv-primary">${setup.symbol}</h6>
                        <p class="mb-1"><strong>${setup.setup_type}</strong> - ${setup.direction}</p>
                        <small class="text-tv-secondary">Quality: ${setup.quality_score?.toFixed(1)}/100</small>
                    </div>
                    <span class="badge confidence-${setup.confidence}">${setup.confidence}</span>
                </div>
                <div class="mt-2">
                    <small class="text-tv-secondary">
                        Entry: <span class="text-tv-primary">$${setup.entry_price?.toFixed(2)}</span> | 
                        Target: <span class="text-success">$${setup.target1?.toFixed(2)}</span> | 
                        Stop: <span class="text-danger">$${setup.stop_loss?.toFixed(2)}</span>
                    </small>
                </div>
            `;
            container.appendChild(div);
        });
    }

    displaySupportResistance(container, symbol, data) {
        if (!data.levels || data.levels.length === 0) return;

        const div = document.createElement('div');
        div.className = 'mb-3';
        div.innerHTML = `
            <h6 class="text-tv-primary">${symbol}</h6>
            <div class="row">
                ${data.levels.slice(0, 4).map(level => `
                    <div class="col-md-3">
                        <div class="text-center p-2 border-tv rounded bg-tv-surface-light">
                            <div class="indicator-value level-${level.level_type}">
                                $${level.level.toFixed(2)}
                            </div>
                            <div class="indicator-label">${level.level_type}</div>
                            <div class="strength-indicator mt-1">
                                <div class="strength-fill" style="width: ${(level.strength || 0) * 10}%"></div>
                            </div>
                            <small class="text-tv-muted">Strength: ${level.strength?.toFixed(1)}</small>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
        container.appendChild(div);
    }

    updateCollectionStatus(status) {
        const successfulRunsEl = document.getElementById('successful-runs');
        const failedRunsEl = document.getElementById('failed-runs');
        const collectedTodayEl = document.getElementById('collected-today');
        const nextRunEl = document.getElementById('next-run');
        const runningEl = document.getElementById('collection-running');
        const statusEl = document.getElementById('collection-status');
        
        if (successfulRunsEl) {
            successfulRunsEl.textContent = status.successful_runs || 0;
        } else {
            console.warn('successful-runs element not found');
        }
        
        if (failedRunsEl) {
            failedRunsEl.textContent = status.failed_runs || 0;
        } else {
            console.warn('failed-runs element not found');
        }
        
        if (collectedTodayEl) {
            collectedTodayEl.textContent = status.collected_today || 0;
        } else {
            console.warn('collected-today element not found');
        }
        
        if (nextRunEl && status.next_run) {
            const nextRun = new Date(status.next_run);
            nextRunEl.textContent = nextRun.toLocaleTimeString();
        } else if (!nextRunEl) {
            console.warn('next-run element not found');
        }
        
        if (runningEl) {
            runningEl.textContent = status.is_running ? 'Running' : 'Idle';
        } else {
            console.warn('collection-running element not found');
        }
        
        if (statusEl) {
            if (status.last_error) {
                statusEl.className = 'alert alert-warning';
                statusEl.innerHTML = `Status: Error - ${status.last_error}`;
            } else {
                statusEl.className = 'alert alert-success';
                statusEl.innerHTML = 'Status: Running normally';
            }
        } else {
            console.warn('collection-status element not found');
        }
    }

    async forceCollection() {
        try {
            const response = await fetch('/api/collection/force', {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showSuccess('Collection triggered successfully');
                setTimeout(() => this.loadCollectionStatus(), 2000);
            } else {
                const error = await response.json();
                throw new Error(error.message || 'Failed to trigger collection');
            }
        } catch (error) {
            console.error('Failed to force collection:', error);
            this.showError('Failed to trigger collection');
        }
    }

    toggleSymbolManagement() {
        const panel = document.getElementById('symbol-management-panel');
        const isCurrentlyHidden = panel.classList.contains('d-none');
        
        panel.classList.toggle('d-none');
        
        // If we're opening the panel (was hidden, now showing), load symbols
        if (isCurrentlyHidden) {
            console.log('📋 Symbol management panel opened - loading current symbols...');
            this.loadSymbolsForManagement();
        }
    }

    async loadSymbolsForManagement() {
        try {
            console.log('🔍 Fetching symbols for management panel...');
            const response = await fetch('/api/symbols');
            console.log('📋 Management symbols API response status:', response.status);
            
            if (response.ok) {
                const data = await response.json();
                console.log('📋 Management symbols data:', data);
                this.displaySymbolsInManagementPanel(data);
            } else {
                console.error('❌ Failed to load symbols for management:', response.status);
                this.showSymbolManagementError();
            }
        } catch (error) {
            console.error('❌ Error loading symbols for management:', error);
            this.showSymbolManagementError();
        }
    }
    
    displaySymbolsInManagementPanel(data) {
        const cardBody = document.querySelector('#symbol-management-panel .card-body');
        if (!cardBody) {
            console.error('❌ Symbol management card body not found');
            return;
        }
        
        const symbols = data.symbols || [];
        console.log(`📊 Displaying ${symbols.length} symbols in management panel`);
        
        let html = `
            <!-- Add Symbol Form -->
            <div class="mb-4">
                <div class="input-group">
                    <input type="text" class="form-control" id="new-symbol-input" placeholder="Enter symbol (e.g., AAPL)" maxlength="10">
                    <button class="btn btn-primary" type="button" id="add-symbol-btn">
                        <i class="bi bi-plus-circle"></i> Add Symbol
                    </button>
                </div>
                <small class="text-tv-muted">Enter a stock symbol to start tracking (letters only, max 10 characters)</small>
            </div>
            
            <!-- Current Symbols Info -->
            <div class="alert alert-info mb-3">
                <i class="bi bi-info-circle me-2"></i>
                Currently tracking <strong>${symbols.length}</strong> symbol${symbols.length !== 1 ? 's' : ''}
            </div>
        `;
        
        if (symbols.length > 0) {
            html += `
                <div class="row" id="symbols-list">
                    ${symbols.map(symbolObj => `
                        <div class="col-md-6 col-lg-4 mb-2" data-symbol="${symbolObj.symbol}">
                            <div class="d-flex justify-content-between align-items-center p-2 border rounded bg-tv-surface-light">
                                <div>
                                    <strong class="text-tv-primary">${symbolObj.symbol}</strong>
                                    <br>
                                    <small class="text-tv-muted">ID: ${symbolObj.id}</small>
                                </div>
                                <div class="text-end">
                                    <div class="mb-1">
                                        <small class="text-tv-secondary d-block">
                                            Added: ${new Date(symbolObj.added_at).toLocaleDateString()}
                                        </small>
                                        <span class="badge ${symbolObj.is_active ? 'bg-success' : 'bg-secondary'}">
                                            ${symbolObj.is_active ? 'Active' : 'Inactive'}
                                        </span>
                                    </div>
                                    <button class="btn btn-outline-danger btn-sm remove-symbol-btn" data-symbol="${symbolObj.symbol}">
                                        <i class="bi bi-trash"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    `).join('')}
                </div>
            `;
        } else {
            html += `
                <div class="alert alert-warning">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    No symbols are currently being tracked. Add your first symbol above!
                </div>
            `;
        }
        
        cardBody.innerHTML = html;
        
        // Add event listeners for the new buttons
        this.setupSymbolManagementEventListeners();
        
        console.log('✅ Symbol management panel updated with add/remove functionality');
    }
    
    setupSymbolManagementEventListeners() {
        // Add symbol button
        const addBtn = document.getElementById('add-symbol-btn');
        const newSymbolInput = document.getElementById('new-symbol-input');
        
        if (addBtn && newSymbolInput) {
            addBtn.addEventListener('click', () => this.addSymbol());
            
            // Allow Enter key to add symbol
            newSymbolInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.addSymbol();
                }
            });
            
            // Auto-uppercase and validate input
            newSymbolInput.addEventListener('input', (e) => {
                e.target.value = e.target.value.toUpperCase().replace(/[^A-Z]/g, '');
            });
        }
        
        // Remove symbol buttons
        const removeButtons = document.querySelectorAll('.remove-symbol-btn');
        removeButtons.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const symbol = e.target.closest('button').dataset.symbol;
                this.removeSymbol(symbol);
            });
        });
    }
    
    async addSymbol() {
        const input = document.getElementById('new-symbol-input');
        const symbol = input.value.trim().toUpperCase();
        
        if (!symbol) {
            this.showError('Please enter a symbol');
            return;
        }
        
        if (symbol.length > 10) {
            this.showError('Symbol must be 10 characters or less');
            return;
        }
        
        // Check if symbol already exists
        if (this.symbols.includes(symbol)) {
            this.showError(`${symbol} is already being tracked`);
            return;
        }
        
        console.log(`➕ Adding symbol: ${symbol}`);
        
        try {
            const response = await fetch('/api/symbols', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ symbol: symbol })
            });
            
            console.log(`Add symbol response for ${symbol}:`, response.status);
            
            if (response.ok) {
                this.showSuccess(`✅ ${symbol} added successfully!`);
                input.value = ''; // Clear input
                
                // Reload the symbols in the management panel
                await this.loadSymbolsForManagement();
                
                // Reload the main dashboard symbols and recreate charts
                await this.loadSymbols();
                this.createChartsContainer();
                this.initializeCharts();
                setTimeout(() => this.loadAllChartData(), 100);
                
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || `Failed to add ${symbol}`;
                this.showError(errorMessage);
                console.error(`Failed to add ${symbol}:`, errorData);
            }
        } catch (error) {
            console.error(`Error adding symbol ${symbol}:`, error);
            this.showError(`Failed to add ${symbol}: ${error.message}`);
        }
    }
    
    async removeSymbol(symbol) {
        console.log(`➖ Removing symbol: ${symbol}`);
        
        try {
            const response = await fetch(`/api/symbols/${symbol}`, {
                method: 'DELETE'
            });
            
            console.log(`Remove symbol response for ${symbol}:`, response.status);
            
            if (response.ok) {
                this.showSuccess(`✅ ${symbol} removed successfully!`);
                
                // Reload the symbols in the management panel
                await this.loadSymbolsForManagement();
                
                // Reload the main dashboard symbols and recreate charts
                await this.loadSymbols();
                this.createChartsContainer();
                this.initializeCharts();
                setTimeout(() => this.loadAllChartData(), 100);
                
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorMessage = errorData.error || `Failed to remove ${symbol}`;
                this.showError(errorMessage);
                console.error(`Failed to remove ${symbol}:`, errorData);
            }
        } catch (error) {
            console.error(`Error removing symbol ${symbol}:`, error);
            this.showError(`Failed to remove ${symbol}: ${error.message}`);
        }
    }

    showSymbolManagementError() {
        const cardBody = document.querySelector('#symbol-management-panel .card-body');
        if (cardBody) {
            cardBody.innerHTML = `
                <div class="alert alert-danger">
                    <i class="bi bi-exclamation-triangle me-2"></i>
                    Failed to load symbols. Please try again.
                </div>
            `;
        }
    }

    refreshCharts() {
        this.loadAllChartData();
    }

    resizeCharts() {
        this.symbols.forEach(symbol => {
            this.createTradingViewChart(symbol);
        });
        setTimeout(() => this.loadAllChartData(), 100);
    }

    startAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            this.loadDashboardData();
        }, this.updateInterval);
    }

    updateLastUpdateTime() {
        const lastUpdateEl = document.getElementById('last-update');
        if (lastUpdateEl) {
            lastUpdateEl.textContent = `Last Update: ${new Date().toLocaleTimeString()}`;
        }
    }

    formatVolume(volume) {
        if (volume >= 1000000) {
            return (volume / 1000000).toFixed(1) + 'M';
        } else if (volume >= 1000) {
            return (volume / 1000).toFixed(1) + 'K';
        }
        return volume.toLocaleString();
    }

    showLoading() {
        this.isLoading = true;
        document.getElementById('refresh-spinner').classList.remove('d-none');
        document.getElementById('refresh-icon').classList.add('d-none');
        document.getElementById('refresh-text').textContent = 'Loading...';
        document.getElementById('refresh-btn').disabled = true;
        document.getElementById('cancel-refresh-btn').classList.remove('d-none');
    }

    hideLoading() {
        this.isLoading = false;
        this.loadingController = null;
        document.getElementById('refresh-spinner').classList.add('d-none');
        document.getElementById('refresh-icon').classList.remove('d-none');
        document.getElementById('refresh-text').textContent = 'Refresh';
        document.getElementById('refresh-btn').disabled = false;
        document.getElementById('cancel-refresh-btn').classList.add('d-none');
    }

    cancelLoading() {
        if (this.loadingController) {
            this.loadingController.abort();
            this.hideLoading();
            this.showSuccess('Loading cancelled');
        }
    }

    showSuccess(message) {
        this.showToast(message, 'success');
    }

    showError(message) {
        this.showToast(message, 'error');
    }

    showToast(message, type) {
        const container = document.getElementById('toast-container');
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <div class="toast-header">
                <strong class="me-auto">${type === 'success' ? 'Success' : 'Error'}</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast"></button>
            </div>
            <div class="toast-body">${message}</div>
        `;
        
        container.appendChild(toast);
        const bsToast = new bootstrap.Toast(toast);
        bsToast.show();
        
        toast.addEventListener('hidden.bs.toast', () => {
            toast.remove();
        });
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, waiting for Bootstrap to initialize...');
    
    // Manual API test to verify connectivity
    console.log('🧪 Testing /api/symbols directly...');
    fetch('/api/symbols')
        .then(response => {
            console.log('🧪 Direct API test - Status:', response.status);
            return response.json();
        })
        .then(data => {
            console.log('🧪 Direct API test - Data:', data);
            console.log('🧪 Symbols found:', data.symbols?.length || 0);
            if (data.symbols && data.symbols.length > 0) {
                console.log('🧪 First symbol:', data.symbols[0]);
            }
        })
        .catch(error => {
            console.error('🧪 Direct API test failed:', error);
        });
    
    // Wait longer for Bootstrap tabs and other components to fully initialize
    setTimeout(() => {
        console.log('Starting dashboard initialization...');
        
        // Check if essential elements exist
        const chartsGrid = document.getElementById('charts-grid');
        const technicalIndicators = document.getElementById('technical-indicators');
        const tradingSetups = document.getElementById('trading-setups');
        const supportResistance = document.getElementById('support-resistance');
        
        console.log('Element check:', {
            chartsGrid: !!chartsGrid,
            technicalIndicators: !!technicalIndicators,
            tradingSetups: !!tradingSetups,
            supportResistance: !!supportResistance
        });
        
        window.dashboard = new MarketWatchDashboard();
    }, 500); // Increased delay to allow Bootstrap tabs to fully initialize
});
