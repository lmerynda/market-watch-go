// Pattern Watcher - Multi-Pattern Analysis Interface
class PatternWatcher {
  constructor() {
    this.symbols = [];
    this.patterns = [];
    this.selectedSymbol = null;
    this.selectedPattern = null;
    this.currentFilter = "all";
    this.currentWatchlistFilter = "all";
    this.currentTimeframe = "60"; // Default to 1 hour
    this.tradingViewWidget = null;
    this.refreshInterval = null;
    this.isLoading = false;

    this.init();
  }

  async init() {
    console.log("Initializing Pattern Watcher...");

    try {
      this.setupEventListeners();
      this.setupThesisPanel();
      
      // Load symbols first
      console.log("Loading symbols during initialization...");
      await this.loadSymbols();
      
      // Load patterns
      console.log("Loading patterns during initialization...");
      await this.loadPatterns();
      
      // Start auto-refresh
      this.startAutoRefresh();

      console.log("Pattern Watcher initialized successfully");
    } catch (error) {
      console.error("Failed to initialize Pattern Watcher:", error);
      this.showError("Failed to initialize Pattern Watcher: " + error.message);
      
      // Ensure UI is in a proper state even if initialization fails
      this.displayNoSymbols();
    }
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

    // Timeframe controls
    document.querySelectorAll('input[name="timeframe"]').forEach((radio) => {
      radio.addEventListener("change", (e) => {
        this.currentTimeframe = e.target.value;
        if (this.selectedSymbol) {
          this.loadTradingViewChart(this.selectedSymbol);
        }
      });
    });

    // Add symbol
    document.getElementById("add-symbol-btn").addEventListener("click", () => {
      this.quickAddSymbol();
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
      .addEventListener("click", async () => {
        console.log("Refresh Data button clicked");
        try {
          // Force reload of both symbols and patterns
          this.isLoading = false; // Reset loading flag in case it's stuck
          await this.loadSymbols();
          await this.loadPatterns();
          this.showSuccess("Data refreshed successfully");
        } catch (error) {
          console.error("Failed to refresh data:", error);
          this.showError("Failed to refresh data: " + error.message);
        }
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
  }

  async loadSymbols() {
    if (this.isLoading) {
      console.log("Already loading symbols, skipping...");
      return;
    }

    this.isLoading = true;
    this.showSymbolsLoading();

    try {
      console.log("Loading symbols...");

      // Get watched symbols from the existing API
      const response = await fetch("/api/symbols", {
        method: "GET",
        headers: {
          "Accept": "application/json",
          "Cache-Control": "no-cache"
        }
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      console.log("Symbols API response:", data);
      console.log(
        `Loaded ${data && data.symbols ? data.symbols.length : 0} symbols`
      );

      this.symbols = data && data.symbols ? data.symbols : [];
      
      // Always call displaySymbols, which will handle empty state appropriately
      this.displaySymbols();
      this.updateSymbolCount();
      
      console.log("Symbols loaded and displayed successfully");
    } catch (error) {
      console.error("Failed to load symbols:", error);
      this.showError("Failed to load symbols: " + error.message);
      this.symbols = [];
      this.displayNoSymbols();
    } finally {
      this.isLoading = false;
      console.log("loadSymbols completed, isLoading:", this.isLoading);
    }
  }

  async loadPatterns() {
    try {
      console.log("Loading patterns...");

      // Use unified patterns endpoint
      const response = await fetch("/api/patterns");

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      console.log(`Loaded patterns:`, data);

      // Extract patterns from the unified API response
      let allPatterns = [];
      if (data.patterns) {
        // Combine head_shoulders and falling_wedge patterns
        if (data.patterns.head_shoulders) {
          allPatterns = allPatterns.concat(data.patterns.head_shoulders.map(p => ({...p, pattern_type: 'head_shoulders'})));
        }
        if (data.patterns.falling_wedge) {
          allPatterns = allPatterns.concat(data.patterns.falling_wedge.map(p => ({...p, pattern_type: 'falling_wedge'})));
        }
      }

      this.patterns = allPatterns;
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

  async scanAllPatterns() {
    try {
      this.showLoading("Scanning all symbols for patterns...");

      // Use unified patterns scan endpoint
      const response = await fetch("/api/patterns/scan", {
        method: "POST",
      });

      if (response.ok) {
        const result = await response.json();
        this.showSuccess(
          `Pattern scan completed. ${
            result.message || "Multiple patterns updated."
          }`
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
        `/api/patterns/scan/${this.selectedSymbol}`,
        {
          method: "POST",
        }
      );

      if (response.ok) {
        const result = await response.json();

        if (result.patterns_found > 0) {
          const patternTypes = [];
          if (result.head_shoulders_pattern) patternTypes.push("Head & Shoulders");
          if (result.falling_wedge_pattern) patternTypes.push("Falling Wedge");
          
          this.showSuccess(`Pattern(s) detected for ${this.selectedSymbol}: ${patternTypes.join(", ")}`);
          await this.loadPatterns();
        } else {
          this.showInfo(
            `No patterns found for ${this.selectedSymbol}.`
          );
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

    // Add click handlers for symbol selection
    container.querySelectorAll(".symbol-item").forEach((item) => {
      item.addEventListener("click", (e) => {
        // Don't select symbol if clicking on remove button
        if (e.target.closest('.remove-symbol-btn')) {
          return;
        }
        const symbol = item.dataset.symbol;
        this.selectSymbol(symbol);
      });
    });

    // Add click handlers for remove buttons
    container.querySelectorAll(".remove-symbol-btn").forEach((btn) => {
      btn.addEventListener("click", (e) => {
        e.stopPropagation(); // Prevent symbol selection
        const symbol = btn.dataset.symbol;
        this.removeSymbol(symbol);
      });
    });
  }

  displayNoSymbols() {
    const container = document.getElementById("symbols-list");
    container.innerHTML = `
            <div class="text-center p-4 text-muted">
                <i class="bi bi-list-ul display-4 opacity-25"></i>
                <p class="mt-3">No symbols in watchlist</p>
                <p class="small text-muted">Use the input field above to add symbols or click "Refresh Data" to reload from server</p>
            </div>
        `;
  }

  createSymbolListItem(symbolData) {
    // Handle both string symbols and WatchedSymbol objects from API
    const symbol =
      typeof symbolData === "string"
        ? symbolData
        : symbolData.symbol || symbolData.Symbol || "";
    const symbolName =
      typeof symbolData === "object"
        ? symbolData.name || symbolData.Name || ""
        : "";

    if (!symbol) {
      console.warn("Invalid symbol data:", symbolData);
      return "";
    }

    const symbolPatterns = this.patterns.filter((p) => p.symbol === symbol);
    const patternBadges = this.createPatternBadges(symbolPatterns);

    return `
            <div class="symbol-item" data-symbol="${symbol}">
                <div class="d-flex justify-content-between align-items-start">
                    <div class="flex-grow-1">
                        <div class="d-flex justify-content-between align-items-start">
                            <div class="symbol-name">${symbol}</div>
                            <button class="btn btn-sm text-danger p-0 ms-2 remove-symbol-btn"
                                    data-symbol="${symbol}"
                                    title="Remove ${symbol} from watchlist"
                                    style="line-height: 1; font-size: 12px;">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
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

    const badges = patterns
      .map(
        (pattern) =>
          `<span class="pattern-badge falling-wedge">Falling Wedge</span>`
      )
      .join("");

    return `<div class="pattern-badges">${badges}</div>`;
  }

  displayPatterns() {
    // Patterns are now integrated into the symbol list display
    console.log(`Loaded ${this.patterns.length} patterns`);
  }

  displayNoPatterns() {
    console.log("No patterns to display");
  }

  getFilteredSymbols() {
    const searchTerm = document
      .getElementById("symbol-search")
      .value.toLowerCase();

    return this.symbols.filter((symbolData) => {
      const symbol =
        typeof symbolData === "string"
          ? symbolData
          : symbolData.symbol || symbolData.Symbol || "";

      if (!symbol) {
        console.warn("Invalid symbol data in filter:", symbolData);
        return false;
      }

      // Check search term match
      const matchesSearch = symbol.toLowerCase().includes(searchTerm);

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

  filterSymbols(searchTerm = "") {
    this.displaySymbols();
  }

  filterPatterns() {
    this.displayPatterns();
  }

  getSymbolStatus(patterns) {
    if (patterns.length === 0) return "no-patterns";
    if (patterns.some((p) => !p.is_complete)) return "patterns-found";
    return "patterns-found";
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

      // Load chart for selected symbol
      this.loadTradingViewChart(symbol);

      this.selectedSymbol = symbol;
    } catch (error) {
      console.error("Failed to select symbol:", error);
      this.showError("Failed to select symbol: " + error.message);
    }
  }

  loadTradingViewChart(symbol) {
    console.log(`Loading TradingView chart for ${symbol}`);

    const container = document.getElementById("tradingview-widget");
    if (!container) {
      console.error("TradingView widget container not found");
      return;
    }

    try {
      // Clear existing widget
      if (this.tradingViewWidget) {
        this.tradingViewWidget.remove();
      }

      // Create new TradingView widget
      this.tradingViewWidget = new TradingView.widget({
        width: "100%",
        height: "100%",
        symbol: symbol,
        interval: this.currentTimeframe,
        timezone: "Etc/UTC",
        theme: "dark",
        style: "1",
        locale: "en",
        toolbar_bg: "#f1f3f6",
        enable_publishing: false,
        hide_top_toolbar: false,
        hide_legend: false,
        save_image: false,
        container_id: "tradingview-widget",
        studies: [
          "Volume@tv-basicstudies",
          "RSI@tv-basicstudies",
          "MACD@tv-basicstudies"
        ],
        onChartReady: () => {
          console.log(
            `TradingView widget loaded successfully for ${symbol} on ${this.currentTimeframe} timeframe`
          );
        }
      });
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

  async removeSymbol(symbol) {
    if (!symbol) return;

    try {
      this.showLoading(`Removing ${symbol} from watchlist...`);

      const response = await fetch(`/api/symbols/${symbol}`, {
        method: "DELETE",
      });

      if (response.ok) {
        this.showSuccess(`Symbol ${symbol} removed from watchlist`);
        
        // If the removed symbol was selected, clear the selection
        if (this.selectedSymbol === symbol) {
          this.selectedSymbol = null;
          document.getElementById("selected-symbol").textContent = "Select a symbol to view chart";
          document.getElementById("pattern-info").textContent = "No symbol selected";
          document.getElementById("detect-pattern-btn").disabled = true;
        }

        // Refresh the symbols list
        await this.loadSymbols();
        
        // Also refresh patterns to remove any patterns for this symbol
        await this.loadPatterns();
      } else {
        const error = await response.json();
        throw new Error(error.message || "Failed to remove symbol");
      }
    } catch (error) {
      this.showError("Failed to remove symbol: " + error.message);
    } finally {
      this.hideLoading();
    }
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

  setupThesisPanel() {
    const panel = document.getElementById("thesis-panel");
    const toggle = document.getElementById("thesis-panel-toggle");

    // Initially collapsed
    panel.classList.remove("expanded");
    toggle.classList.add("bi-chevron-up");
    toggle.classList.remove("bi-chevron-down");
  }

  updateSymbolCount() {
    document.getElementById("symbol-count").textContent = this.symbols.length;
  }

  updatePatternCount() {
    // Pattern count is no longer displayed separately since we removed the patterns panel
    console.log(`Total patterns: ${this.patterns.length}`);
  }

  startAutoRefresh() {
    // Refresh every 5 minutes
    this.refreshInterval = setInterval(() => {
      console.log("Auto-refreshing data...");
      this.loadSymbols();
      this.loadPatterns();
    }, 5 * 60 * 1000);

    // Also schedule a retry in case initial load failed
    setTimeout(() => {
      if (this.symbols.length === 0) {
        console.log("No symbols loaded after 5 seconds, retrying...");
        this.loadSymbols();
      }
    }, 5000);

    // And another retry after 15 seconds in case the server is still starting up
    setTimeout(() => {
      if (this.symbols.length === 0) {
        console.log("No symbols loaded after 15 seconds, retrying...");
        this.loadSymbols();
      }
    }, 15000);
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
}

// Initialize when DOM is loaded
document.addEventListener("DOMContentLoaded", () => {
  console.log("DOM loaded, initializing Pattern Watcher...");
  window.patternWatcher = new PatternWatcher();
});
