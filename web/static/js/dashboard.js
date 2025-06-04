// Market Watch Dashboard JavaScript with D3.js Advanced Financial Charts
class MarketWatchDashboard {
    constructor() {
        this.charts = {};
        this.symbols = [];
        this.currentTimeRange = '1W';
        this.refreshInterval = null;
        this.updateInterval = 30000; // 30 seconds
        this.isLoading = false; // Track loading state
        this.loadingController = null; // AbortController for cancelling requests
        
        this.init();
    }

    async init() {
        await this.loadSymbols();
        this.setupEventListeners();
        this.createChartsContainer();
        this.initializeCharts();
        
        // Add a small delay to ensure DOM is fully rendered
        await new Promise(resolve => setTimeout(resolve, 100));
        
        this.syncTimeRangeFromHTML();
        console.log('Dashboard initialized with time range:', this.currentTimeRange);
        this.loadDashboardData();
        this.startAutoRefresh();
    }

    async loadSymbols() {
        try {
            const response = await fetch('/api/symbols');
            const data = await response.json();
            
            if (response.ok) {
                this.symbols = data.symbols.map(s => s.symbol);
                console.log('Loaded symbols:', this.symbols);
            } else {
                console.error('Failed to load symbols:', data.message);
                // Fallback to default symbols
                this.symbols = ['PLTR', 'TSLA', 'BBAI', 'MSFT', 'NPWR'];
            }
        } catch (error) {
            console.error('Error loading symbols:', error);
            // Fallback to default symbols
            this.symbols = ['PLTR', 'TSLA', 'BBAI', 'MSFT', 'NPWR'];
        }
    }

    setupEventListeners() {
        // Time range selector
        document.querySelectorAll('input[name="timeRange"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.currentTimeRange = e.target.value;
                this.clearHistoricalNotifications();
                this.refreshCharts();
            });
        });

        // Refresh button
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadDashboardData();
        });

        // Cancel refresh button
        document.getElementById('cancel-refresh-btn').addEventListener('click', () => {
            this.cancelLoading();
        });

        // Force collection button
        document.getElementById('force-collection-btn').addEventListener('click', () => {
            this.forceCollection();
        });

        // Manage symbols button
        document.getElementById('manage-symbols-btn').addEventListener('click', () => {
            this.toggleSymbolManagement();
        });

        // Add symbol button
        document.getElementById('add-symbol-btn').addEventListener('click', () => {
            this.addNewSymbol();
        });

        // Enter key in symbol input
        document.getElementById('new-symbol-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.addNewSymbol();
            }
        });

        // Enter key in symbol name input
        document.getElementById('new-symbol-name').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.addNewSymbol();
            }
        });

        // Window resize handler
        window.addEventListener('resize', () => {
            this.resizeCharts();
        });
    }

    syncTimeRangeFromHTML() {
        // Debug: log all radio buttons
        const allRadios = document.querySelectorAll('input[name="timeRange"]');
        console.log('All time range radio buttons:', Array.from(allRadios).map(r => ({ id: r.id, value: r.value, checked: r.checked })));
        
        // Read the checked radio button value to sync with HTML state
        const checkedRadio = document.querySelector('input[name="timeRange"]:checked');
        if (checkedRadio) {
            this.currentTimeRange = checkedRadio.value;
            console.log('Synced time range from HTML:', this.currentTimeRange);
        } else {
            console.warn('No checked radio button found, keeping default:', this.currentTimeRange);
        }
    }

    createChartsContainer() {
        // Find the charts grid container
        const chartsGrid = document.getElementById('charts-grid');
        
        if (!chartsGrid) {
            console.error('Charts grid container not found');
            return;
        }
        
        // Clear existing chart containers
        chartsGrid.innerHTML = '';
        
        // Create chart containers for each symbol
        this.symbols.forEach(symbol => {
            const chartContainer = document.createElement('div');
            chartContainer.className = 'col-lg-6 col-xl-4 mb-4';
            chartContainer.innerHTML = `
                <div class="card h-100">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h6 class="mb-0">${symbol} Financial Chart</h6>
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
                                Current: <span id="${symbol}-last-price">--</span>
                                | Volume: <span id="${symbol}-last-volume">--</span>
                            </small>
                        </div>
                    </div>
                </div>
            `;
            chartsGrid.appendChild(chartContainer);
        });
    }

    initializeCharts() {
        this.symbols.forEach(symbol => {
            this.createD3Chart(symbol);
        });
    }

    createD3Chart(symbol) {
        const svgElement = document.getElementById(`chart-${symbol}`);
        if (!svgElement) return;

        // Clear any existing content
        d3.select(svgElement).selectAll("*").remove();

        const container = svgElement.parentElement;
        const margin = { top: 20, right: 60, bottom: 80, left: 60 };
        const width = container.clientWidth - margin.left - margin.right;
        const totalHeight = 350;
        
        // Split height: 70% for price chart, 30% for volume
        const priceHeight = Math.floor(totalHeight * 0.7) - margin.top - 40;
        const volumeHeight = Math.floor(totalHeight * 0.3) - 40 - margin.bottom;

        const svg = d3.select(svgElement)
            .attr("width", width + margin.left + margin.right)
            .attr("height", totalHeight);

        // Price chart group
        const priceGroup = svg.append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`);

        // Volume chart group
        const volumeGroup = svg.append("g")
            .attr("transform", `translate(${margin.left},${margin.top + priceHeight + 40})`);

        // Create scales
        const xScale = d3.scaleTime().range([0, width]);
        const priceScale = d3.scaleLinear().range([priceHeight, 0]);
        const volumeScale = d3.scaleLinear().range([volumeHeight, 0]);

        // Create axes
        const xAxis = d3.axisBottom(xScale)
            .tickFormat(d => {
                if (this.currentTimeRange === '1D') {
                    return d3.timeFormat("%H:%M")(d);
                } else {
                    return d3.timeFormat("%m/%d")(d);
                }
            });

        const priceAxis = d3.axisLeft(priceScale)
            .tickFormat(d => `$${d.toFixed(2)}`);

        const volumeAxis = d3.axisLeft(volumeScale)
            .tickFormat(d => this.formatVolumeShort(d));

        // Add grid lines for price chart
        const priceGrid = priceGroup.append("g")
            .attr("class", "grid price-grid")
            .style("opacity", 0.3);

        const volumeGrid = volumeGroup.append("g")
            .attr("class", "grid volume-grid")
            .style("opacity", 0.3);

        // Add axes containers
        const priceAxisContainer = priceGroup.append("g")
            .attr("class", "axis price-axis");

        const volumeAxisContainer = volumeGroup.append("g")
            .attr("class", "axis volume-axis");

        const xAxisContainer = volumeGroup.append("g")
            .attr("class", "axis x-axis")
            .attr("transform", `translate(0,${volumeHeight})`);

        // Add price right axis
        const priceRightAxisContainer = priceGroup.append("g")
            .attr("class", "axis price-right-axis")
            .attr("transform", `translate(${width},0)`);

        // Chart content containers
        const candlestickContainer = priceGroup.append("g").attr("class", "candlesticks");
        const movingAverageContainer = priceGroup.append("g").attr("class", "moving-averages");
        const volumeBarsContainer = volumeGroup.append("g").attr("class", "volume-bars");

        // Add chart titles
        priceGroup.append("text")
            .attr("class", "chart-title")
            .attr("x", width / 2)
            .attr("y", -5)
            .attr("text-anchor", "middle")
            .style("font-size", "12px")
            .style("font-weight", "bold")
            .style("fill", "#333")
            .text(`${symbol} Price & Moving Averages`);

        volumeGroup.append("text")
            .attr("class", "chart-title")
            .attr("x", width / 2)
            .attr("y", -5)
            .attr("text-anchor", "middle")
            .style("font-size", "12px")
            .style("font-weight", "bold")
            .style("fill", "#333")
            .text("Volume");

        // Add tooltip
        const tooltip = d3.select(container)
            .append("div")
            .attr("class", "financial-tooltip")
            .style("opacity", 0)
            .style("position", "absolute")
            .style("background", "rgba(0, 0, 0, 0.9)")
            .style("color", "white")
            .style("border-radius", "6px")
            .style("padding", "10px")
            .style("font-size", "12px")
            .style("pointer-events", "none")
            .style("z-index", "1000")
            .style("box-shadow", "0 4px 12px rgba(0, 0, 0, 0.3)");

        // Store chart components
        this.charts[symbol] = {
            svg,
            priceGroup,
            volumeGroup,
            xScale,
            priceScale,
            volumeScale,
            xAxis,
            priceAxis,
            volumeAxis,
            priceGrid,
            volumeGrid,
            priceAxisContainer,
            volumeAxisContainer,
            xAxisContainer,
            priceRightAxisContainer,
            candlestickContainer,
            movingAverageContainer,
            volumeBarsContainer,
            tooltip,
            width,
            priceHeight,
            volumeHeight,
            margin
        };
    }

    calculateMovingAverage(data, period) {
        const result = [];
        for (let i = 0; i < data.length; i++) {
            if (i < period - 1) {
                result.push(null);
            } else {
                const sum = data.slice(i - period + 1, i + 1)
                    .reduce((acc, d) => acc + d.close, 0);
                result.push({
                    time: data[i].time,
                    value: sum / period
                });
            }
        }
        return result.filter(d => d !== null);
    }

    async loadDashboardData() {
        // Don't start new loading if already loading
        if (this.isLoading) {
            console.log('Already loading, skipping...');
            return;
        }

        console.log('Loading dashboard data...');
        this.showLoading();
        
        // Clear any previous historical data notifications
        this.clearHistoricalNotifications();
        
        // Create AbortController for cancelling requests
        this.loadingController = new AbortController();
        
        try {
            // Load dashboard summary
            console.log('Loading dashboard summary...');
            await this.loadDashboardSummary();
            
            // Check if cancelled
            if (this.loadingController.signal.aborted) return;
            
            // Load chart data for all symbols
            console.log('Loading chart data...');
            await this.loadAllChartData();
            
            // Check if cancelled
            if (this.loadingController.signal.aborted) return;
            
            // Load collection status
            console.log('Loading collection status...');
            await this.loadCollectionStatus();
            
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

    async loadDashboardSummary() {
        try {
            const response = await fetch('/api/dashboard/summary', {
                signal: this.loadingController?.signal
            });
            const data = await response.json();
            
            if (response.ok) {
                this.updateMarketStatus(data.market_hours);
            } else {
                throw new Error(data.message || 'Failed to load summary');
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error; // Re-throw abort errors
            console.error('Failed to load dashboard summary:', error);
        }
    }

    async loadAllChartData() {
        const promises = this.symbols.map(symbol => this.loadChartData(symbol));
        await Promise.all(promises);
    }

    async loadChartData(symbol) {
        try {
            let url = `/api/price/${symbol}/chart?range=${this.currentTimeRange}`;
            let isShowingHistoricalDay = false;
            let historicalDate = null;
            
            // For daily view, check if we need to show last trading day
            if (this.currentTimeRange === '1D') {
                const today = new Date();
                const dayOfWeek = today.getDay();
                
                // If it's weekend (Saturday=6, Sunday=0), show last Friday
                if (dayOfWeek === 0 || dayOfWeek === 6) {
                    const lastTradingDay = this.getLastTradingDay(today);
                    const fromDate = new Date(lastTradingDay);
                    const toDate = new Date(lastTradingDay);
                    toDate.setDate(toDate.getDate() + 1); // Next day to include full trading day
                    
                    url = `/api/price/${symbol}?range=1D&from=${fromDate.toISOString().split('T')[0]}&to=${toDate.toISOString().split('T')[0]}`;
                    isShowingHistoricalDay = true;
                    historicalDate = lastTradingDay;
                    console.log(`Weekend detected, loading last trading day for ${symbol}: ${url}`);
                }
            }
            
            console.log(`Loading chart data for ${symbol} with URL: ${url}`);
            const response = await fetch(url, {
                signal: this.loadingController?.signal
            });
            const data = await response.json();
            
            console.log(`Chart data response for ${symbol}:`, data);
            
            if (response.ok) {
                let chartData = [];
                
                // Handle price data response format
                if (data.data && Array.isArray(data.data)) {
                    // Price API format - use full OHLC data for candlesticks
                    chartData = data.data.map(point => ({
                        time: new Date(point.time * 1000), // Convert from seconds to milliseconds
                        open: point.open,
                        high: point.high,
                        low: point.low,
                        close: point.close,
                        volume: point.volume
                    }));
                }
                
                // Filter out weekend data and non-trading hours
                chartData = chartData.filter(point => {
                    const day = point.time.getDay();
                    // Skip weekends (0 = Sunday, 6 = Saturday)
                    if (day === 0 || day === 6) {
                        return false;
                    }
                    
                    // Filter out non-trading hours for all time ranges
                    const hour = point.time.getHours();
                    const minute = point.time.getMinutes();
                    
                    // Market hours: 9:30 AM to 4:00 PM ET
                    // Before 9:30 AM
                    if (hour < 9 || (hour === 9 && minute < 30)) {
                        return false;
                    }
                    // After 4:00 PM
                    if (hour > 16) {
                        return false;
                    }
                    
                    return true;
                });
                
                console.log(`Processed chart data for ${symbol}:`, chartData.length, 'points (weekends filtered)');
                
                // Update chart with D3 candlesticks
                this.updateD3Chart(symbol, chartData);
                
                // Show notification if displaying historical data
                if (isShowingHistoricalDay && historicalDate) {
                    this.showHistoricalDataNotification(symbol, historicalDate);
                }
                
                // Update symbol info if we have data
                if (chartData.length > 0) {
                    this.updateSymbolInfo(symbol, chartData);
                }
            } else {
                console.warn(`No chart data for ${symbol}:`, data);
                // Clear the chart if no data
                this.updateD3Chart(symbol, []);
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error; // Re-throw abort errors
            console.error(`Failed to load chart data for ${symbol}:`, error);
        }
    }

    updateD3Chart(symbol, data) {
        const chart = this.charts[symbol];
        if (!chart || !data) return;

        const { 
            xScale, priceScale, volumeScale, 
            xAxis, priceAxis, volumeAxis,
            priceGrid, volumeGrid,
            priceAxisContainer, volumeAxisContainer, xAxisContainer, priceRightAxisContainer,
            candlestickContainer, movingAverageContainer, volumeBarsContainer,
            tooltip, width, priceHeight, volumeHeight 
        } = chart;

        if (data.length === 0) {
            // Clear charts
            candlestickContainer.selectAll("*").remove();
            movingAverageContainer.selectAll("*").remove();
            volumeBarsContainer.selectAll("*").remove();
            return;
        }

        // Update scales domains
        const xExtent = d3.extent(data, d => d.time);
        const priceExtent = d3.extent(data, d => [d.low, d.high]).flat();
        const volumeExtent = d3.extent(data, d => d.volume);
        
        const pricePadding = (priceExtent[1] - priceExtent[0]) * 0.1;
        const volumePadding = volumeExtent[1] * 0.1;

        xScale.domain(xExtent);
        priceScale.domain([priceExtent[0] - pricePadding, priceExtent[1] + pricePadding]);
        volumeScale.domain([0, volumeExtent[1] + volumePadding]);

        // Calculate moving averages
        const ma20 = this.calculateMovingAverage(data, 20);
        const ma50 = this.calculateMovingAverage(data, 50);

        // Update grids
        priceGrid.call(d3.axisLeft(priceScale)
            .tickSize(-width)
            .tickFormat("")
        );

        volumeGrid.call(d3.axisLeft(volumeScale)
            .tickSize(-width)
            .tickFormat("")
        );

        // Update axes
        priceAxisContainer.call(priceAxis);
        volumeAxisContainer.call(volumeAxis);
        xAxisContainer.call(xAxis);
        priceRightAxisContainer.call(d3.axisRight(priceScale).tickFormat(d => `$${d.toFixed(2)}`));

        // Calculate candlestick width
        const candleWidth = Math.max(2, Math.min(8, width / data.length * 0.7));

        // Clear existing content
        candlestickContainer.selectAll("*").remove();
        movingAverageContainer.selectAll("*").remove();
        volumeBarsContainer.selectAll("*").remove();

        // Draw volume bars first (background)
        const volumeBars = volumeBarsContainer.selectAll(".volume-bar")
            .data(data)
            .enter()
            .append("rect")
            .attr("class", "volume-bar")
            .attr("x", d => xScale(d.time) - candleWidth / 2)
            .attr("y", d => volumeScale(d.volume))
            .attr("width", candleWidth)
            .attr("height", d => volumeHeight - volumeScale(d.volume))
            .style("fill", d => d.close >= d.open ? "rgba(44, 160, 44, 0.6)" : "rgba(214, 39, 40, 0.6)")
            .style("stroke", "none");

        // Draw moving averages
        if (ma20.length > 0) {
            const ma20Line = d3.line()
                .x(d => xScale(d.time))
                .y(d => priceScale(d.value))
                .curve(d3.curveMonotoneX);

            movingAverageContainer.append("path")
                .datum(ma20)
                .attr("class", "ma-20")
                .attr("d", ma20Line)
                .style("fill", "none")
                .style("stroke", "#ff7f0e")
                .style("stroke-width", 2)
                .style("opacity", 0.8);
        }

        if (ma50.length > 0) {
            const ma50Line = d3.line()
                .x(d => xScale(d.time))
                .y(d => priceScale(d.value))
                .curve(d3.curveMonotoneX);

            movingAverageContainer.append("path")
                .datum(ma50)
                .attr("class", "ma-50")
                .attr("d", ma50Line)
                .style("fill", "none")
                .style("stroke", "#1f77b4")
                .style("stroke-width", 2)
                .style("opacity", 0.8);
        }

        // Draw candlesticks
        const candlesticks = candlestickContainer.selectAll(".candlestick")
            .data(data)
            .enter()
            .append("g")
            .attr("class", "candlestick")
            .attr("transform", d => `translate(${xScale(d.time)}, 0)`);

        // Add high-low lines (wicks)
        candlesticks.append("line")
            .attr("class", "wick")
            .attr("x1", 0)
            .attr("x2", 0)
            .attr("y1", d => priceScale(d.high))
            .attr("y2", d => priceScale(d.low))
            .style("stroke", "#666")
            .style("stroke-width", 1);

        // Add open-close rectangles (bodies)
        candlesticks.append("rect")
            .attr("class", "body")
            .attr("x", -candleWidth / 2)
            .attr("y", d => priceScale(Math.max(d.open, d.close)))
            .attr("width", candleWidth)
            .attr("height", d => Math.abs(priceScale(d.open) - priceScale(d.close)) || 1)
            .style("fill", d => d.close >= d.open ? "#2ca02c" : "#d62728")
            .style("stroke", d => d.close >= d.open ? "#2ca02c" : "#d62728")
            .style("stroke-width", 1);

        // Add tooltip interactions for both charts
        const createTooltipHandler = (isVolumeChart = false) => {
            return {
                mouseover: (event, d) => {
                    tooltip.style("opacity", 1);
                    
                    const formatTime = this.currentTimeRange === '1D' 
                        ? d3.timeFormat("%H:%M")
                        : d3.timeFormat("%m/%d %H:%M");
                    
                    const change = d.close - d.open;
                    const changePercent = ((change / d.open) * 100).toFixed(2);
                    const changeColor = change >= 0 ? "#2ca02c" : "#d62728";
                    
                    // Find MA values for this time
                    const ma20Value = ma20.find(ma => ma.time.getTime() === d.time.getTime());
                    const ma50Value = ma50.find(ma => ma.time.getTime() === d.time.getTime());
                    
                    tooltip.html(`
                        <div style="font-weight: bold; margin-bottom: 4px;">${symbol} - ${formatTime(d.time)}</div>
                        <div>Open: $${d.open.toFixed(2)}</div>
                        <div>High: $${d.high.toFixed(2)}</div>
                        <div>Low: $${d.low.toFixed(2)}</div>
                        <div>Close: $${d.close.toFixed(2)}</div>
                        <div style="color: ${changeColor};">
                            Change: ${change >= 0 ? '+' : ''}$${change.toFixed(2)} (${changePercent}%)
                        </div>
                        <div>Volume: ${this.formatVolume(d.volume)}</div>
                        ${ma20Value ? `<div style="color: #ff7f0e;">MA20: $${ma20Value.value.toFixed(2)}</div>` : ''}
                        ${ma50Value ? `<div style="color: #1f77b4;">MA50: $${ma50Value.value.toFixed(2)}</div>` : ''}
                    `)
                    .style("left", (event.pageX + 10) + "px")
                    .style("top", (event.pageY - 10) + "px");
                },
                mouseout: () => {
                    tooltip.style("opacity", 0);
                },
                mousemove: (event) => {
                    tooltip
                        .style("left", (event.pageX + 10) + "px")
                        .style("top", (event.pageY - 10) + "px");
                }
            };
        };

        const tooltipHandler = createTooltipHandler();

        // Add interactions to candlesticks
        candlesticks
            .on("mouseover", tooltipHandler.mouseover)
            .on("mouseout", tooltipHandler.mouseout)
            .on("mousemove", tooltipHandler.mousemove);

        // Add interactions to volume bars
        volumeBars
            .on("mouseover", tooltipHandler.mouseover)
            .on("mouseout", tooltipHandler.mouseout)
            .on("mousemove", tooltipHandler.mousemove);

        // Add legend for moving averages
        const legend = movingAverageContainer.append("g")
            .attr("class", "legend")
            .attr("transform", `translate(${width - 120}, 20)`);

        if (ma20.length > 0) {
            const ma20Legend = legend.append("g").attr("transform", "translate(0, 0)");
            ma20Legend.append("line")
                .attr("x1", 0).attr("x2", 15)
                .attr("y1", 0).attr("y2", 0)
                .style("stroke", "#ff7f0e").style("stroke-width", 2);
            ma20Legend.append("text")
                .attr("x", 20).attr("y", 0).attr("dy", "0.35em")
                .style("font-size", "10px").style("fill", "#333")
                .text("MA20");
        }

        if (ma50.length > 0) {
            const ma50Legend = legend.append("g").attr("transform", "translate(0, 15)");
            ma50Legend.append("line")
                .attr("x1", 0).attr("x2", 15)
                .attr("y1", 0).attr("y2", 0)
                .style("stroke", "#1f77b4").style("stroke-width", 2);
            ma50Legend.append("text")
                .attr("x", 20).attr("y", 0).attr("dy", "0.35em")
                .style("font-size", "10px").style("fill", "#333")
                .text("MA50");
        }
    }

    formatVolumeShort(volume) {
        if (volume >= 1000000) {
            return (volume / 1000000).toFixed(0) + 'M';
        } else if (volume >= 1000) {
            return (volume / 1000).toFixed(0) + 'K';
        }
        return volume.toString();
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
            if (error.name === 'AbortError') throw error; // Re-throw abort errors
            console.error('Failed to load collection status:', error);
        }
    }

    updateMarketStatus(isOpen) {
        const statusElement = document.getElementById('market-status');
        if (statusElement) {
            statusElement.textContent = `Market Status: ${isOpen ? 'Open' : 'Closed'}`;
            statusElement.className = `badge ${isOpen ? 'bg-success' : 'bg-danger'}`;
        }
    }

    updateSymbolInfo(symbol, chartData) {
        if (!chartData || chartData.length === 0) return;
        
        const latest = chartData[chartData.length - 1];
        const previous = chartData.length > 1 ? chartData[chartData.length - 2] : null;
        
        // Update current price
        const currentPriceEl = document.getElementById(`${symbol}-current-price`);
        if (currentPriceEl) {
            currentPriceEl.textContent = this.formatPrice(latest.close);
        }
        
        // Update price change
        const priceChangeEl = document.getElementById(`${symbol}-price-change`);
        if (priceChangeEl && previous) {
            const change = ((latest.close - previous.close) / previous.close) * 100;
            const changeText = change > 0 ? `+${change.toFixed(2)}%` : `${change.toFixed(2)}%`;
            priceChangeEl.textContent = changeText;
            priceChangeEl.className = `badge ${change > 0 ? 'bg-success' : change < 0 ? 'bg-danger' : 'bg-secondary'}`;
        }
        
        // Update last price
        const lastPriceEl = document.getElementById(`${symbol}-last-price`);
        if (lastPriceEl) {
            lastPriceEl.textContent = this.formatPrice(latest.close);
        }

        // Update last volume
        const lastVolumeEl = document.getElementById(`${symbol}-last-volume`);
        if (lastVolumeEl) {
            lastVolumeEl.textContent = this.formatVolume(latest.volume);
        }
    }

    updateCollectionStatus(status) {
        document.getElementById('successful-runs').textContent = status.successful_runs || 0;
        document.getElementById('failed-runs').textContent = status.failed_runs || 0;
        document.getElementById('collected-today').textContent = status.collected_today || 0;
        
        const nextRunEl = document.getElementById('next-run');
        if (status.next_run) {
            const nextRun = new Date(status.next_run);
            nextRunEl.textContent = nextRun.toLocaleTimeString();
        }
        
        const runningEl = document.getElementById('collection-running');
        runningEl.textContent = status.is_running ? 'Running' : 'Idle';
        
        const statusEl = document.getElementById('collection-status');
        if (status.last_error) {
            statusEl.className = 'alert alert-warning';
            statusEl.innerHTML = `Status: Error - ${status.last_error}`;
        } else {
            statusEl.className = 'alert alert-success';
            statusEl.innerHTML = 'Status: Running normally';
        }
    }

    async forceCollection() {
        try {
            const response = await fetch('/api/collection/force', {
                method: 'POST'
            });
            
            if (response.ok) {
                this.showSuccess('Collection triggered successfully');
                // Refresh status after a short delay
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

    refreshCharts() {
        this.loadAllChartData();
    }

    resizeCharts() {
        // Recreate charts on resize
        this.symbols.forEach(symbol => {
            this.createD3Chart(symbol);
        });
        // Reload data to update the resized charts
        setTimeout(() => this.loadAllChartData(), 100);
    }

    startAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            this.loadDashboardData();
        }, this.updateInterval);
    }

    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    updateLastUpdateTime() {
        const lastUpdateEl = document.getElementById('last-update');
        if (lastUpdateEl) {
            lastUpdateEl.textContent = `Last Update: ${new Date().toLocaleTimeString()}`;
        }
    }

    getSymbolColor(symbol, alpha = 1) {
        const colors = {
            'PLTR': '#1f77b4',
            'TSLA': '#ff7f0e',
            'BBAI': '#2ca02c',
            'MSFT': '#d62728',
            'NPWR': '#9467bd'
        };
        const color = colors[symbol] || '#1f77b4';
        
        if (alpha === 1) {
            return color;
        } else {
            // Convert hex to rgba
            const r = parseInt(color.slice(1, 3), 16);
            const g = parseInt(color.slice(3, 5), 16);
            const b = parseInt(color.slice(5, 7), 16);
            return `rgba(${r}, ${g}, ${b}, ${alpha})`;
        }
    }

    showHistoricalDataNotification(symbol, historicalDate) {
        // Show global notification if not already shown
        if (!document.getElementById('historical-data-banner')) {
            this.showGlobalHistoricalNotification(historicalDate);
        }
        
        // Add badge to specific chart
        const chartCard = document.querySelector(`#chart-${symbol}`).closest('.card');
        if (chartCard) {
            const header = chartCard.querySelector('.card-header');
            if (header && !header.querySelector('.historical-badge')) {
                const badge = document.createElement('span');
                badge.className = 'badge bg-warning historical-badge ms-2';
                badge.textContent = `Showing ${historicalDate.toLocaleDateString([], {month: 'short', day: 'numeric'})}`;
                header.appendChild(badge);
            }
        }
    }

    showGlobalHistoricalNotification(historicalDate) {
        const dateStr = historicalDate.toLocaleDateString([], {
            weekday: 'long',
            month: 'long',
            day: 'numeric'
        });
        
        const banner = document.createElement('div');
        banner.id = 'historical-data-banner';
        banner.className = 'alert alert-info alert-dismissible fade show mb-3';
        banner.innerHTML = `
            <i class="bi bi-info-circle me-2"></i>
            <strong>Showing historical data:</strong> Displaying trading data from ${dateStr} (last trading day)
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        // Insert banner after the controls section
        const controlsRow = document.querySelector('.row.mb-4');
        if (controlsRow) {
            controlsRow.insertAdjacentElement('afterend', banner);
        }
    }

    clearHistoricalNotifications() {
        // Remove global banner
        const banner = document.getElementById('historical-data-banner');
        if (banner) {
            banner.remove();
        }
        
        // Remove chart badges
        document.querySelectorAll('.historical-badge').forEach(badge => {
            badge.remove();
        });
    }

    getLastTradingDay(date) {
        const result = new Date(date);
        const dayOfWeek = result.getDay();
        
        if (dayOfWeek === 0) {
            // Sunday - go back to Friday
            result.setDate(result.getDate() - 2);
        } else if (dayOfWeek === 6) {
            // Saturday - go back to Friday
            result.setDate(result.getDate() - 1);
        } else if (dayOfWeek === 1) {
            // Monday - go back to Friday (3 days)
            result.setDate(result.getDate() - 3);
        } else {
            // Tuesday-Friday - go back one day
            result.setDate(result.getDate() - 1);
        }
        
        return result;
    }

    formatVolume(volume) {
        if (volume >= 1000000) {
            return (volume / 1000000).toFixed(1) + 'M';
        } else if (volume >= 1000) {
            return (volume / 1000).toFixed(1) + 'K';
        }
        return volume.toLocaleString();
    }

    formatPrice(price) {
        return '$' + price.toFixed(2);
    }

    showLoading() {
        this.isLoading = true;
        
        // Show spinner, hide icon
        document.getElementById('refresh-spinner').classList.remove('d-none');
        document.getElementById('refresh-icon').classList.add('d-none');
        
        // Update text and disable button
        document.getElementById('refresh-text').textContent = 'Loading...';
        document.getElementById('refresh-btn').disabled = true;
        
        // Show cancel button
        document.getElementById('cancel-refresh-btn').classList.remove('d-none');
    }

    hideLoading() {
        this.isLoading = false;
        this.loadingController = null;
        
        // Hide spinner, show icon
        document.getElementById('refresh-spinner').classList.add('d-none');
        document.getElementById('refresh-icon').classList.remove('d-none');
        
        // Reset text and enable button
        document.getElementById('refresh-text').textContent = 'Refresh';
        document.getElementById('refresh-btn').disabled = false;
        
        // Hide cancel button
        document.getElementById('cancel-refresh-btn').classList.add('d-none');
    }

    cancelLoading() {
        if (this.loadingController) {
            console.log('Cancelling data loading...');
            this.loadingController.abort();
            this.hideLoading();
            this.showSuccess('Data loading cancelled');
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
        const toastId = 'toast-' + Date.now();
        
        const toast = document.createElement('div');
        toast.id = toastId;
        toast.className = `toast toast-${type}`;
        toast.setAttribute('role', 'alert');
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
        
        // Remove toast element after it's hidden
        toast.addEventListener('hidden.bs.toast', () => {
            toast.remove();
        });
    }

    // Symbol Management Methods
    toggleSymbolManagement() {
        const panel = document.getElementById('symbol-management-panel');
        const btn = document.getElementById('manage-symbols-btn');
        
        if (panel.classList.contains('d-none')) {
            panel.classList.remove('d-none');
            btn.innerHTML = '<i class="bi bi-x-lg"></i> Close';
            this.loadWatchedSymbolsList();
        } else {
            panel.classList.add('d-none');
            btn.innerHTML = '<i class="bi bi-gear"></i> Manage Symbols';
        }
    }

    async loadWatchedSymbolsList() {
        try {
            const response = await fetch('/api/symbols');
            const data = await response.json();
            
            const container = document.getElementById('watched-symbols-list');
            
            if (response.ok && data.symbols && data.symbols.length > 0) {
                container.innerHTML = '';
                
                data.symbols.forEach(symbol => {
                    const symbolBadge = document.createElement('div');
                    symbolBadge.className = 'badge bg-primary fs-6 p-2 d-flex align-items-center';
                    symbolBadge.innerHTML = `
                        <span class="me-2">${symbol.symbol}</span>
                        <span class="data-status" id="status-${symbol.symbol}">
                            <i class="bi bi-hourglass-split text-warning" title="Checking data..."></i>
                        </span>
                        <button class="btn btn-sm btn-outline-light border-0 p-0 ms-1"
                                onclick="window.dashboard.removeSymbol('${symbol.symbol}')"
                                title="Remove ${symbol.symbol}">
                            <i class="bi bi-x-lg"></i>
                        </button>
                    `;
                    container.appendChild(symbolBadge);
                    
                    // Check data availability for this symbol
                    this.checkSymbolData(symbol.symbol);
                });
            } else {
                container.innerHTML = '<div class="text-muted">No symbols are currently being watched.</div>';
            }
        } catch (error) {
            console.error('Failed to load watched symbols:', error);
            document.getElementById('watched-symbols-list').innerHTML =
                '<div class="text-danger">Failed to load symbols</div>';
        }
    }

    async addNewSymbol() {
        const symbolInput = document.getElementById('new-symbol-input');
        const nameInput = document.getElementById('new-symbol-name');
        const addBtn = document.getElementById('add-symbol-btn');
        const spinner = document.getElementById('add-symbol-spinner');
        const btnText = document.getElementById('add-symbol-text');
        
        const symbol = symbolInput.value.trim().toUpperCase();
        const name = nameInput.value.trim();
        
        if (!symbol) {
            this.showError('Please enter a ticker symbol');
            symbolInput.focus();
            return;
        }

        // Show loading state
        spinner.classList.remove('d-none');
        btnText.textContent = 'Adding...';
        addBtn.disabled = true;

        try {
            const response = await fetch('/api/symbols', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    symbol: symbol,
                    name: name
                })
            });

            const data = await response.json();

            if (response.ok) {
                this.showSuccess(`Symbol ${symbol} added successfully`);
                
                // Clear inputs
                symbolInput.value = '';
                nameInput.value = '';
                
                // Reload symbols list
                this.loadWatchedSymbolsList();
                
                // Refresh the dashboard with new symbols
                setTimeout(async () => {
                    await this.loadSymbols();
                    this.createChartsContainer();
                    this.initializeCharts();
                    this.loadDashboardData();
                }, 1000);
                
            } else {
                throw new Error(data.message || 'Failed to add symbol');
            }
        } catch (error) {
            console.error('Failed to add symbol:', error);
            this.showError('Failed to add symbol: ' + error.message);
        } finally {
            // Hide loading state
            spinner.classList.add('d-none');
            btnText.textContent = 'Add Symbol';
            addBtn.disabled = false;
        }
    }

    async removeSymbol(symbol) {
        try {
            const response = await fetch(`/api/symbols/${symbol}`, {
                method: 'DELETE'
            });

            const data = await response.json();

            if (response.ok) {
                this.showSuccess(`Symbol ${symbol} removed successfully`);
                
                // Reload symbols list
                this.loadWatchedSymbolsList();
                
                // Refresh the dashboard with updated symbols
                setTimeout(async () => {
                    await this.loadSymbols();
                    this.createChartsContainer();
                    this.initializeCharts();
                    this.loadDashboardData();
                }, 1000);
                
            } else {
                throw new Error(data.message || 'Failed to remove symbol');
            }
        } catch (error) {
            console.error('Failed to remove symbol:', error);
            this.showError('Failed to remove symbol: ' + error.message);
        }
    }

    async checkSymbolData(symbol) {
        try {
            const response = await fetch(`/api/symbols/${symbol}/check`);
            const data = await response.json();
            
            const statusElement = document.getElementById(`status-${symbol}`);
            if (!statusElement) return;
            
            if (response.ok) {
                if (data.has_data && data.data_points > 0) {
                    // Symbol has data
                    statusElement.innerHTML = `
                        <i class="bi bi-check-circle-fill text-success"
                           title="${data.data_points} data points"></i>
                    `;
                } else {
                    // Symbol has no data - show collect button
                    statusElement.innerHTML = `
                        <button class="btn btn-sm btn-outline-warning border-0 p-0"
                                onclick="window.dashboard.collectSymbolData('${symbol}')"
                                title="No data - click to collect">
                            <i class="bi bi-exclamation-triangle-fill"></i>
                        </button>
                    `;
                }
            } else {
                // Error checking data
                statusElement.innerHTML = `
                    <i class="bi bi-question-circle text-secondary"
                       title="Unable to check data status"></i>
                `;
            }
        } catch (error) {
            console.error('Failed to check symbol data:', error);
            const statusElement = document.getElementById(`status-${symbol}`);
            if (statusElement) {
                statusElement.innerHTML = `
                    <i class="bi bi-question-circle text-secondary"
                       title="Error checking data"></i>
                `;
            }
        }
    }

    async collectSymbolData(symbol) {
        try {
            const statusElement = document.getElementById(`status-${symbol}`);
            if (statusElement) {
                statusElement.innerHTML = `
                    <span class="spinner-border spinner-border-sm text-info"
                          title="Collecting data..."></span>
                `;
            }

            const response = await fetch(`/api/symbols/${symbol}/collect`, {
                method: 'POST'
            });

            const data = await response.json();

            if (response.ok) {
                this.showSuccess(`Data collection started for ${symbol}`);
                
                // Check status again after a delay
                setTimeout(() => {
                    this.checkSymbolData(symbol);
                }, 3000);
                
                // Also refresh the dashboard data
                setTimeout(() => {
                    this.loadDashboardData();
                }, 5000);
                
            } else {
                throw new Error(data.message || 'Failed to trigger data collection');
            }
        } catch (error) {
            console.error('Failed to collect symbol data:', error);
            this.showError('Failed to collect data: ' + error.message);
            
            // Reset status indicator
            setTimeout(() => {
                this.checkSymbolData(symbol);
            }, 1000);
        }
    }
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Small delay to ensure all elements are rendered
    setTimeout(() => {
        window.dashboard = new MarketWatchDashboard();
    }, 100);
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (window.dashboard) {
        window.dashboard.stopAutoRefresh();
    }
});
