// Market Watch Dashboard JavaScript
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
                        <h6 class="mb-0">${symbol} Volume</h6>
                        <div class="d-flex align-items-center">
                            <span id="${symbol}-current-volume" class="badge bg-primary me-2">--</span>
                            <span id="${symbol}-volume-change" class="badge bg-secondary">--</span>
                        </div>
                    </div>
                    <div class="card-body">
                        <div class="chart-container">
                            <canvas id="chart-${symbol}" width="400" height="300"></canvas>
                        </div>
                        <div class="mt-2">
                            <small class="text-muted">
                                Last: <span id="${symbol}-last-price">--</span> |
                                Volume: <span id="${symbol}-last-volume">--</span> |
                                Ratio: <span id="${symbol}-volume-ratio">--</span>
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
            const ctx = document.getElementById(`chart-${symbol}`);
            if (ctx) {
                this.charts[symbol] = new Chart(ctx, {
                    type: 'line',
                    data: {
                        datasets: [{
                            label: `${symbol} Volume`,
                            data: [],
                            borderColor: this.getSymbolColor(symbol),
                            backgroundColor: this.getSymbolColor(symbol, 0.1),
                            fill: true,
                            tension: 0.2,
                            pointRadius: 0,
                            pointHoverRadius: 4,
                            borderWidth: 2,
                            spanGaps: false  // Don't connect points across gaps
                        }]
                    },
                    options: {
                        responsive: true,
                        maintainAspectRatio: false,
                        interaction: {
                            intersect: false,
                            mode: 'index'
                        },
                        plugins: {
                            legend: {
                                display: false
                            },
                            tooltip: {
                                callbacks: {
                                    title: function(tooltipItems) {
                                        const date = new Date(tooltipItems[0].parsed.x);
                                        const timeRange = window.dashboard?.currentTimeRange || '1W';
                                        
                                        // Format tooltip with time component
                                        if (timeRange === '1D') {
                                            // Daily: Show date and time with hours:minutes
                                            return date.toLocaleDateString([], {month: 'short', day: 'numeric'}) + ' at ' +
                                                   date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', hour12: true});
                                        } else {
                                            // Weekly+: Show month/day and time
                                            return date.toLocaleDateString([], {month: 'short', day: 'numeric'}) + ' at ' +
                                                   date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit', hour12: true});
                                        }
                                    },
                                    label: function(context) {
                                        return `Volume: ${context.parsed.y.toLocaleString()}`;
                                    }
                                }
                            }
                        },
                        scales: {
                            x: {
                                type: 'time',
                                time: {
                                    displayFormats: {
                                        minute: 'HH',
                                        hour: 'HH',
                                        day: 'MMM D',
                                        week: 'MMM D'
                                    }
                                },
                                ticks: {
                                    callback: function(value, index, values) {
                                        const date = new Date(value);
                                        const day = date.getDay();
                                        // Skip weekends (0 = Sunday, 6 = Saturday)
                                        if (day === 0 || day === 6) {
                                            return null;
                                        }
                                        
                                        // Get time range from dashboard instance
                                        const timeRange = window.dashboard?.currentTimeRange || '1W';
                                        
                                        // Skip non-trading hours for all time ranges
                                        const hour = date.getHours();
                                        const minute = date.getMinutes();
                                        
                                        // Skip non-trading hours (before 9:30 AM or after 4:00 PM)
                                        if (hour < 9 || (hour === 9 && minute < 30) || hour > 16) {
                                            return null;
                                        }
                                        
                                        // Format based on time range
                                        if (timeRange === '1D') {
                                            return date.toLocaleTimeString([], {hour: '2-digit', hour12: false});
                                        } else {
                                            return date.toLocaleDateString([], {month: 'short', day: 'numeric'});
                                        }
                                    }
                                },
                                grid: {
                                    display: false
                                }
                            },
                            y: {
                                beginAtZero: true,
                                ticks: {
                                    callback: function(value) {
                                        return value.toLocaleString();
                                    }
                                },
                                grid: {
                                    color: 'rgba(0,0,0,0.1)'
                                }
                            }
                        }
                    }
                });
            }
        });
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
            let url = `/api/volume/${symbol}/chart?range=${this.currentTimeRange}`;
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
                    
                    url = `/api/volume/${symbol}?from=${fromDate.toISOString().split('T')[0]}&to=${toDate.toISOString().split('T')[0]}&interval=5m`;
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
                
                // Handle different API response formats
                if (data.datasets && data.datasets.length > 0) {
                    // Chart API format
                    chartData = data.datasets[0].data;
                } else if (data.data && Array.isArray(data.data)) {
                    // Volume API format
                    chartData = data.data.map(point => ({
                        x: point.timestamp,
                        y: point.volume
                    }));
                }
                
                // Filter out weekend data and non-trading hours
                chartData = chartData
                    .map(point => ({
                        x: new Date(point.x),
                        y: point.y
                    }))
                    .filter(point => {
                        const day = point.x.getDay();
                        // Skip weekends (0 = Sunday, 6 = Saturday)
                        if (day === 0 || day === 6) {
                            return false;
                        }
                        
                        // Filter out non-trading hours for all time ranges
                        const hour = point.x.getHours();
                        const minute = point.x.getMinutes();
                        
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
                
                // Update chart time format based on current range
                this.updateChartTimeFormat(symbol);
                
                this.charts[symbol].data.datasets[0].data = chartData;
                this.charts[symbol].update('none');
                
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
                this.charts[symbol].data.datasets[0].data = [];
                this.charts[symbol].update('none');
            }
        } catch (error) {
            if (error.name === 'AbortError') throw error; // Re-throw abort errors
            console.error(`Failed to load chart data for ${symbol}:`, error);
        }
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
        
        // Update current volume
        const currentVolumeEl = document.getElementById(`${symbol}-current-volume`);
        if (currentVolumeEl) {
            currentVolumeEl.textContent = this.formatVolume(latest.y);
        }
        
        // Update volume change
        const volumeChangeEl = document.getElementById(`${symbol}-volume-change`);
        if (volumeChangeEl && previous) {
            const change = ((latest.y - previous.y) / previous.y) * 100;
            const changeText = change > 0 ? `+${change.toFixed(1)}%` : `${change.toFixed(1)}%`;
            volumeChangeEl.textContent = changeText;
            volumeChangeEl.className = `badge ${change > 0 ? 'bg-success' : change < 0 ? 'bg-danger' : 'bg-secondary'}`;
        }
        
        // Update last volume
        const lastVolumeEl = document.getElementById(`${symbol}-last-volume`);
        if (lastVolumeEl) {
            lastVolumeEl.textContent = this.formatVolume(latest.y);
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

    updateChartTimeFormat(symbol) {
        const chart = this.charts[symbol];
        if (!chart) return;
        
        // Update time display format based on current time range
        let displayFormats;
        let unit;
        
        if (this.currentTimeRange === '1D') {
            displayFormats = {
                minute: 'HH',
                hour: 'HH',
                day: 'HH'
            };
            unit = 'hour';
        } else {
            // 1W and 2W
            displayFormats = {
                hour: 'MMM D',
                day: 'MMM D',
                week: 'MMM D'
            };
            unit = 'day';
        }
        
        chart.options.scales.x.time.displayFormats = displayFormats;
        chart.options.scales.x.time.unit = unit;
    }

    refreshCharts() {
        this.loadAllChartData();
    }

    resizeCharts() {
        Object.values(this.charts).forEach(chart => {
            chart.resize();
        });
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
            'PLTR': alpha === 1 ? '#1f77b4' : `rgba(31, 119, 180, ${alpha})`,
            'TSLA': alpha === 1 ? '#ff7f0e' : `rgba(255, 127, 14, ${alpha})`,
            'BBAI': alpha === 1 ? '#2ca02c' : `rgba(44, 160, 44, ${alpha})`,
            'MSFT': alpha === 1 ? '#d62728' : `rgba(214, 39, 40, ${alpha})`,
            'NPWR': alpha === 1 ? '#9467bd' : `rgba(148, 103, 189, ${alpha})`
        };
        return colors[symbol] || (alpha === 1 ? '#1f77b4' : `rgba(31, 119, 180, ${alpha})`);
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
