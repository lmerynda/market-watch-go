// Pattern Watcher - Multi-Pattern Analysis Interface
class PatternWatcher {
  constructor() {
    this.symbols = [];
    this.patterns = [];
    this.selectedSymbol = null;
    this.selectedPattern = null;
    this.currentFilter = "all";
    this.currentWatchlistFilter = "all";
    this.tradingViewWidget = null;
    this.refreshInterval = null;
    this.isLoading = false;

    this.init();
  }

  async init() {
    console.log("Initializing Pattern Watcher...");

    this.setupEventListeners();
    this.setupThesisPanel();
    await this.loadSymbols();
    await this.loadPatterns();
    this.startAutoRefresh();

    console.log("Pattern Watcher initialized");
  }

  setupEventListeners() {
    // Symbol search
    document.getElementById("symbol-search").addEventListener("input", (e) => {
      this.filterSymbols(e.target.value);
    });

    // Watchlist filters
    document
      .querySelectorAll('input[name="watchlist-filter"]')
      .forEach((radio) => {
        radio.addEventListener("change", (e) => {
          this.currentWatchlistFilter = e.target.value;
          this.filterSymbols();
        });
      });

    // Add symbol
    document.getElementById("add-symbol-btn").addEventListener("click", () => {
      this.showAddSymbolModal();
    });

    document
      .getElementById("add-symbol-input")
      .addEventListener("keypress", (e) => {
        if (e.key === "Enter") {
          this.quickAddSymbol();
        }
      });

    // Filter dropdown
    document.querySelectorAll("[data-filter]").forEach((item) => {
      item.addEventListener("click", (e) => {
        e.preventDefault();
        this.currentFilter = e.target.dataset.filter;
        this.filterPatterns();
      });
    });

    // Control buttons
    document
      .getElementById("refresh-symbols-btn")
      .addEventListener("click", () => {
        this.loadSymbols();
      });

    document
      .getElementById("scan-patterns-btn")
      .addEventListener("click", () => {
        this.scanAllPatterns();
      });

    document
      .getElementById("detect-pattern-btn")
      .addEventListener("click", () => {
        this.detectPatternsForSelected();
      });

    // Chart overlay toggles
    document
      .getElementById("toggle-pattern-overlay")
      .addEventListener("click", () => {
        this.togglePatternOverlay();
      });

    document
      .getElementById("toggle-annotations")
      .addEventListener("click", () => {
        this.toggleAnnotations();
      });

    // Thesis panel
    document
      .getElementById("thesis-panel-header")
      .addEventListener("click", () => {
        this.toggleThesisPanel();
      });

    // Save buttons
    document.getElementById("save-symbol-btn").addEventListener("click", () => {
      this.saveSymbol();
    });

    document
      .getElementById("save-component-btn")
      .addEventListener("click", () => {
        this.saveComponentUpdate();
      });

    // Confidence slider
    document
      .getElementById("update-confidence")
      .addEventListener("input", (e) => {
        document.getElementById("confidence-display").textContent =
          e.target.value + "%";
      });
  }

  setupThesisPanel() {
    const panel = document.getElementById("thesis-panel");
    const toggle = document.getElementById("thesis-panel-toggle");

    // Initially collapsed
    panel.classList.remove("expanded");
    toggle.classList.add("bi-chevron-up");
    toggle.classList.remove("bi-chevron-down");
  }

  async loadSymbols() {
    if (this.isLoading) return;

    this.isLoading = true;
    this.showSymbolsLoading();

    try {
      console.log("Loading symbols...");

      // Get watched symbols from the existing API
      const response = await fetch("/api/symbols");

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      console.log(
        `Loaded ${data && data.symbols ? data.symbols.length : 0} symbols`
      );

      this.symbols = data && data.symbols ? data.symbols : [];
      this.displaySymbols();
      this.updateSymbolCount();
    } catch (error) {
      console.error("Failed to load symbols:", error);
      this.showError("Failed to load symbols: " + error.message);
      this.symbols = [];
      this.displayNoSymbols();
    } finally {
      this.isLoading = false;
    }
  }

  async loadPatterns() {
    try {
      console.log("Loading patterns...");

      const response = await fetch("/api/head-shoulders/patterns");

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const patterns = await response.json();
      console.log(`Loaded ${patterns ? patterns.length : 0} patterns`);

      this.patterns = patterns || [];
      this.displayPatterns();
      this.updatePatternCount();

      // Refresh symbol display to show updated pattern information
      this.displaySymbols();
    } catch (error) {
      console.error("Failed to load patterns:", error);
      this.showError("Failed to load patterns: " + error.message);
      this.patterns = [];
      this.displayNoPatterns();
    }
  }

  displaySymbols() {
    const container = document.getElementById("symbols-list");

    if (this.symbols.length === 0) {
      this.displayNoSymbols();
      return;
    }

    const filteredSymbols = this.getFilteredSymbols();

    container.innerHTML = filteredSymbols
      .map((symbol) => this.createSymbolListItem(symbol))
      .join("");

    // Add click handlers
    container.querySelectorAll(".symbol-item").forEach((item) => {
      item.addEventListener("click", () => {
        const symbol = item.dataset.symbol;
        this.selectSymbol(symbol);
      });
    });
  }

  displayNoSymbols() {
    const container = document.getElementById("symbols-list");
    container.innerHTML = `
            <div class="text-center p-4 text-muted">
                <i class="bi bi-list-ul display-4 opacity-25"></i>
                <p class="mt-3">No symbols in watchlist</p>
                <button class="btn btn-primary btn-sm" onclick="window.patternWatcher.showAddSymbolModal()">
                    <i class="bi bi-plus"></i> Add Symbol
                </button>
            </div>
        `;
  }

  createSymbolListItem(symbolData) {
    // Handle both string symbols and symbol objects
    const symbol =
      typeof symbolData === "string"
        ? symbolData
        : symbolData.symbol || symbolData.Symbol;
    const symbolName =
      typeof symbolData === "object"
        ? symbolData.name || symbolData.Name || ""
        : "";

    const symbolPatterns = this.patterns.filter((p) => p.symbol === symbol);
    const patternBadges = this.createPatternBadges(symbolPatterns);

    return `
            <div class="symbol-item" data-symbol="${symbol}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="symbol-name">${symbol}</div>
                        ${
                          symbolName
                            ? `<div class="text-muted small">${symbolName}</div>`
                            : ""
                        }
                        ${patternBadges}
                        <div class="symbol-stats">
                            <div class="stat-item">
                                <div class="stat-value">${
                                  symbolPatterns.length
                                }</div>
                                <div>Patterns</div>
                            </div>
                            <div class="stat-item">
                                <div class="stat-value">${
                                  symbolPatterns.filter((p) => !p.is_complete)
                                    .length
                                }</div>
                                <div>Active</div>
                            </div>
                            <div class="stat-item">
                                <div class="stat-value">${this.getAvgCompletion(
                                  symbolPatterns
                                )}%</div>
                                <div>Avg Progress</div>
                            </div>
                        </div>
                    </div>
                    <div class="symbol-status ${this.getSymbolStatus(
                      symbolPatterns
                    )}"></div>
                </div>
            </div>
        `;
  }

  createPatternBadges(patterns) {
    if (patterns.length === 0) {
      return '<div class="pattern-badges"><span class="pattern-badge">No patterns</span></div>';
    }

    const uniqueTypes = [...new Set(patterns.map((p) => p.pattern_type))];
    const badges = uniqueTypes
      .map(
        (type) =>
          `<span class="pattern-badge ${type.replace(
            /_/g,
            "-"
          )}">${this.formatPatternType(type)}</span>`
      )
      .join("");

    return `<div class="pattern-badges">${badges}</div>`;
  }

  displayPatterns() {
    // Since we removed the patterns-list panel, we don't need to display patterns separately
    // Patterns are now integrated into the symbol list display
    console.log(`Loaded ${this.patterns.length} patterns`);
  }

  displayNoPatterns() {
    // No longer needed since patterns are integrated into symbol display
    console.log("No patterns to display");
  }

  createPatternListItem(pattern) {
    const timeAgo = this.getTimeAgo(new Date(pattern.detected_at));
    const completionPercent =
      pattern.thesis_components?.completion_percent || 0;
    const phaseClass = this.getPhaseClass(pattern.current_phase);
    const phaseIcon = this.getPhaseIcon(pattern.current_phase);

    return `
            <div class="pattern-item" data-pattern-id="${
              pattern.id
            }" data-symbol="${pattern.symbol}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="d-flex align-items-center mb-2">
                            <h6 class="mb-0 pattern-type pattern-type-${pattern.pattern_type.replace(
                              /_/g,
                              "-"
                            )}">${pattern.symbol}</h6>
                            <span class="badge pattern-phase ${phaseClass} ms-2">
                                <i class="bi bi-${phaseIcon} me-1"></i>${this.formatPhase(
      pattern.current_phase
    )}
                            </span>
                        </div>
                        <div class="small text-muted mb-1">${this.formatPatternType(
                          pattern.pattern_type
                        )}</div>
                        <div class="d-flex justify-content-between text-sm mb-1">
                            <span class="text-muted">Progress:</span>
                            <span class="fw-bold text-pw-primary">
                                ${
                                  pattern.thesis_components
                                    ?.completed_components || 0
                                }/${
      pattern.thesis_components?.total_components || 0
    }
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
            </div>
        `;
  }

  async selectSymbol(symbol) {
    console.log(`Selecting symbol ${symbol}`);

    try {
      // Update selected state in UI
      document.querySelectorAll(".symbol-item").forEach((item) => {
        item.classList.remove("selected");
      });
      document
        .querySelector(`[data-symbol="${symbol}"]`)
        .classList.add("selected");

      // Update header
      document.getElementById("selected-symbol").textContent = symbol;
      document.getElementById(
        "pattern-info"
      ).textContent = `Ready to analyze patterns for ${symbol}`;

      // Enable detect button
      document.getElementById("detect-pattern-btn").disabled = false;

      // Load chart
      this.loadTradingViewChart(symbol);

      // Filter patterns for this symbol
      this.filterPatternsForSymbol(symbol);

      this.selectedSymbol = symbol;
    } catch (error) {
      console.error("Failed to select symbol:", error);
      this.showError("Failed to select symbol: " + error.message);
    }
  }

  async selectPattern(patternId) {
    console.log(`Selecting pattern ${patternId}`);

    try {
      const pattern = this.patterns.find((p) => p.id === patternId);
      if (!pattern) {
        throw new Error("Pattern not found");
      }

      // Update selected state in UI
      document.querySelectorAll(".pattern-item").forEach((item) => {
        item.classList.remove("selected");
      });
      document
        .querySelector(`[data-pattern-id="${patternId}"]`)
        .classList.add("selected");

      // Load thesis components
      this.loadThesisComponents(pattern);

      this.selectedPattern = pattern;
    } catch (error) {
      console.error("Failed to select pattern:", error);
      this.showError("Failed to select pattern: " + error.message);
    }
  }

  loadTradingViewChart(symbol) {
    console.log(`Loading TradingView chart for ${symbol}`);

    const container = document.getElementById("tradingview-widget");
    if (!container) {
      console.error("TradingView container not found!");
      return;
    }

    // Clear container
    container.innerHTML = "";

    // Show loading state
    container.innerHTML = `
      <div class="d-flex align-items-center justify-content-center h-100 text-muted">
        <div class="text-center">
          <div class="loading-spinner me-2"></div>
          <p>Loading chart for ${symbol}...</p>
        </div>
      </div>
    `;

    // Check if TradingView is available
    if (typeof TradingView === "undefined") {
      console.error("TradingView library not loaded!");
      container.innerHTML = `
        <div class="d-flex align-items-center justify-content-center h-100 text-muted">
          <div class="text-center">
            <i class="bi bi-exclamation-triangle display-1 opacity-25"></i>
            <h5 class="mt-3">TradingView Not Available</h5>
            <p>Please check your internet connection</p>
          </div>
        </div>
      `;
      return;
    }

    try {
      // Wait a bit for the container to be ready
      setTimeout(() => {
        this.createTradingViewWidget(symbol, container);
      }, 100);
    } catch (error) {
      console.error("Failed to load TradingView chart:", error);
      container.innerHTML = `
        <div class="d-flex align-items-center justify-content-center h-100 text-muted">
          <div class="text-center">
            <i class="bi bi-exclamation-triangle display-1 opacity-25"></i>
            <h5 class="mt-3">Chart Loading Failed</h5>
            <p>Unable to load chart for ${symbol}</p>
            <small class="text-danger">${error.message}</small>
          </div>
        </div>
      `;
    }
  }

  createTradingViewWidget(
    symbol,
    container,
    fallbackExchanges = ["NASDAQ", "NYSE", "AMEX"]
  ) {
    // Get container dimensions
    const containerWidth = container.clientWidth;
    const containerHeight = container.clientHeight;

    console.log(`Trying to load ${symbol}`);

    this.tradingViewWidget = new TradingView.widget({
      width: containerWidth || "100%",
      height: containerHeight || 500,
      symbol: symbol,
      interval: "15",
      timezone: "America/New_York",
      theme: "dark",
      style: "1",
      locale: "en",
      toolbar_bg: "#131722",
      enable_publishing: false,
      hide_top_toolbar: false,
      hide_legend: false,
      save_image: false,
      container_id: "tradingview-widget",
      autosize: true,
      show_popup_button: true,
      popup_width: "1000",
      popup_height: "650",
      details: true,
      hotlist: true,
      calendar: false,
      studies: ["Volume@tv-basicstudies", "RSI@tv-basicstudies"],
      overrides: {
        "paneProperties.background": "#131722",
        "paneProperties.vertGridProperties.color": "#363a45",
        "paneProperties.horzGridProperties.color": "#363a45",
        "symbolWatermarkProperties.transparency": 90,
        "scalesProperties.textColor": "#787b86",
        "scalesProperties.lineColor": "#363a45",
      },
      disabled_features: ["header_symbol_search", "symbol_search_hot_key"],
      enabled_features: ["study_templates"],
      onChartReady: () => {
        console.log(`TradingView widget loaded successfully for ${symbol}`);
      },
      // If symbol fails to load, TradingView will show an error in the widget
    });

    console.log("TradingView widget created successfully");
  }

  loadThesisComponents(pattern) {
    const container = document.getElementById("thesis-content");
    const thesis = pattern.thesis_components;

    if (!thesis) {
      container.innerHTML = `
                <div class="text-center text-muted p-4">
                    <i class="bi bi-list-check display-4 opacity-25"></i>
                    <p class="mt-3">No thesis data available for this pattern</p>
                </div>
            `;
      return;
    }

    // Update thesis panel header
    document.getElementById("thesis-symbol").textContent = `- ${
      pattern.symbol
    } (${this.formatPatternType(pattern.pattern_type)})`;
    document.getElementById("thesis-progress").style.width = `${
      thesis.completion_percent || 0
    }%`;
    document.getElementById("thesis-completion").textContent = `${
      thesis.completed_components || 0
    }/${thesis.total_components || 0}`;

    // Create simplified thesis display
    container.innerHTML = `
            <div class="thesis-section">
                <h6 class="section-title">Pattern Analysis Progress</h6>
                <div class="mb-3">
                    <div class="d-flex justify-content-between">
                        <span>Pattern Type:</span>
                        <span class="text-pw-primary">${this.formatPatternType(
                          pattern.pattern_type
                        )}</span>
                    </div>
                    <div class="d-flex justify-content-between">
                        <span>Current Phase:</span>
                        <span class="text-pw-primary">${this.formatPhase(
                          pattern.current_phase
                        )}</span>
                    </div>
                    <div class="d-flex justify-content-between">
                        <span>Completion:</span>
                        <span class="text-pw-primary">${
                          thesis.completion_percent || 0
                        }%</span>
                    </div>
                </div>
                <div class="pattern-progress">
                    <div class="pattern-progress-fill" style="width: ${
                      thesis.completion_percent || 0
                    }%"></div>
                </div>
                <small class="text-muted mt-2 d-block">
                    Click on individual components to update their status manually
                </small>
            </div>
        `;
  }

  // Helper methods
  getFilteredSymbols() {
    const searchTerm = document
      .getElementById("symbol-search")
      .value.toLowerCase();

    return this.symbols.filter((symbolData) => {
      // Handle both string symbols and symbol objects
      const symbol =
        typeof symbolData === "string"
          ? symbolData
          : symbolData.symbol || symbolData.Symbol;

      // Check search term match
      const matchesSearch = symbol && symbol.toLowerCase().includes(searchTerm);

      // Check watchlist filter
      const symbolPatterns = this.patterns.filter((p) => p.symbol === symbol);
      let matchesFilter = true;

      switch (this.currentWatchlistFilter) {
        case "with-patterns":
          matchesFilter = symbolPatterns.length > 0;
          break;
        case "forming":
          matchesFilter = symbolPatterns.some(
            (p) =>
              p.current_phase === "formation" || p.current_phase === "breakout"
          );
          break;
        case "all":
        default:
          matchesFilter = true;
          break;
      }

      return matchesSearch && matchesFilter;
    });
  }

  getFilteredPatterns() {
    let filtered = this.patterns;

    if (this.currentFilter !== "all") {
      filtered = filtered.filter(
        (pattern) => pattern.pattern_type === this.currentFilter
      );
    }

    return filtered;
  }

  filterSymbols(searchTerm = "") {
    this.displaySymbols();
  }

  filterPatterns() {
    this.displayPatterns();
  }

  filterPatternsForSymbol(symbol) {
    const symbolPatterns = this.patterns.filter((p) => p.symbol === symbol);
    // Update pattern display for selected symbol
    this.displayPatterns();
  }

  formatPatternType(type) {
    return type.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());
  }

  formatPhase(phase) {
    return phase.replace(/_/g, " ").replace(/\b\w/g, (l) => l.toUpperCase());
  }

  getPhaseClass(phase) {
    switch (phase) {
      case "formation":
        return "formation";
      case "breakout":
        return "breakout";
      case "target_pursuit":
        return "target_pursuit";
      case "completed":
        return "completed";
      default:
        return "formation";
    }
  }

  getPhaseIcon(phase) {
    switch (phase) {
      case "formation":
        return "diagram-3";
      case "breakout":
        return "arrow-up-circle";
      case "target_pursuit":
        return "bullseye";
      case "completed":
        return "check-circle";
      default:
        return "diagram-3";
    }
  }

  getPatternIndicators(pattern) {
    // Simplified indicators
    return `
            <i class="bi bi-graph-up text-info" title="Pattern Detected"></i>
            <i class="bi bi-clock text-muted" title="In Progress"></i>
        `;
  }

  getSymbolStatus(patterns) {
    if (patterns.length === 0) return "no-patterns";
    if (patterns.some((p) => !p.is_complete)) return "patterns-found";
    return "patterns-found";
  }

  getAvgCompletion(patterns) {
    if (patterns.length === 0) return 0;
    const total = patterns.reduce(
      (sum, p) => sum + (p.thesis_components?.completion_percent || 0),
      0
    );
    return Math.round(total / patterns.length);
  }

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
      return "Recent";
    }
  }

  // Action methods
  showAddSymbolModal() {
    const modal = new bootstrap.Modal(
      document.getElementById("add-symbol-modal")
    );

    // Pre-fill the modal input if there's text in the quick add input
    const quickAddInput = document.getElementById("add-symbol-input");
    const modalInput = document.getElementById("symbol-input");

    if (quickAddInput.value.trim()) {
      modalInput.value = quickAddInput.value.trim().toUpperCase();
      quickAddInput.value = ""; // Clear quick add input
    }

    modal.show();

    // Focus on the input field when modal opens
    setTimeout(() => {
      modalInput.focus();
    }, 300);
  }

  async quickAddSymbol() {
    const input = document.getElementById("add-symbol-input");
    const symbol = input.value.trim().toUpperCase();

    if (!symbol) return;

    try {
      this.showLoading("Adding symbol and scanning for patterns...");

      const response = await fetch("/api/symbols", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ symbol: symbol }),
      });

      if (response.ok) {
        input.value = "";
        this.showSuccess(
          `Symbol ${symbol} added to watchlist. Automatic pattern detection started.`
        );
        await this.loadSymbols();

        // Refresh patterns after a short delay to allow detection to run
        setTimeout(() => {
          this.loadPatterns();
        }, 5000);
      } else {
        throw new Error("Failed to add symbol");
      }
    } catch (error) {
      this.showError("Failed to add symbol: " + error.message);
    } finally {
      this.hideLoading();
    }
  }

  async scanAllPatterns() {
    try {
      this.showLoading("Scanning all symbols for patterns...");

      // Trigger pattern detection for all symbols
      const response = await fetch("/api/head-shoulders/patterns/monitor", {
        method: "POST",
      });

      if (response.ok) {
        const result = await response.json();
        this.showSuccess(
          `Pattern scan completed. ${
            result.patterns_updated || "Multiple"
          } patterns updated.`
        );

        // Refresh patterns after scan
        setTimeout(() => {
          this.loadPatterns();
        }, 2000);
      } else {
        throw new Error("Pattern scan failed");
      }
    } catch (error) {
      this.showError("Failed to scan patterns: " + error.message);
    } finally {
      this.hideLoading();
    }
  }

  async detectPatternsForSelected() {
    if (!this.selectedSymbol) return;

    try {
      this.showLoading("Detecting patterns...");

      const response = await fetch(
        `/api/head-shoulders/symbols/${this.selectedSymbol}/detect`,
        {
          method: "POST",
        }
      );

      if (response.ok) {
        const result = await response.json();

        if (result.status === "pattern_detected") {
          this.showSuccess(`Pattern detected for ${this.selectedSymbol}!`);
          await this.loadPatterns();
        } else if (result.status === "no_pattern") {
          this.showInfo(
            `No patterns found for ${this.selectedSymbol}. ${result.message}`
          );
        } else {
          this.showSuccess(
            `Pattern detection completed for ${this.selectedSymbol}`
          );
          await this.loadPatterns();
        }
      } else {
        const error = await response.json();
        throw new Error(error.message || "Pattern detection failed");
      }
    } catch (error) {
      this.showError("Failed to detect patterns: " + error.message);
    } finally {
      this.hideLoading();
    }
  }

  togglePatternOverlay() {
    const overlay = document.getElementById("pattern-overlay");
    const btn = document.getElementById("toggle-pattern-overlay");

    if (overlay.classList.contains("d-none")) {
      overlay.classList.remove("d-none");
      btn.innerHTML = '<i class="bi bi-eye-slash"></i> Hide Patterns';
    } else {
      overlay.classList.add("d-none");
      btn.innerHTML = '<i class="bi bi-eye"></i> Show Patterns';
    }
  }

  toggleAnnotations() {
    this.showInfo("Annotations feature coming soon!");
  }

  toggleThesisPanel() {
    const panel = document.getElementById("thesis-panel");
    const toggle = document.getElementById("thesis-panel-toggle");

    if (panel.classList.contains("expanded")) {
      panel.classList.remove("expanded");
      toggle.classList.add("bi-chevron-up");
      toggle.classList.remove("bi-chevron-down");
    } else {
      panel.classList.add("expanded");
      toggle.classList.remove("bi-chevron-up");
      toggle.classList.add("bi-chevron-down");
    }
  }

  updateSymbolCount() {
    document.getElementById("symbol-count").textContent = this.symbols.length;
  }

  updatePatternCount() {
    // Pattern count is no longer displayed separately since we removed the patterns panel
    // Pattern information is now integrated into the symbol list
    console.log(`Total patterns: ${this.patterns.length}`);
  }

  startAutoRefresh() {
    // Refresh every 5 minutes
    this.refreshInterval = setInterval(() => {
      console.log("Auto-refreshing data...");
      this.loadSymbols();
      this.loadPatterns();
    }, 5 * 60 * 1000);
  }

  // UI helper methods
  showSymbolsLoading() {
    document.getElementById("symbols-list").innerHTML = `
            <div class="text-center p-4 text-muted">
                <div class="loading-spinner me-2"></div>
                Loading symbols...
            </div>
        `;
  }

  showLoading(message = "Loading...") {
    // TODO: Show loading indicator
    console.log(message);
  }

  hideLoading() {
    // TODO: Hide loading indicator
  }

  showSuccess(message) {
    this.showToast(message, "success");
  }

  showError(message) {
    this.showToast(message, "error");
  }

  showInfo(message) {
    this.showToast(message, "info");
  }

  showToast(message, type) {
    const container = document.getElementById("toast-container");
    const toastId = "toast-" + Date.now();

    const toast = document.createElement("div");
    toast.id = toastId;
    toast.className = `toast toast-${type}`;
    toast.innerHTML = `
            <div class="toast-header">
                <strong class="me-auto">${
                  type === "success"
                    ? "Success"
                    : type === "error"
                    ? "Error"
                    : "Info"
                }</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast"></button>
            </div>
            <div class="toast-body">${message}</div>
        `;

    container.appendChild(toast);
    const bsToast = new bootstrap.Toast(toast);
    bsToast.show();

    toast.addEventListener("hidden.bs.toast", () => {
      toast.remove();
    });
  }

  // Add symbol to watchlist
  async saveSymbol() {
    const symbolInput = document.getElementById("symbol-input");
    const symbol = symbolInput.value.trim().toUpperCase();

    if (!symbol) {
      this.showError("Please enter a symbol");
      return;
    }

    try {
      const response = await fetch("/api/symbols", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ symbol: symbol }),
      });

      if (response.ok) {
        this.showSuccess(
          `Symbol ${symbol} added to watchlist. Automatic pattern detection started.`
        );
        symbolInput.value = "";

        // Close modal
        const modal = bootstrap.Modal.getInstance(
          document.getElementById("add-symbol-modal")
        );
        if (modal) modal.hide();

        // Refresh symbols list and trigger pattern detection
        await this.loadSymbols();
        setTimeout(() => {
          this.loadPatterns();
        }, 3000);
      } else {
        const error = await response.json();
        throw new Error(error.message || "Failed to add symbol");
      }
    } catch (error) {
      this.showError("Failed to add symbol: " + error.message);
    }
  }

  async saveComponentUpdate() {
    const patternId = document.getElementById("update-pattern-id").value;
    const componentName = document.getElementById(
      "update-component-name"
    ).value;
    const isCompleted = document.getElementById("update-is-completed").checked;
    const confidenceLevel = document.getElementById("update-confidence").value;
    const evidence = document
      .getElementById("update-evidence")
      .value.split("\n")
      .filter((line) => line.trim());
    const notes = document.getElementById("update-notes").value;

    if (!patternId || !componentName) {
      this.showError("Missing pattern ID or component name");
      return;
    }

    try {
      const response = await fetch(
        `/api/head-shoulders/pattern/${patternId}/thesis/${componentName}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            is_completed: isCompleted,
            confidence_level: parseFloat(confidenceLevel),
            evidence: evidence,
            notes: notes,
          }),
        }
      );

      if (response.ok) {
        const result = await response.json();
        this.showSuccess(`Component "${componentName}" updated successfully`);

        // Close modal
        const modal = bootstrap.Modal.getInstance(
          document.getElementById("component-update-modal")
        );
        if (modal) modal.hide();

        // Refresh pattern details if this pattern is selected
        if (this.selectedPattern && this.selectedPattern.id == patternId) {
          this.loadThesisComponents(this.selectedPattern);
        }
      } else {
        const error = await response.json();
        throw new Error(error.message || "Failed to update component");
      }
    } catch (error) {
      this.showError("Failed to update component: " + error.message);
    }
  }
}

// Initialize when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  console.log("DOM loaded, initializing Pattern Watcher...");
  window.patternWatcher = new PatternWatcher();
});
