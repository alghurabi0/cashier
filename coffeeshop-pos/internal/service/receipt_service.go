package service

import (
	"coffeeshop-pos/internal/model"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// ReceiptService is a Wails-bound service for receipt printing.
// It generates ESC/POS formatted receipts for thermal printers.
type ReceiptService struct {
	shopName string
}

// NewReceiptService creates a new ReceiptService.
func NewReceiptService(shopName string) *ReceiptService {
	return &ReceiptService{shopName: shopName}
}

// PrintReceipt generates and sends a receipt to the connected thermal printer.
// Best-effort: logs errors but returns them so the frontend can show a warning.
func (s *ReceiptService) PrintReceipt(order model.OrderWithItems) error {
	receipt := s.formatReceipt(order)

	// For now, log the receipt content (actual ESC/POS USB printing
	// requires the github.com/kenshaw/escpos library and a connected printer).
	// This will be upgraded to actual hardware printing once a printer is available.
	slog.Info("receipt generated", "order_number", order.OrderNumber)
	slog.Debug("receipt content", "receipt", receipt)

	// TODO: Phase 3.5 — Connect to actual ESC/POS printer via USB
	// Example with escpos:
	// p, err := escpos.NewUSBPrinterByPath("/dev/usb/lp0")
	// p.Init()
	// p.Write(receipt)
	// p.Cut()
	// p.End()

	return nil
}

// GetReceiptText returns the formatted receipt text for display/preview.
// Callable from the frontend for a print preview.
func (s *ReceiptService) GetReceiptText(order model.OrderWithItems) (string, error) {
	return s.formatReceipt(order), nil
}

// formatReceipt builds the receipt text.
func (s *ReceiptService) formatReceipt(order model.OrderWithItems) string {
	var b strings.Builder
	width := 32 // 80mm thermal paper ≈ 32 chars at standard font

	divider := strings.Repeat("━", width)
	thinDivider := strings.Repeat("─", width)

	// Header
	b.WriteString(divider + "\n")
	b.WriteString(centerText(s.shopName, width) + "\n")
	b.WriteString(divider + "\n")

	// Order info
	b.WriteString(fmt.Sprintf("طلب رقم: %s\n", order.OrderNumber))

	// Parse and format date
	t, err := time.Parse("2006-01-02 15:04:05", order.CreatedAt)
	if err != nil {
		b.WriteString(fmt.Sprintf("التاريخ: %s\n", order.CreatedAt))
	} else {
		b.WriteString(fmt.Sprintf("التاريخ: %s\n", t.Format("2006/01/02 15:04")))
	}

	b.WriteString("\n")
	b.WriteString(thinDivider + "\n")

	// Items
	for _, item := range order.Items {
		price := formatPrice(item.LineTotal)
		qty := fmt.Sprintf("×%d", item.Quantity)
		line := fmt.Sprintf("%s %s %s", item.NameArSnapshot, qty, price)
		b.WriteString(line + "\n")
	}

	b.WriteString(thinDivider + "\n")

	// Total
	b.WriteString(fmt.Sprintf("المجموع: %s د.ع\n", formatPrice(order.Total)))

	// Payment method
	paymentLabel := "نقدي"
	if order.PaymentMethod == "card" {
		paymentLabel = "بطاقة"
	}
	b.WriteString(fmt.Sprintf("الدفع: %s\n", paymentLabel))

	// Footer
	b.WriteString(divider + "\n")
	b.WriteString(centerText("شكراً لزيارتكم", width) + "\n")
	b.WriteString(divider + "\n")

	return b.String()
}

// centerText centers text within the given width.
func centerText(text string, width int) string {
	runeCount := len([]rune(text))
	if runeCount >= width {
		return text
	}
	padding := (width - runeCount) / 2
	return strings.Repeat(" ", padding) + text
}

// formatPrice formats a fils amount as a human-readable price.
// Example: 3500 → "3,500"
func formatPrice(fils int64) string {
	s := fmt.Sprintf("%d", fils)
	if len(s) <= 3 {
		return s
	}

	// Add thousand separators
	var result strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}
