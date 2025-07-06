// Stock Watchlist JavaScript
console.log("[watchlist.js] Script loaded");

class WatchlistManager {
  constructor() {
    this.strategies = [];
    this.stocks = [];
    this.selectedStrategyId = null;
    this.currentSort = { field: 'symbol', direction: 'asc' };
    this.searchTerm = '';

    this.init();
  }

  async init() {
    console.log("Initializing Watchlist Manager...");

    try {
      this.setupEventListeners();
      await this.loadStrategies();
      await this.loadStocks();

      console.log("Watchlist Manager initialized successfully");
    } catch (error) {
      console.error("Failed to initialize Watchlist Manager:", error);
      this.showError("Failed to initialize watchlist: " + error.message);
    }
  }

  setupEventListeners() {
    // Add stock button
    document.getElementById("add-stock-btn").addEventListener("click", () => {
      this.openAddStockModal();
    });

    // Add strategy button
    document.getElementById("add-strategy-btn").addEventListener("click", () => {
      this.openAddStrategyModal();
    });

    // Search input
    document.getElementById("stock-search").addEventListener("input", (e) => {
      this.searchTerm = e.target.value.toLowerCase();
      this.filterAndDisplayStocks();
    });

    // Sort dropdown
    document.querySelectorAll("[data-sort]").forEach((item) => {
      item.addEventListener("click", (e) => {
        e.preventDefault();
        const field = e.target.dataset.sort;
        this.sortStocks(field);
      });
    });

    // Refresh prices button
    document.getElementById("refresh-prices-btn").addEventListener("click", () => {
      this.refreshPrices();
    });

    // Save stock button
    document.getElementById("save-stock-btn").addEventListener("click", () => {
      this.saveStock();
    });

    // Update stock button
    document.getElementById("update-stock-btn").addEventListener("click", () => {
      this.updateStock();
    });

    // Save strategy button
    document.getElementById("save-strategy-btn").addEventListener("click", () => {
      this.saveStrategy();
    });
  }

  // Strategies Management

  async loadStrategies() {
    try {
      const response = await fetch("/api/watchlist/strategies");
      if (!response.ok) {
        console.error("HTTP error loading strategies:", response.status, response.statusText);
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      console.log("Loaded strategies:", data);
      this.strategies = data.strategies || [];
      this.displayStrategies();
      this.populateStrategyDropdowns();
    } catch (error) {
      console.error("Failed to load strategies:", error);
      this.showError("Failed to load strategies: " + error.message);
    }
  }

  displayStrategies() {
    const container = document.getElementById("strategies-list");
    
    if (this.strategies.length === 0) {
      container.innerHTML = `
        <div class="text-center p-3 text-muted">
          <i class="bi bi-tags opacity-25"></i>
          <p class="mt-2 small">No strategies</p>
          <small>Click + to add one</small>
        </div>
      `;
      return;
    }

    // Add "All Strategies" option
    let html = `
      <div class="category-item ${this.selectedStrategyId === null ? 'active' : ''}"
           onclick="window.watchlist.selectStrategy(null)">
        <div class="d-flex align-items-center">
          <span class="category-color-dot" style="background-color: #6c757d;"></span>
          <span class="small">All Strategies</span>
        </div>
        <div class="small text-muted">${this.stocks.length} stocks</div>
      </div>
    `;

    // Add individual strategies
    this.strategies.forEach(strategy => {
      const stockCount = strategy.stocks ? strategy.stocks.length : 0;
      const isActive = this.selectedStrategyId === strategy.id;
      
      html += `
        <div class="category-item ${isActive ? 'active' : ''}"
             onclick="window.watchlist.selectStrategy(${strategy.id})">
          <div class="d-flex align-items-center justify-content-between">
            <div class="d-flex align-items-center">
              <span class="category-color-dot" style="background-color: ${strategy.color};"></span>
              <span class="small">${strategy.name}</span>
            </div>
            <div class="d-flex align-items-center">
              <span class="small text-muted me-2">${stockCount}</span>
              <button class="btn btn-sm text-danger p-0"
                      onclick="event.stopPropagation(); window.watchlist.deleteStrategy(${strategy.id})"
                      title="Delete strategy">
                <i class="bi bi-trash" style="font-size: 10px;"></i>
              </button>
            </div>
          </div>
          ${strategy.description ? `<div class="small text-muted mt-1">${strategy.description}</div>` : ''}
        </div>
      `;
    });

    container.innerHTML = html;
  }

  populateStrategyDropdowns() {
    const dropdowns = ['stock-strategy', 'edit-stock-strategy'];
    
    dropdowns.forEach(id => {
      const select = document.getElementById(id);
      if (!select) return;

      // Clear existing options except the first one
      select.innerHTML = '<option value="">No Strategy</option>';
      
      this.strategies.forEach(strategy => {
        const option = document.createElement('option');
        option.value = strategy.id;
        option.textContent = strategy.name;
        select.appendChild(option);
      });
    });
  }

  selectStrategy(strategyId) {
    this.selectedStrategyId = strategyId;
    this.displayStrategies();
    this.filterAndDisplayStocks();
    
    // Update title
    if (strategyId === null) {
      document.getElementById("watchlist-title").textContent = "All Stocks";
    } else {
      const strategy = this.strategies.find(s => s.id === strategyId);
      document.getElementById("watchlist-title").textContent = strategy ? strategy.name : "Unknown Strategy";
    }
  }

  // Stocks Management

  async loadStocks() {
    try {
      const response = await fetch("/api/watchlist/stocks");
      if (!response.ok) {
        console.error("HTTP error loading stocks:", response.status, response.statusText);
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      console.log("Loaded stocks:", data);
      this.stocks = data.stocks.map(stock => ({
        ...stock,
        strategies: stock.strategies || []
      })) || [];
      this.filterAndDisplayStocks();
    } catch (error) {
      console.error("Failed to load stocks:", error);
      this.showError("Failed to load stocks: " + error.message);
    }
  }

  filterAndDisplayStocks() {
      let filteredStocks = this.stocks;
      let selectedStrategyDetails = null;
  
      // Filter by strategy
      if (this.selectedStrategyId !== null) {
          const selectedStrategy = this.strategies.find(s => s.id === this.selectedStrategyId);
          if (selectedStrategy) {
              filteredStocks = selectedStrategy.stocks.map(stock => ({
                  ...stock,
                  strategies: this.strategies.filter(strategy => strategy.stocks.some(s => s.id === stock.id))
              }));
              selectedStrategyDetails = selectedStrategy;
          }
      }
  
      // Filter by search term
      if (this.searchTerm) {
          filteredStocks = filteredStocks.filter(stock =>
              stock.symbol.toLowerCase().includes(this.searchTerm) ||
              (stock.name && stock.name.toLowerCase().includes(this.searchTerm)) ||
              (stock.notes && stock.notes.toLowerCase().includes(this.searchTerm)) ||
              (stock.tags && stock.tags.toLowerCase().includes(this.searchTerm))
          );
      }
  
      // Sort stocks
      if (!Array.isArray(filteredStocks)) {
          filteredStocks = [];
      }
      filteredStocks.sort((a, b) => {
          const { field, direction } = this.currentSort;
          let aVal = a[field] || '';
          let bVal = b[field] || '';
  
          if (typeof aVal === 'string') {
              aVal = aVal.toLowerCase();
              bVal = bVal.toLowerCase();
          }
  
          if (direction === 'asc') {
              return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
          } else {
              return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
          }
      });
  
      this.displayStocks(filteredStocks, selectedStrategyDetails);
  }

  displayStocks(stocks) {
    const tbody = document.getElementById("stocks-tbody");
    
    if (stocks.length === 0) {
      tbody.innerHTML = `
        <tr>
          <td colspan="9" class="text-center p-4 text-muted">
            <i class="bi bi-inbox opacity-25 display-4"></i>
            <p class="mt-3">No stocks found</p>
            <button class="btn btn-primary btn-sm" onclick="window.watchlist.openAddStockModal()">
              <i class="bi bi-plus-circle"></i> Add Your First Stock
            </button>
          </td>
        </tr>
      `;
      return;
    }

    const html = stocks.map(stock => this.createStockRow(stock)).join('');
    tbody.innerHTML = html;
  }

  createStockRow(stock) {
    const changeClass = stock.change > 0 ? 'price-positive' : stock.change < 0 ? 'price-negative' : 'price-neutral';
    const changeIcon = stock.change > 0 ? 'bi-arrow-up' : stock.change < 0 ? 'bi-arrow-down' : 'bi-dash';
    
    console.log("Stock data:", stock);
    const strategyBadge = stock.strategies && stock.strategies.length > 0
      ? stock.strategies.map(strategy =>
          `<span class="category-badge" style="border-color: ${strategy.color}; color: ${strategy.color};">
            ${strategy.name}
          </span>`
        ).join(' ')
      : '<span class="text-muted small">No associated strategies</span>';

    const tags = stock.tags ? 
      stock.tags.split(',').map(tag => 
        `<span class="tag-pill">${tag.trim()}</span>`
      ).join('') : '';

    const addedDate = new Date(stock.added_at).toLocaleDateString();

    return `
      <tr onclick="window.watchlist.viewStockDetails(${stock.id})">
        <td>
          <strong>${stock.symbol}</strong>
        </td>
        <td>
          <div>${stock.name || stock.symbol}</div>
          ${tags ? `<div class="stock-tags mt-1">${tags}</div>` : ''}
        </td>
        <td>${strategyBadge}</td>
        <td>
          <span class="fw-bold">$${stock.price !== undefined ? stock.price.toFixed(2) : '-'}</span>
        </td>
        <td>
          <span class="fw-bold">${stock.ema_9 !== undefined ? stock.ema_9.toFixed(2) : '-'}</span>
        </td>
        <td>
          <span class="fw-bold">${stock.ema_50 !== undefined ? stock.ema_50.toFixed(2) : '-'}</span>
        </td>
        <td>
          <span class="fw-bold">${stock.ema_200 !== undefined ? stock.ema_200.toFixed(2) : '-'}</span>
        </td>
        <td class="${changeClass}">
          <i class="bi ${changeIcon} me-1"></i>
          ${stock.change_percent !== undefined ? stock.change_percent.toFixed(2) : '-'}%
          <div class="small">$${stock.change !== undefined ? stock.change.toFixed(2) : '-'}</div>
        </td>
        <td class="volume-formatted">
          ${this.formatVolume(stock.volume)}
        </td>
        <td>
          <div class="notes-preview">${stock.notes}</div>
        </td>
        <td class="small text-muted">
          ${addedDate}
        </td>
        <td>
          <div class="action-buttons">
            <button class="btn btn-outline-primary btn-sm" 
                    onclick="event.stopPropagation(); window.watchlist.editStock(${stock.id})"
                    title="Edit stock">
              <i class="bi bi-pencil"></i>
            </button>
            ${
              this.selectedStrategyId !== null
                ? `<button class="btn btn-outline-danger btn-sm"
                     onclick="event.stopPropagation(); window.watchlist.deleteStock(${stock.id})"
                     title="Remove from watchlist">
                     <i class="bi bi-trash"></i>
                   </button>`
                : ""
            }
          </div>
        </td>
      </tr>
    `;
  }

  formatVolume(volume) {
    if (volume >= 1000000) {
      return (volume / 1000000).toFixed(1) + 'M';
    } else if (volume >= 1000) {
      return (volume / 1000).toFixed(1) + 'K';
    }
    return volume.toString();
  }

  sortStocks(field) {
    if (this.currentSort.field === field) {
      this.currentSort.direction = this.currentSort.direction === 'asc' ? 'desc' : 'asc';
    } else {
      this.currentSort.field = field;
      this.currentSort.direction = 'asc';
    }
    
    this.filterAndDisplayStocks();
  }

  updateStockCount() {
    const count = this.selectedStrategyId === null ?
      this.stocks.length :
      this.stocks.filter(s => s.strategy_id === this.selectedStrategyId).length;
    
    document.getElementById("stock-count").textContent = count;
  }

  // Modal Operations

  openAddStockModal() {
    document.getElementById("add-stock-form").reset();
    const modalElement = document.getElementById("add-stock-modal");
    const modal = new bootstrap.Modal(modalElement);
    modalElement.addEventListener("hidden.bs.modal", () => {
        document.body.classList.remove("modal-open");
        const backdrop = document.querySelector(".modal-backdrop");
        if (backdrop) backdrop.remove();
    });
    modal.show();
  }

  openAddStrategyModal() {
    document.getElementById("add-strategy-form").reset();
    const modal = new bootstrap.Modal(document.getElementById("add-strategy-modal"));
    modal.show();
  }

  async saveStock() {
    const symbol = document.getElementById("stock-symbol").value.trim().toUpperCase();
    const name = document.getElementById("stock-name").value.trim();
    const strategyId = document.getElementById("stock-strategy").value || null;
    const tags = document.getElementById("stock-tags").value.trim();
    const notes = document.getElementById("stock-notes").value.trim();

    if (!symbol) {
      this.showError("Stock symbol is required");
      return;
    }

    try {
      const response = await fetch("/api/watchlist/stocks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          symbol,
          name,
          strategy_id: strategyId ? parseInt(strategyId) : null,
          tags,
          notes
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to add stock");
      }

      this.showSuccess(`${symbol} added to watchlist`);
      
      // Close modal
      const modal = bootstrap.Modal.getInstance(document.getElementById("add-stock-modal"));
      modal.hide();
      
      // Reload stocks
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to add stock: " + error.message);
    }
  }

  async saveStrategy() {
    const name = document.getElementById("strategy-name").value.trim();
    const description = document.getElementById("strategy-description").value.trim();
    const color = document.getElementById("strategy-color").value;

    if (!name) {
      this.showError("Strategy name is required");
      return;
    }

    try {
      const response = await fetch("/api/watchlist/strategies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description, color })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to create strategy");
      }

      this.showSuccess(`Strategy "${name}" created`);
      
      // Close modal
      const modal = bootstrap.Modal.getInstance(document.getElementById("add-strategy-modal"));
      modal.hide();
      
      // Reload strategies
      await this.loadStrategies();
      
    } catch (error) {
      this.showError("Failed to create strategy: " + error.message);
    }
  }

  async editStock(id) {
    const stock = this.stocks.find(s => s.id === id);
    if (!stock) return;

    // Populate edit form
    document.getElementById("edit-stock-id").value = stock.id;
    document.getElementById("edit-stock-symbol").value = stock.symbol;
    document.getElementById("edit-stock-name").value = stock.name || '';
    document.getElementById("edit-stock-strategy").value = stock.strategy_id || '';
    document.getElementById("edit-stock-tags").value = stock.tags || '';
    document.getElementById("edit-stock-notes").value = stock.notes || '';

    // Show modal
    const modal = new bootstrap.Modal(document.getElementById("edit-stock-modal"));
    modal.show();
  }

  async updateStock() {
    const id = document.getElementById("edit-stock-id").value;
    const name = document.getElementById("edit-stock-name").value.trim();
    const strategyId = document.getElementById("edit-stock-strategy").value || null;
    const tags = document.getElementById("edit-stock-tags").value.trim();
    const notes = document.getElementById("edit-stock-notes").value.trim();

    try {
      const response = await fetch(`/api/watchlist/stocks/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          strategy_id: strategyId ? parseInt(strategyId) : null,
          tags,
          notes
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to update stock");
      }

      this.showSuccess("Stock updated successfully");
      
      // Close modal
      const modal = bootstrap.Modal.getInstance(document.getElementById("edit-stock-modal"));
      modal.hide();
      
      // Reload stocks
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to update stock: " + error.message);
    }
  }

  async deleteStock(id) {
    if (this.selectedStrategyId === null) {
      this.showError("Stocks cannot be removed from the 'All Strategies' list.");
      return;
    }

    const stock = this.stocks.find(s => s.id === id);
    if (!stock) return;


    try {
      const response = await fetch(`/api/watchlist/strategies/${this.selectedStrategyId}/stocks/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to delete stock");
      }

      this.showSuccess(`${stock.symbol} removed from watchlist`);

      // Remove stock from associated strategies
      const strategy = this.strategies.find(s => s.id === this.selectedStrategyId);
      if (strategy) {
        strategy.stocks = strategy.stocks.filter(s => s.id !== id);
      }

      await this.loadStrategies();
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to delete stock: " + error.message);
    }
  }

  async deleteStrategy(id) {
    const strategy = this.strategies.find(s => s.id === id);
    if (!strategy) return;

    // Store strategy data for potential restoration
    const strategyData = { ...strategy };

    try {
      const response = await fetch(`/api/watchlist/strategies/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to delete strategy");
      }

      // Show success message with undo option
      this.showUndoableSuccess(`Strategy "${strategy.name}" deleted`, () => {
        this.restoreStrategy(strategyData);
      });
      
      // If the deleted strategy was selected, select "All Strategies"
      if (this.selectedStrategyId === id) {
        this.selectedStrategyId = null;
      }
      
      await this.loadStrategies();
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to delete strategy: " + error.message);
    }
  }

  async restoreStrategy(strategyData) {
    try {
      const response = await fetch("/api/watchlist/strategies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: strategyData.name,
          description: strategyData.description,
          color: strategyData.color
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to restore strategy");
      }

      this.showSuccess(`Strategy "${strategyData.name}" restored`);
      await this.loadStrategies();
      
    } catch (error) {
      this.showError("Failed to restore strategy: " + error.message);
    }
  }

  async refreshPrices() {
    try {
      // Show loading spinner or disable button
      const btn = document.getElementById("refresh-prices-btn");
      btn.disabled = true;
      btn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Refreshing...';

      const response = await fetch("/api/watchlist/refresh", { method: "POST" });
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to refresh prices");
      }
      this.showSuccess("Prices and EMAs refreshed");
      await this.loadStocks();
    } catch (error) {
      this.showError("Failed to refresh prices: " + error.message);
    } finally {
      // Restore button state
      const btn = document.getElementById("refresh-prices-btn");
      btn.disabled = false;
      btn.innerHTML = '<i class="bi bi-arrow-repeat"></i> Refresh Prices';
    }
  }

  viewStockDetails(id) {
        // This could open a detailed view or redirect to a stock analysis page
        const stock = this.stocks.find(s => s.id === id);
        if (!stock) return;
        // Implement modal or redirect logic here
        this.showInfo(`Details for ${stock.symbol} coming soon!`);
    }

  // --- User Feedback Helpers ---

  showError(message) {
    this.showBanner(message, 'danger');
  }

  showSuccess(message) {
    this.showBanner(message, 'success');
  }

  showInfo(message) {
    this.showBanner(message, 'info');
  }

  showUndoableSuccess(message, undoCallback) {
    this.showBanner(message + ' <button class="btn btn-link btn-sm p-0 ms-2" id="undo-btn">Undo</button>', 'success', true, undoCallback);
  }

  showBanner(message, type = 'info', html = false, undoCallback = null) {
    let banner = document.getElementById('watchlist-banner');
    if (!banner) {
      banner = document.createElement('div');
      banner.id = 'watchlist-banner';
      banner.className = 'alert alert-' + type + ' position-fixed top-0 start-50 translate-middle-x mt-3 shadow';
      banner.style.zIndex = 2000;
      banner.style.minWidth = '300px';
      banner.style.maxWidth = '90vw';
      banner.style.textAlign = 'center';
      document.body.appendChild(banner);
    }
    banner.className = 'alert alert-' + type + ' position-fixed top-0 start-50 translate-middle-x mt-3 shadow';
    if (html) {
      banner.innerHTML = message;
    } else {
      banner.textContent = message;
    }
    banner.style.display = 'block';
    if (undoCallback) {
      setTimeout(() => {
        const undoBtn = document.getElementById('undo-btn');
        if (undoBtn) {
          undoBtn.onclick = () => {
            banner.style.display = 'none';
            undoCallback();
          };
        }
      }, 0);
    }
    setTimeout(() => {
      banner.style.display = 'none';
    }, 3500);
  }
}

// Ensure WatchlistManager is instantiated and attached to window
window.addEventListener('DOMContentLoaded', () => {
  console.log("[watchlist.js] DOMContentLoaded");
  if (!window.watchlist) {
    window.watchlist = new WatchlistManager();
  }
});
