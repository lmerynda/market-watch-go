// Market Watch Dashboard JavaScript
class MarketWatchDashboard {
    constructor() {
        this.charts = {};
        this.symbols = ['PLTR', 'TSLA', 'BBAI', 'MSFT', 'NPWR'];
        this.currentTimeRange = '1W';
        this.refreshInterval = null;
        this.updateInterval = 30000; // 30 seconds
        this.isLoading = false; // Track loading state
        this.loadingController = null; // AbortController for cancelling requests
        
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.initializeCharts();
        
        // Add a small delay to ensure DOM is fully rendered
        await new Promise(resolve => setTimeout(resolve, 100));
        
        this.syncTimeRangeFromHTML();
        console.log('Dashboard initialized with time range:', this.currentTimeRange);
        this.loadDashboardData();
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Time range selector
        document.querySelectorAll('input[name="timeRange"]').forEach(radio => {
            radio.addEventListener('change', (e) => {
                this.currentTimeRange = e.target.value;
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
                            borderWidth: 2
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
                                        return date.toLocaleString();
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
                                        minute: 'HH:mm',
                                        hour: 'HH:mm',
                                        day: 'MMM DD',
                                        week: 'MMM DD'
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
                this.updateVolumeStats(data.symbols);
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
            const url = `/api/volume/${symbol}/chart?range=${this.currentTimeRange}`;
            console.log(`Loading chart data for ${symbol} with URL: ${url}`);
            const response = await fetch(url, {
                signal: this.loadingController?.signal
            });
            const data = await response.json();
            
            console.log(`Chart data response for ${symbol}:`, data);
            
            if (response.ok && data.datasets && data.datasets.length > 0) {
                const chartData = data.datasets[0].data.map(point => ({
                    x: new Date(point.x),
                    y: point.y
                }));
                
                console.log(`Processed chart data for ${symbol}:`, chartData.length, 'points');
                
                this.charts[symbol].data.datasets[0].data = chartData;
                this.charts[symbol].update('none');
                
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

    updateVolumeStats(symbols) {
        const container = document.getElementById('volume-stats');
        if (!container || !symbols) return;

        container.innerHTML = '';
        
        symbols.forEach(stat => {
            const col = document.createElement('div');
            col.className = 'col-md-2 col-sm-4 col-6 mb-3';
            
            const ratio = stat.volume_ratio || 0;
            const changeClass = ratio > 1.2 ? 'positive' : ratio < 0.8 ? 'negative' : 'neutral';
            const changeText = ratio > 1 ? `+${((ratio - 1) * 100).toFixed(1)}%` : 
                             ratio < 1 ? `-${((1 - ratio) * 100).toFixed(1)}%` : '0%';
            
            col.innerHTML = `
                <div class="volume-stat-card">
                    <div class="volume-stat-value">${stat.current_volume.toLocaleString()}</div>
                    <div class="volume-stat-label">${stat.symbol}</div>
                    <div class="volume-stat-change ${changeClass}">${changeText}</div>
                </div>
            `;
            
            container.appendChild(col);
        });
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
