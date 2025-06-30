package database

import (
	"strings"
)

// RefreshStock updates stock price and EMA data
func (db *Database) RefreshStock(symbol string, price, ema9, ema50, ema200 float64) error {
	query := `UPDATE stocks SET price = ?, ema_9 = ?, ema_50 = ?, ema_200 = ?, updated_at = CURRENT_TIMESTAMP WHERE symbol = ?`
	_, err := db.conn.Exec(query, price, ema9, ema50, ema200, strings.ToUpper(symbol))
	return err
}

// DeleteStock deletes a stock from the stocks table
func (db *Database) DeleteStock(stockID int) error {
	// First remove all strategy associations
	if err := db.RemoveAllStockStrategies(stockID); err != nil {
		return err
	}

	// Then delete the stock
	query := `DELETE FROM stocks WHERE id = ?`
	_, err := db.conn.Exec(query, stockID)
	return err
}
