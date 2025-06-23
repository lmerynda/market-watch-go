// Stock Watchlist JavaScript
console.log("[watchlist.js] Script loaded");

class WatchlistManager {
  constructor() {
    this.categories = [];
    this.stocks = [];
    this.selectedCategoryId = null;
    this.currentSort = { field: 'symbol', direction: 'asc' };
    this.searchTerm = '';

    this.init();
  }

  async init() {
    console.log("Initializing Watchlist Manager...");

    try {
      this.setupEventListeners();
      await this.loadCategories();
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

    // Add category button
    document.getElementById("add-category-btn").addEventListener("click", () => {
      this.openAddCategoryModal();
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

    // Save category button
    document.getElementById("save-category-btn").addEventListener("click", () => {
      this.saveCategory();
    });
  }

  // Categories Management

  async loadCategories() {
    try {
      const response = await fetch("/api/watchlist/categories");
      if (!response.ok) {
        console.error("HTTP error loading categories:", response.status, response.statusText);
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      console.log("Loaded categories:", data);
      this.categories = data.categories || [];
      this.displayCategories();
      this.populateCategoryDropdowns();
    } catch (error) {
      console.error("Failed to load categories:", error);
      this.showError("Failed to load categories: " + error.message);
    }
  }

  displayCategories() {
    const container = document.getElementById("categories-list");
    
    if (this.categories.length === 0) {
      container.innerHTML = `
        <div class="text-center p-3 text-muted">
          <i class="bi bi-tags opacity-25"></i>
          <p class="mt-2 small">No categories</p>
          <small>Click + to add one</small>
        </div>
      `;
      return;
    }

    // Add "All Categories" option
    let html = `
      <div class="category-item ${this.selectedCategoryId === null ? 'active' : ''}" 
           onclick="window.watchlist.selectCategory(null)">
        <div class="d-flex align-items-center">
          <span class="category-color-dot" style="background-color: #6c757d;"></span>
          <span class="small">All Categories</span>
        </div>
        <div class="small text-muted">${this.stocks.length} stocks</div>
      </div>
    `;

    // Add individual categories
    this.categories.forEach(category => {
      const stockCount = this.stocks.filter(s => s.category_id === category.id).length;
      const isActive = this.selectedCategoryId === category.id;
      
      html += `
        <div class="category-item ${isActive ? 'active' : ''}" 
             onclick="window.watchlist.selectCategory(${category.id})">
          <div class="d-flex align-items-center justify-content-between">
            <div class="d-flex align-items-center">
              <span class="category-color-dot" style="background-color: ${category.color};"></span>
              <span class="small">${category.name}</span>
            </div>
            <div class="d-flex align-items-center">
              <span class="small text-muted me-2">${stockCount}</span>
              <button class="btn btn-sm text-danger p-0" 
                      onclick="event.stopPropagation(); window.watchlist.deleteCategory(${category.id})"
                      title="Delete category">
                <i class="bi bi-trash" style="font-size: 10px;"></i>
              </button>
            </div>
          </div>
          ${category.description ? `<div class="small text-muted mt-1">${category.description}</div>` : ''}
        </div>
      `;
    });

    container.innerHTML = html;
  }

  populateCategoryDropdowns() {
    const dropdowns = ['stock-category', 'edit-stock-category'];
    
    dropdowns.forEach(id => {
      const select = document.getElementById(id);
      if (!select) return;

      // Clear existing options except the first one
      select.innerHTML = '<option value="">No Category</option>';
      
      this.categories.forEach(category => {
        const option = document.createElement('option');
        option.value = category.id;
        option.textContent = category.name;
        select.appendChild(option);
      });
    });
  }

  selectCategory(categoryId) {
    this.selectedCategoryId = categoryId;
    this.displayCategories();
    this.filterAndDisplayStocks();
    
    // Update title
    if (categoryId === null) {
      document.getElementById("watchlist-title").textContent = "All Stocks";
    } else {
      const category = this.categories.find(c => c.id === categoryId);
      document.getElementById("watchlist-title").textContent = category ? category.name : "Unknown Category";
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
      this.stocks = data.stocks || [];
      this.filterAndDisplayStocks();
    } catch (error) {
      console.error("Failed to load stocks:", error);
      this.showError("Failed to load stocks: " + error.message);
    }
  }

  filterAndDisplayStocks() {
    let filteredStocks = this.stocks;

    // Filter by category
    if (this.selectedCategoryId !== null) {
      filteredStocks = filteredStocks.filter(stock => stock.category_id === this.selectedCategoryId);
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

    this.displayStocks(filteredStocks);
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
    
    const categoryBadge = stock.category_name ? 
      `<span class="category-badge" style="border-color: ${stock.category_color}; color: ${stock.category_color};">
        ${stock.category_name}
      </span>` : 
      '<span class="text-muted small">None</span>';

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
        <td>${categoryBadge}</td>
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
            <button class="btn btn-outline-danger btn-sm" 
                    onclick="event.stopPropagation(); window.watchlist.deleteStock(${stock.id})"
                    title="Remove from watchlist">
              <i class="bi bi-trash"></i>
            </button>
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
    const count = this.selectedCategoryId === null ? 
      this.stocks.length : 
      this.stocks.filter(s => s.category_id === this.selectedCategoryId).length;
    
    document.getElementById("stock-count").textContent = count;
  }

  // Modal Operations

  openAddStockModal() {
    document.getElementById("add-stock-form").reset();
    const modal = new bootstrap.Modal(document.getElementById("add-stock-modal"));
    modal.show();
  }

  openAddCategoryModal() {
    document.getElementById("add-category-form").reset();
    const modal = new bootstrap.Modal(document.getElementById("add-category-modal"));
    modal.show();
  }

  async saveStock() {
    const symbol = document.getElementById("stock-symbol").value.trim().toUpperCase();
    const name = document.getElementById("stock-name").value.trim();
    const categoryId = document.getElementById("stock-category").value || null;
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
          category_id: categoryId ? parseInt(categoryId) : null,
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

  async saveCategory() {
    const name = document.getElementById("category-name").value.trim();
    const description = document.getElementById("category-description").value.trim();
    const color = document.getElementById("category-color").value;

    if (!name) {
      this.showError("Category name is required");
      return;
    }

    try {
      const response = await fetch("/api/watchlist/categories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description, color })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to create category");
      }

      this.showSuccess(`Category "${name}" created`);
      
      // Close modal
      const modal = bootstrap.Modal.getInstance(document.getElementById("add-category-modal"));
      modal.hide();
      
      // Reload categories
      await this.loadCategories();
      
    } catch (error) {
      this.showError("Failed to create category: " + error.message);
    }
  }

  async editStock(id) {
    const stock = this.stocks.find(s => s.id === id);
    if (!stock) return;

    // Populate edit form
    document.getElementById("edit-stock-id").value = stock.id;
    document.getElementById("edit-stock-symbol").value = stock.symbol;
    document.getElementById("edit-stock-name").value = stock.name || '';
    document.getElementById("edit-stock-category").value = stock.category_id || '';
    document.getElementById("edit-stock-tags").value = stock.tags || '';
    document.getElementById("edit-stock-notes").value = stock.notes || '';

    // Show modal
    const modal = new bootstrap.Modal(document.getElementById("edit-stock-modal"));
    modal.show();
  }

  async updateStock() {
    const id = document.getElementById("edit-stock-id").value;
    const name = document.getElementById("edit-stock-name").value.trim();
    const categoryId = document.getElementById("edit-stock-category").value || null;
    const tags = document.getElementById("edit-stock-tags").value.trim();
    const notes = document.getElementById("edit-stock-notes").value.trim();

    try {
      const response = await fetch(`/api/watchlist/stocks/${id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          category_id: categoryId ? parseInt(categoryId) : null,
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
    const stock = this.stocks.find(s => s.id === id);
    if (!stock) return;

    if (!confirm(`Remove ${stock.symbol} from watchlist?`)) return;

    try {
      const response = await fetch(`/api/watchlist/stocks/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to delete stock");
      }

      this.showSuccess(`${stock.symbol} removed from watchlist`);
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to delete stock: " + error.message);
    }
  }

  async deleteCategory(id) {
    const category = this.categories.find(c => c.id === id);
    if (!category) return;

    // Store category data for potential restoration
    const categoryData = { ...category };

    try {
      const response = await fetch(`/api/watchlist/categories/${id}`, {
        method: "DELETE"
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to delete category");
      }

      // Show success message with undo option
      this.showUndoableSuccess(`Category "${category.name}" deleted`, () => {
        this.restoreCategory(categoryData);
      });
      
      // If the deleted category was selected, select "All Categories"
      if (this.selectedCategoryId === id) {
        this.selectedCategoryId = null;
      }
      
      await this.loadCategories();
      await this.loadStocks();
      
    } catch (error) {
      this.showError("Failed to delete category: " + error.message);
    }
  }

  async restoreCategory(categoryData) {
    try {
      const response = await fetch("/api/watchlist/categories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: categoryData.name,
          description: categoryData.description,
          color: categoryData.color
        })
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.details || "Failed to restore category");
      }

      this.showSuccess(`Category "${categoryData.name}" restored`);
      await this.loadCategories();
      
    } catch (error) {
      this.showError("Failed to restore category: " + error.message);
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
