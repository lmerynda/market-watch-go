// Head & Shoulders Pattern Monitor - JavaScript
class HeadShouldersMonitor {
    constructor() {
        this.patterns = [];
        this.selectedPattern = null;
        this.currentFilter = 'all';
        this.tradingViewWidget = null;
        this.refreshInterval = null;
        this.isLoading = false;
        
        this.init();
    }

    async init() {
        console.log('Initializing Head & Shoulders Monitor...');
        
        this.setupEventListeners();
        this.setupThesisPanel();
        await this.loadPatterns();
        this.startAutoRefresh();
        
        console.log('Head & Shoulders Monitor initialized');
    }

    setupEventListeners() {
        // Pattern search
        document.getElementById('pattern-search').addEventListener('input', (e) => {
            this.filterPatterns(e.target.value);
        });

        // Filter dropdown
        document.querySelectorAll('[data-filter]').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentFilter = e.target.dataset.filter;
                this.filterPatterns();
            });
        });

        // Control buttons
        document.getElementById('refresh-patterns-btn').addEventListener('click', () => {
            this.loadPatterns();
        });

        document.getElementById('monitor-patterns-btn').addEventListener('click', () => {
            this.monitorAllPatterns();
        });

        document.getElementById('detect-pattern-btn').addEventListener('click', () => {
            this.detectPatternForSelected();
        });

        // Chart overlay toggles
        document.getElementById('toggle-pattern-overlay').addEventListener('click', () => {
            this.togglePatternOverlay();
        });

        document.getElementById('toggle-thesis-overlay').addEventListener('click', () => {
            this.toggleThesisPanel();
        });

        // Thesis panel
        document.getElementById('thesis-panel-header').addEventListener('click', () => {
            this.toggleThesisPanel();
        });

        // Component update modal
        document.getElementById('save-component-btn').addEventListener('click', () => {
            this.saveComponentUpdate();
        });

        // Confidence slider
        document.getElementById('update-confidence').addEventListener('input', (e) => {
            document.getElementById('confidence-display').textContent = e.target.value + '%';
        });
    }

    setupThesisPanel() {
        const panel = document.getElementById('thesis-panel');
        const toggle = document.getElementById('thesis-panel-toggle');
        
        // Initially collapsed
        panel.classList.remove('expanded');
        toggle.classList.add('bi-chevron-up');
        toggle.classList.remove('bi-chevron-down');
    }

    async loadPatterns() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        this.showLoading();
        
        try {
            console.log('Loading head and shoulders patterns...');
            
            const response = await fetch('/api/head-shoulders/patterns?pattern_type=inverse_head_shoulders');
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const patterns = await response.json();
            console.log(`Loaded ${patterns.length} patterns`);
            
            this.patterns = patterns;
            this.displayPatterns();
            this.updatePatternCount();
            
        } catch (error) {
            console.error('Failed to load patterns:', error);
            this.showError('Failed to load patterns: ' + error.message);
            this.patterns = [];
            this.displayNoPatterns();
        } finally {
            this.isLoading = false;
            this.hideLoading();
        }
    }

    displayPatterns() {
        const container = document.getElementById('patterns-list');
        
        if (this.patterns.length === 0) {
            this.displayNoPatterns();
            return;
        }

        const filteredPatterns = this.getFilteredPatterns();
        
        container.innerHTML = filteredPatterns.map(pattern => 
            this.createPatternListItem(pattern)
        ).join('');

        // Add click handlers
        container.querySelectorAll('.pattern-item').forEach(item => {
            item.addEventListener('click', () => {
                const patternId = parseInt(item.dataset.patternId);
                this.selectPattern(patternId);
            });
        });
    }

    displayNoPatterns() {
        const container = document.getElementById('patterns-list');
        container.innerHTML = `
            <div class="text-center p-4 text-muted">
                <i class="bi bi-graph-up-arrow display-4 opacity-25"></i>
                <p class="mt-3">No patterns found</p>
                <button class="btn btn-primary btn-sm" onclick="window.location.reload()">
                    <i class="bi bi-arrow-clockwise"></i> Refresh
                </button>
            </div>
        `;
    }

    createPatternListItem(pattern) {
        const timeAgo = this.getTimeAgo(new Date(pattern.detected_at));
        const completionPercent = pattern.thesis_components.completion_percent || 0;
        const phaseClass = this.getPhaseClass(pattern.current_phase);
        const phaseIcon = this.getPhaseIcon(pattern.current_phase);
        
        return `
            <div class="pattern-item" data-pattern-id="${pattern.id}" data-symbol="${pattern.symbol}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="d-flex align-items-center mb-2">
                            <h6 class="mb-0 pattern-symbol text-primary">${pattern.symbol}</h6>
                            <span class="badge pattern-phase ${phaseClass} ms-2">
                                <i class="bi bi-${phaseIcon} me-1"></i>${this.formatPhase(pattern.current_phase)}
                            </span>
                        </div>
                        <div class="d-flex justify-content-between text-sm mb-1">
                            <span class="text-muted">Progress:</span>
                            <span class="fw-bold text-tv-primary">
                                ${pattern.thesis_components.completed_components}/${pattern.thesis_components.total_components}
                            </span>
                        </div>
                        <div class="pattern-progress">
                            <div class="pattern-progress-fill" style="width: ${completionPercent}%"></div>
                        </div>
                    </div>
                    <div class="text-end ms-3">
                        <div class="text-xs text-muted mb-1">${timeAgo}</div>
                        <div class="pattern-indicators">
                            ${this.getPatternIndicators(pattern)}
                        </div>
                    </div>
                </div>
                <div class="mt-2 pt-2 border-top border-tv">
                    <small class="text-tv-secondary">
                        Target: <span class="text-success">$${this.calculateTarget(pattern).toFixed(2)}</span> |
                        Symmetry: <span class="text-info">${pattern.symmetry.toFixed(1)}%</span>
                    </small>
                </div>
            </div>
        `;
    }

    calculateTarget(pattern) {
        if (pattern.pattern_type === 'inverse_head_shoulders') {
            return pattern.neckline_level + pattern.pattern_height;
        }
        return pattern.neckline_level - pattern.pattern_height;
    }

    getPhaseClass(phase) {
        switch (phase) {
            case 'formation': return 'formation';
            case 'breakout': return 'breakout';
            case 'target_pursuit': return 'target_pursuit';
            case 'completed': return 'completed';
            default: return 'formation';
        }
    }

    getPhaseIcon(phase) {
        switch (phase) {
            case 'formation': return 'diagram-3';
            case 'breakout': return 'arrow-up-circle';
            case 'target_pursuit': return 'bullseye';
            case 'completed': return 'check-circle';
            default: return 'diagram-3';
        }
    }

    formatPhase(phase) {
        switch (phase) {
            case 'formation': return 'Formation';
            case 'breakout': return 'Breakout';
            case 'target_pursuit': return 'Target Pursuit';
            case 'completed': return 'Completed';
            default: return phase;
        }
    }

    getPatternIndicators(pattern) {
        const indicators = [];
        
        // Volume indicator
        if (pattern.thesis_components.breakout_volume?.is_completed) {
            indicators.push('<i class="bi bi-bar-chart-fill text-volume" title="Volume Confirmed"></i>');
        } else {
            indicators.push('<i class="bi bi-bar-chart text-muted" title="Volume Pending"></i>');
        }
        
        // Breakout indicator
        if (pattern.thesis_components.neckline_breakout?.is_completed) {
            indicators.push('<i class="bi bi-arrow-up-circle-fill text-breakout" title="Breakout Confirmed"></i>');
        } else {
            indicators.push('<i class="bi bi-arrow-up-circle text-muted" title="Breakout Pending"></i>');
        }
        
        // Target indicator
        if (pattern.thesis_components.full_target?.is_completed) {
            indicators.push('<i class="bi bi-bullseye text-target" title="Target Reached"></i>');
        } else if (pattern.thesis_components.partial_fill_t1?.is_completed) {
            indicators.push('<i class="bi bi-record-circle text-target" title="Partial Target"></i>');
        } else {
            indicators.push('<i class="bi bi-circle text-muted" title="Target Pending"></i>');
        }
        
        return indicators.join('');
    }

    getFilteredPatterns() {
        let filtered = this.patterns;
        
        // Apply phase filter
        if (this.currentFilter !== 'all') {
            filtered = filtered.filter(pattern => pattern.current_phase === this.currentFilter);
        }
        
        // Apply search filter
        const searchTerm = document.getElementById('pattern-search').value.toLowerCase();
        if (searchTerm) {
            filtered = filtered.filter(pattern => 
                pattern.symbol.toLowerCase().includes(searchTerm)
            );
        }
        
        return filtered;
    }

    filterPatterns(searchTerm = '') {
        this.displayPatterns();
    }

    async selectPattern(patternId) {
        console.log(`Selecting pattern ${patternId}`);
        
        try {
            // Find pattern in local array first
            const pattern = this.patterns.find(p => p.id === patternId);
            if (!pattern) {
                throw new Error('Pattern not found');
            }
            
            // Update selected state in UI
            document.querySelectorAll('.pattern-item').forEach(item => {
                item.classList.remove('selected');
            });
            document.querySelector(`[data-pattern-id="${patternId}"]`).classList.add('selected');
            
            // Update header
            document.getElementById('selected-symbol').textContent = pattern.symbol;
            document.getElementById('pattern-status').textContent = `${this.formatPhase(pattern.current_phase)} - ${pattern.thesis_components.completed_components}/${pattern.thesis_components.total_components} components`;
            
            // Enable detect button
            document.getElementById('detect-pattern-btn').disabled = false;
            
            // Load chart
            this.loadTradingViewChart(pattern.symbol);
            
            // Load thesis components
            this.loadThesisComponents(pattern);
            
            this.selectedPattern = pattern;
            
        } catch (error) {
            console.error('Failed to select pattern:', error);
            this.showError('Failed to select pattern: ' + error.message);
        }
    }

    loadTradingViewChart(symbol) {
        console.log(`Loading TradingView chart for ${symbol}`);
        
        const container = document.getElementById('tradingview-widget');
        container.innerHTML = '';
        
        try {
            this.tradingViewWidget = new TradingView.widget({
                "width": "100%",
                "height": "100%",
                "symbol": `NASDAQ:${symbol}`,
                "interval": "D",
                "timezone": "Etc/UTC",
                "theme": "dark",
                "style": "1",
                "locale": "en",
                "toolbar_bg": "#131722",
                "enable_publishing": false,
                "hide_top_toolbar": false,
                "hide_legend": false,
                "save_image": false,
                "container_id": "tradingview-widget",
                "studies": [
                    "Volume@tv-basicstudies",
                    "RSI@tv-basicstudies"
                ]
            });
            
        } catch (error) {
            console.error('Failed to load TradingView chart:', error);
            container.innerHTML = `
                <div class="d-flex align-items-center justify-content-center h-100 text-muted">
                    <div class="text-center">
                        <i class="bi bi-exclamation-triangle display-1 opacity-25"></i>
                        <h5 class="mt-3">Chart Loading Failed</h5>
                        <p>Unable to load chart for ${symbol}</p>
                    </div>
                </div>
            `;
        }
    }

    loadThesisComponents(pattern) {
        const container = document.getElementById('thesis-content');
        const thesis = pattern.thesis_components;
        
        // Update thesis panel header
        document.getElementById('thesis-symbol').textContent = `- ${pattern.symbol}`;
        document.getElementById('thesis-progress').style.width = `${thesis.completion_percent}%`;
        document.getElementById('thesis-completion').textContent = `${thesis.completed_components}/${thesis.total_components}`;
        
        // Create thesis sections
        const sections = [
            {
                title: 'Pattern Formation',
                icon: 'diagram-3',
                components: [
                    thesis.left_shoulder_formed,
                    thesis.head_formed,
                    thesis.head_lower_low,
                    thesis.right_shoulder_formed,
                    thesis.right_shoulder_symmetry,
                    thesis.neckline_established
                ]
            },
            {
                title: 'Volume Analysis',
                icon: 'bar-chart-fill',
                components: [
                    thesis.left_shoulder_volume,
                    thesis.head_volume_spike,
                    thesis.right_shoulder_volume,
                    thesis.breakout_volume
                ]
            },
            {
                title: 'Breakout Confirmation',
                icon: 'arrow-up-circle',
                components: [
                    thesis.neckline_breakout,
                    thesis.neckline_retest
                ]
            },
            {
                title: 'Target Achievement',
                icon: 'bullseye',
                components: [
                    thesis.target_projected,
                    thesis.partial_fill_t1,
                    thesis.partial_fill_t2,
                    thesis.full_target
                ]
            }
        ];
        
        container.innerHTML = sections.map(section => 
            this.createThesisSection(section, pattern.id)
        ).join('');
    }

    createThesisSection(section, patternId) {
        const completedCount = section.components.filter(c => c && c.is_completed).length;
        const totalCount = section.components.filter(c => c && c.name).length;
        
        return `
            <div class="thesis-section">
                <h6 class="section-title d-flex align-items-center">
                    <i class="bi bi-${section.icon} me-2"></i>${section.title}
                    <span class="badge bg-secondary ms-2">${completedCount}/${totalCount}</span>
                </h6>
                <div class="row">
                    ${section.components.map(component => 
                        component && component.name ? this.createThesisComponent(component, patternId) : ''
                    ).join('')}
                </div>
            </div>
        `;
    }

    createThesisComponent(component, patternId) {
        if (!component || !component.name) return '';
        
        const statusIcon = component.is_completed ? 'check-circle-fill' : 
                          component.is_required ? 'exclamation-circle' : 'circle';
        const statusClass = component.is_completed ? 'text-success' : 
                           component.is_required ? 'text-warning' : 'text-muted';
        
        const completedAt = component.completed_at ? 
            new Date(component.completed_at).toLocaleDateString() : '';
        
        return `
            <div class="col-md-6 mb-2">
                <div class="thesis-component ${component.is_completed ? 'completed' : ''} ${component.is_required ? 'required' : ''}"
                     onclick="window.hsMonitor.openComponentModal('${component.name}', ${patternId})">
                    <div class="component-status">
                        <i class="bi bi-${statusIcon} ${statusClass}"></i>
                        <span class="flex-grow-1">${component.name}</span>
                        ${completedAt ? `<small class="text-muted">${completedAt}</small>` : ''}
                    </div>
                    ${component.description ? `<div class="component-evidence">${component.description}</div>` : ''}
                    ${component.evidence && component.evidence.length > 0 ? 
                        `<div class="component-evidence">• ${component.evidence.join('<br>• ')}</div>` : ''}
                    ${component.confidence_level > 0 ? 
                        `<div class="component-confidence">Confidence: ${component.confidence_level.toFixed(0)}%</div>` : ''}
                </div>
            </div>
        `;
    }

    openComponentModal(componentName, patternId) {
        const pattern = this.patterns.find(p => p.id === patternId);
        if (!pattern) return;
        
        // Find the component
        const allComponents = pattern.thesis_components;
        let component = null;
        
        // Search through all component properties
        for (const key in allComponents) {
            if (allComponents[key] && allComponents[key].name === componentName) {
                component = allComponents[key];
                break;
            }
        }
        
        if (!component) {
            console.error('Component not found:', componentName);
            return;
        }
        
        // Populate modal
        document.getElementById('update-pattern-id').value = patternId;
        document.getElementById('update-component-name').value = componentName;
        document.getElementById('display-component-name').value = componentName;
        document.getElementById('update-is-completed').checked = component.is_completed;
        document.getElementById('update-confidence').value = component.confidence_level || 50;
        document.getElementById('confidence-display').textContent = (component.confidence_level || 50) + '%';
        document.getElementById('update-evidence').value = component.evidence ? component.evidence.join('\n') : '';
        document.getElementById('update-notes').value = '';
        
        // Show modal
        const modal = new bootstrap.Modal(document.getElementById('component-update-modal'));
        modal.show();
    }

    async saveComponentUpdate() {
        const patternId = document.getElementById('update-pattern-id').value;
        const componentName = document.getElementById('update-component-name').value;
        const isCompleted = document.getElementById('update-is-completed').checked;
        const confidenceLevel = parseFloat(document.getElementById('update-confidence').value);
        const evidenceText = document.getElementById('update-evidence').value;
        const notes = document.getElementById('update-notes').value;
        
        const evidence = evidenceText.split('\n').filter(line => line.trim());
        
        try {
            const response = await fetch(`/api/head-shoulders/pattern/${patternId}/thesis/${encodeURIComponent(componentName)}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    is_completed: isCompleted,
                    confidence_level: confidenceLevel,
                    evidence: evidence,
                    notes: notes
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            console.log('Component updated successfully:', result);
            
            this.showSuccess('Component updated successfully');
            
            // Close modal
            const modal = bootstrap.Modal.getInstance(document.getElementById('component-update-modal'));
            modal.hide();
            
            // Refresh the pattern to show updated data
            await this.loadPatterns();
            if (this.selectedPattern) {
                this.selectPattern(this.selectedPattern.id);
            }
            
        } catch (error) {
            console.error('Failed to update component:', error);
            this.showError('Failed to update component: ' + error.message);
        }
    }

    togglePatternOverlay() {
        const overlay = document.getElementById('pattern-overlay');
        const btn = document.getElementById('toggle-pattern-overlay');
        
        if (overlay.classList.contains('d-none')) {
            overlay.classList.remove('d-none');
            btn.innerHTML = '<i class="bi bi-eye-slash"></i> Hide Pattern';
            this.drawPatternOverlay();
        } else {
            overlay.classList.add('d-none');
            btn.innerHTML = '<i class="bi bi-eye"></i> Show Pattern';
        }
    }

    drawPatternOverlay() {
        if (!this.selectedPattern) return;
        
        const svg = document.getElementById('pattern-svg');
        // TODO: Implement pattern overlay drawing
        // This would require mapping pattern points to chart coordinates
        console.log('Drawing pattern overlay for', this.selectedPattern.symbol);
    }

    toggleThesisPanel() {
        const panel = document.getElementById('thesis-panel');
        const toggle = document.getElementById('thesis-panel-toggle');
        
        if (panel.classList.contains('expanded')) {
            panel.classList.remove('expanded');
            toggle.classList.add('bi-chevron-up');
            toggle.classList.remove('bi-chevron-down');
        } else {
            panel.classList.add('expanded');
            toggle.classList.remove('bi-chevron-up');
            toggle.classList.add('bi-chevron-down');
        }
    }

    async monitorAllPatterns() {
        try {
            this.showLoading('Monitoring patterns...');
            
            const response = await fetch('/api/head-shoulders/patterns/monitor', {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            console.log('Pattern monitoring completed:', result);
            
            this.showSuccess(`Monitoring completed. ${result.patterns_updated} patterns updated.`);
            
            // Refresh patterns to show updates
            await this.loadPatterns();
            
        } catch (error) {
            console.error('Failed to monitor patterns:', error);
            this.showError('Failed to monitor patterns: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    async detectPatternForSelected() {
        if (!this.selectedPattern) return;
        
        try {
            this.showLoading('Detecting pattern...');
            
            const response = await fetch(`/api/head-shoulders/symbols/${this.selectedPattern.symbol}/detect`, {
                method: 'POST'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const result = await response.json();
            console.log('Pattern detection completed:', result);
            
            this.showSuccess(`Pattern detection completed for ${this.selectedPattern.symbol}`);
            
            // Refresh patterns
            await this.loadPatterns();
            
        } catch (error) {
            console.error('Failed to detect pattern:', error);
            this.showError('Failed to detect pattern: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    updatePatternCount() {
        const count = this.getFilteredPatterns().length;
        document.getElementById('pattern-count').textContent = count;
    }

    startAutoRefresh() {
        // Refresh patterns every 5 minutes
        this.refreshInterval = setInterval(() => {
            console.log('Auto-refreshing patterns...');
            this.loadPatterns();
        }, 5 * 60 * 1000);
    }

    // Utility methods
    getTimeAgo(date) {
        const now = new Date();
        const diffMs = now - date;
        const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
        const diffDays = Math.floor(diffHours / 24);
        
        if (diffDays > 0) {
            return `${diffDays}d ago`;
        } else if (diffHours > 0) {
            return `${diffHours}h ago`;
        } else {
            return 'Recent';
        }
    }

    showLoading(message = 'Loading...') {
        const loadingHtml = `
            <div class="text-center p-4 text-muted">
                <div class="loading-spinner me-2"></div>
                ${message}
            </div>
        `;
        document.getElementById('patterns-list').innerHTML = loadingHtml;
    }

    hideLoading() {
        // Loading will be hidden when patterns are displayed
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

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM loaded, initializing Head & Shoulders Monitor...');
    window.hsMonitor = new HeadShouldersMonitor();
});
