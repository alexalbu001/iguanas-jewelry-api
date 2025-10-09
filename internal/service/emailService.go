package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService interface {
	SendOrderConfirmation(ctx context.Context, orderSummary OrderSummary) error
	SendWelcome(ctx context.Context, userName, userEmail string) error
	SendAdminOrderNotification(ctx context.Context, orderSummary OrderSummary, adminEmail string) error
}

type SendgridEmailService struct {
	client    *sendgrid.Client
	fromEmail string
	fromName  string
}

func NewSendgridEmailService(client *sendgrid.Client, fromEmail, fromName string) *SendgridEmailService {
	return &SendgridEmailService{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (s *SendgridEmailService) SendOrderConfirmation(ctx context.Context, orderSummary OrderSummary) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail(orderSummary.ShippingName, orderSummary.ShippingAddress.Email)
	subject := "Iguanas-Jewellery Order Confirmation"
	plainTextContent := fmt.Sprintf(`
Thank you for your order!

Order #: %s
Total: $%.2f
Order Date: %s

Items Ordered:
%s

Shipping Address:
%s
%s %s %s, %s

We'll send you a shipping confirmation once your order is on its way.

Thank you for choosing Iguanas jewellery!
`,
		orderSummary.ID,
		orderSummary.Total,
		orderSummary.CreatedDate.Format("January 2, 2006"),
		formatOrderItems(orderSummary.OrderItems),
		// Format the address
		orderSummary.ShippingAddress.AddressLine1,
		orderSummary.ShippingAddress.City, orderSummary.ShippingAddress.State,
		orderSummary.ShippingAddress.PostalCode, orderSummary.ShippingAddress.Country)

	htmlContent := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333; line-height: 1.6;">
			<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
				<h1 style="color: #2c5aa0;">Thank you for your order!</h1>
				
				<div style="background: #f8f9fa; padding: 20px; border-left: 4px solid #2c5aa0; margin: 20px 0;">
					<h3>Order #%s</h3>
					<p><strong>Total:</strong> $%.2f</p>
					<p><strong>Order Date:</strong> %s</p>
				</div>
				
				<h3>Items Ordered:</h3>
				%s
				
				<h3>Shipping Address:</h3>
				<div style="background: #f8f9fa; padding: 15px;">
					<p>%s<br>%s<br>%s, %s %s<br>%s</p>
				</div>
				
				<p style="margin-top: 30px;">We'll send you a shipping confirmation once your order is on its way.</p>
				<p><strong>Thank you for choosing Iguanas jewellery!</strong></p>
			</div>
		</body>
		</html>`,
		orderSummary.ID, orderSummary.Total, orderSummary.CreatedDate.Format("January 2, 2006"),
		formatOrderItemsHTML(orderSummary.OrderItems), // Only helper function you need
		orderSummary.ShippingName, orderSummary.ShippingAddress.AddressLine1,
		orderSummary.ShippingAddress.City, orderSummary.ShippingAddress.State,
		orderSummary.ShippingAddress.PostalCode, orderSummary.ShippingAddress.Country)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func formatOrderItems(items []OrderItemSummary) string {
	var result strings.Builder

	for _, item := range items {
		result.WriteString(fmt.Sprintf("- %s (Qty: %d) - $%.2f each = $%.2f\n",
			item.ProductName,
			item.Quantity,
			item.Price,
			item.Subtotal))
	}

	return result.String()
}

func formatOrderItemsHTML(items []OrderItemSummary) string {
	var result strings.Builder

	// Table header
	result.WriteString(`
		<table style="width: 100%; border-collapse: collapse; margin: 10px 0;">
			<thead>
				<tr style="background-color: #f8f9fa;">
					<th style="padding: 12px; text-align: left; border-bottom: 2px solid #dee2e6;">Product</th>
					<th style="padding: 12px; text-align: center; border-bottom: 2px solid #dee2e6;">Quantity</th>
					<th style="padding: 12px; text-align: right; border-bottom: 2px solid #dee2e6;">Price</th>
					<th style="padding: 12px; text-align: right; border-bottom: 2px solid #dee2e6;">Subtotal</th>
				</tr>
			</thead>
			<tbody>
	`)

	// Table rows
	for _, item := range items {
		result.WriteString(fmt.Sprintf(`
				<tr style="border-bottom: 1px solid #dee2e6;">
					<td style="padding: 12px;">%s</td>
					<td style="padding: 12px; text-align: center;">%d</td>
					<td style="padding: 12px; text-align: right;">$%.2f</td>
					<td style="padding: 12px; text-align: right; font-weight: bold;">$%.2f</td>
				</tr>
		`, item.ProductName, item.Quantity, item.Price, item.Subtotal))
	}

	// Table footer
	result.WriteString(`
			</tbody>
		</table>
	`)

	return result.String()
}

func (s *SendgridEmailService) SendWelcome(ctx context.Context, userName, userEmail string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail(userName, userEmail)
	subject := "Welcome to Iguanas Jewellery!"

	plainTextContent := fmt.Sprintf(`
Welcome to Iguanas Jewellery, %s!

We're thrilled to have you join our community of jewellery lovers.

Here's what you can do with your new account:
- Browse our exclusive collection of handcrafted jewellery
- Save your favorite pieces to your wishlist
- Enjoy fast and secure checkout
- Track your orders and view your purchase history

Start exploring our collection and discover the perfect piece for you or someone special.

If you have any questions, our customer support team is here to help.

Happy shopping!
The Iguanas Jewellery Team
`, userName)

	htmlContent := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; color: #333; line-height: 1.6;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c5aa0;">Welcome to Iguanas Jewellery!</h1>
        
        <p style="font-size: 18px;">Hi %s,</p>
        
        <p>We're thrilled to have you join our community of jewellery lovers.</p>
        
        <div style="background: #f8f9fa; padding: 20px; border-left: 4px solid #2c5aa0; margin: 20px 0;">
            <h3 style="margin-top: 0;">Here's what you can do with your new account:</h3>
            <ul style="margin: 10px 0; padding-left: 20px;">
                <li style="margin: 8px 0;">Browse our exclusive collection of handcrafted jewellery</li>
                <li style="margin: 8px 0;">Save your favorite pieces to your wishlist</li>
                <li style="margin: 8px 0;">Enjoy fast and secure checkout</li>
                <li style="margin: 8px 0;">Track your orders and view your purchase history</li>
            </ul>
        </div>
        
        <p>Start exploring our collection and discover the perfect piece for you or someone special.</p>
        
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://iguanas-jewellery.com/products" style="background-color: #2c5aa0; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: bold;">Browse Collection</a>
        </div>
        
        <p style="margin-top: 30px;">If you have any questions, our customer support team is here to help.</p>
        
        <p><strong>Happy shopping!</strong><br>The Iguanas Jewellery Team</p>
    </div>
</body>
</html>`, userName)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *SendgridEmailService) SendAdminOrderNotification(ctx context.Context, orderSummary OrderSummary, adminEmail string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	to := mail.NewEmail("Admin", adminEmail)
	subject := fmt.Sprintf("🛍️ New Order Received - #%s", orderSummary.ID[:8])

	plainTextContent := fmt.Sprintf(`
New Order Received!

Order Details:
Order #: %s
Customer: %s
Email: %s
Phone: %s
Total: £%.2f
Status: %s
Order Date: %s

Items Ordered:
%s

Shipping Address:
%s
%s
%s, %s %s
%s

Action Required:
- Review and process this order in the admin panel
- Prepare items for shipment
- Update order status when shipped

View order in admin panel: https://admin.iguanasjewellery.com/orders
`,
		orderSummary.ID,
		orderSummary.ShippingName,
		orderSummary.ShippingAddress.Email,
		orderSummary.ShippingAddress.Phone,
		orderSummary.Total,
		orderSummary.Status,
		orderSummary.CreatedDate.Format("January 2, 2006 at 3:04 PM"),
		formatOrderItems(orderSummary.OrderItems),
		orderSummary.ShippingName,
		orderSummary.ShippingAddress.AddressLine1,
		orderSummary.ShippingAddress.City,
		orderSummary.ShippingAddress.State,
		orderSummary.ShippingAddress.PostalCode,
		orderSummary.ShippingAddress.Country)

	htmlContent := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; color: #333; line-height: 1.6;">
	<div style="max-width: 700px; margin: 0 auto; padding: 20px;">
		<div style="background: linear-gradient(135deg, #2c5aa0 0%%, #1a3d6b 100%%); color: white; padding: 30px; border-radius: 10px 10px 0 0; text-align: center;">
			<h1 style="margin: 0; font-size: 28px;">🛍️ New Order Received!</h1>
			<p style="margin: 10px 0 0 0; font-size: 18px; opacity: 0.9;">A customer has just placed an order</p>
		</div>
		
		<div style="background: #f8f9fa; padding: 25px; border-left: 4px solid #28a745;">
			<h2 style="margin-top: 0; color: #2c5aa0;">Order Information</h2>
			<table style="width: 100%%; border-collapse: collapse;">
				<tr>
					<td style="padding: 8px 0; font-weight: bold; width: 130px;">Order #:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Customer:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Email:</td>
					<td style="padding: 8px 0;"><a href="mailto:%s" style="color: #2c5aa0;">%s</a></td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Phone:</td>
					<td style="padding: 8px 0;"><a href="tel:%s" style="color: #2c5aa0;">%s</a></td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Total:</td>
					<td style="padding: 8px 0; font-size: 24px; color: #28a745; font-weight: bold;">£%.2f</td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Status:</td>
					<td style="padding: 8px 0;"><span style="background: #ffc107; color: #000; padding: 4px 12px; border-radius: 15px; font-size: 12px; font-weight: bold; text-transform: uppercase;">%s</span></td>
				</tr>
				<tr>
					<td style="padding: 8px 0; font-weight: bold;">Order Date:</td>
					<td style="padding: 8px 0;">%s</td>
				</tr>
			</table>
		</div>
		
		<div style="margin: 25px 0;">
			<h2 style="color: #2c5aa0; border-bottom: 2px solid #dee2e6; padding-bottom: 10px;">Items Ordered</h2>
			%s
		</div>
		
		<div style="background: #fff3cd; padding: 20px; border-left: 4px solid #ffc107; margin: 25px 0;">
			<h2 style="margin-top: 0; color: #856404;">Shipping Address</h2>
			<p style="margin: 5px 0; font-size: 16px;"><strong>%s</strong></p>
			<p style="margin: 5px 0;">%s</p>
			<p style="margin: 5px 0;">%s, %s %s</p>
			<p style="margin: 5px 0;">%s</p>
		</div>
		
		<div style="background: #d1ecf1; padding: 20px; border-left: 4px solid #17a2b8; margin: 25px 0;">
			<h3 style="margin-top: 0; color: #0c5460;">⚡ Action Required</h3>
			<ul style="margin: 10px 0; padding-left: 20px; color: #0c5460;">
				<li style="margin: 8px 0;">Review and process this order in the admin panel</li>
				<li style="margin: 8px 0;">Prepare items for shipment</li>
				<li style="margin: 8px 0;">Update order status when shipped</li>
			</ul>
		</div>
		
		<div style="text-align: center; margin: 30px 0;">
			<a href="https://admin.iguanasjewellery.com/orders" style="background-color: #2c5aa0; color: white; padding: 15px 40px; text-decoration: none; border-radius: 5px; font-weight: bold; font-size: 16px; display: inline-block;">View Order in Admin Panel</a>
		</div>
		
		<div style="text-align: center; padding: 20px; color: #666; font-size: 14px; border-top: 1px solid #dee2e6;">
			<p>This is an automated notification from Iguanas Jewellery</p>
		</div>
	</div>
</body>
</html>`,
		orderSummary.ID,
		orderSummary.ShippingName,
		orderSummary.ShippingAddress.Email,
		orderSummary.ShippingAddress.Email,
		orderSummary.ShippingAddress.Phone,
		orderSummary.ShippingAddress.Phone,
		orderSummary.Total,
		orderSummary.Status,
		orderSummary.CreatedDate.Format("January 2, 2006 at 3:04 PM"),
		formatOrderItemsHTML(orderSummary.OrderItems),
		orderSummary.ShippingName,
		orderSummary.ShippingAddress.AddressLine1,
		orderSummary.ShippingAddress.City,
		orderSummary.ShippingAddress.State,
		orderSummary.ShippingAddress.PostalCode,
		orderSummary.ShippingAddress.Country)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send admin notification email: %w", err)
	}
	return nil
}
