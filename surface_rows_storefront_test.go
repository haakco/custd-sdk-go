package custd

import "context"

var storefrontSurfaceRows = []surfaceRow{
	{id: "billing GET /api/v1/storefronts/{storefront}/offers", path: "/api/v1/storefronts/shop/offers", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListOffers(context.Background(), "shop")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/offers", path: "/api/v1/storefronts/shop/offers", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.PublishOffer(context.Background(), "shop", StorefrontPublishOffer{Name: "Starter", BillingInterval: "month", UnitPrice: StorefrontUnitPrice{Currency: "ZAR", Amount: "15.25"}, TaxRate: "0.15", TaxLabel: "VAT"})
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/carts", path: "/api/v1/storefronts/shop/carts", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CreateCart(context.Background(), "shop")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/carts/{cart}", path: "/api/v1/storefronts/shop/carts/cart-1", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetCart(context.Background(), "shop", "cart-1")
		return err
	}},
	{id: "billing PUT /api/v1/storefronts/{storefront}/carts/{cart}/lines/{offer}", path: "/api/v1/storefronts/shop/carts/cart-1/lines/offer-1", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.UpdateLine(context.Background(), "shop", "cart-1", "offer-1", StorefrontLineQuantity{Quantity: "2.50"})
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/carts/{cart}/quotes", path: "/api/v1/storefronts/shop/carts/cart-1/quotes", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CreateQuote(context.Background(), "shop", "cart-1")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/quotes/{quote}", path: "/api/v1/storefronts/shop/quotes/quote-1", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetQuote(context.Background(), "shop", "quote-1")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/quotes/{quote}/checkout", path: "/api/v1/storefronts/shop/quotes/quote-1/checkout", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.Checkout(context.Background(), "shop", "quote-1")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/purchases/{invoice}", path: "/api/v1/storefronts/shop/purchases/invoice-1", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPurchase(context.Background(), "shop", "invoice-1")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/purchases/{invoice}/payment", path: "/api/v1/storefronts/shop/purchases/invoice-1/payment", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.StartPayment(context.Background(), "shop", "invoice-1")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/invoices/{invoice}/credit-note", path: "/api/v1/storefronts/shop/invoices/invoice-1/credit-note", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.IssueCreditNote(context.Background(), "shop", "invoice-1", StorefrontAdjustmentRequest{Reason: "Duplicate sale"})
		return err
	}},
	{id: "billing DELETE /api/v1/billing/merchant", path: "/api/v1/billing/merchant", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RevokeMerchant(context.Background(), RevokeMerchant{})
		return err
	}},
	{id: "billing DELETE /api/v1/billing/storefronts/{storefront}/analytics-destination", path: "/api/v1/billing/storefronts/shop/analytics-destination", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.DisableAnalyticsDestination(context.Background(), "shop")
		return err
	}},
	{id: "billing DELETE /api/v1/storefronts/{storefront}/usage-enrollments/{enrollment}/scheduled-issuance", path: "/api/v1/storefronts/shop/usage-enrollments/enrollment/scheduled-issuance", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CancelUsageSchedule(context.Background(), "shop", "enrollment")
		return err
	}},
	{id: "billing GET /api/v1/billing/accounting-export", path: "/api/v1/billing/accounting-export", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ReadAccountingExport(context.Background(), AccountingExportRequest{Month: "2026-11", BusinessTimeZone: "UTC"})
		return err
	}},
	{id: "billing GET /api/v1/billing/merchant", path: "/api/v1/billing/merchant", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetMerchant(context.Background())
		return err
	}},
	{id: "billing GET /api/v1/billing/partners", path: "/api/v1/billing/partners", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListBillingPartners(context.Background(), BillingPartnerPageRequest{})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/referral-reviews", path: "/api/v1/billing/partners/referral-reviews", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListReferralReviews(context.Background(), PartnerReferralPageRequest{})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/sellers", path: "/api/v1/billing/partners/sellers", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListBillingPartnerSellers(context.Background(), BillingPartnerPageRequest{})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/sellers/{seller}/settlement-preference", path: "/api/v1/billing/partners/sellers/seller/settlement-preference", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetSellerSettlementPreference(context.Background(), "seller")
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/sellers/{seller}/tax-disclosure", path: "/api/v1/billing/partners/sellers/seller/tax-disclosure", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetSellerTaxDisclosure(context.Background(), "seller")
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/{partner}/fee-balance", path: "/api/v1/billing/partners/partner/fee-balance", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPartnerFeeBalance(context.Background(), "partner", PartnerFeeBalanceRequest{AsOf: "2026-11-01"})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/{partner}/payouts", path: "/api/v1/billing/partners/partner/payouts", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListPartnerPayouts(context.Background(), PartnerPayoutPageRequest{PartnerCompanyUUID: "partner"})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/{partner}/referral-terms", path: "/api/v1/billing/partners/partner/referral-terms", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListPartnerReferralTerms(context.Background(), "partner", PartnerReferralPageRequest{})
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/{partner}/settlement-preference", path: "/api/v1/billing/partners/partner/settlement-preference", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPartnerSettlementPreference(context.Background(), "partner")
		return err
	}},
	{id: "billing GET /api/v1/billing/partners/{partner}/tax-disclosure", path: "/api/v1/billing/partners/partner/tax-disclosure", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPartnerTaxDisclosure(context.Background(), "partner")
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts", path: "/api/v1/billing/storefronts", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListOwnedStorefronts(context.Background())
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/analytics-destination", path: "/api/v1/billing/storefronts/shop/analytics-destination", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetAnalyticsDestination(context.Background(), "shop")
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/invoices", path: "/api/v1/billing/storefronts/shop/invoices", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListSellerInvoices(context.Background(), SellerInvoiceListRequest{Storefront: "shop"})
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/offers/{offer}/partner-terms", path: "/api/v1/billing/storefronts/shop/offers/offer/partner-terms", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListPartnerProductTerms(context.Background(), PartnerProductTermsPageRequest{Storefront: "shop", OfferUUID: "offer"})
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/partners/{partner}/statement", path: "/api/v1/billing/storefronts/shop/partners/partner/statement", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPartnerStatement(context.Background(), StorefrontPartnerStatementRequest{Storefront: "shop", PartnerCompanyUUID: "partner", Month: "2026-11"})
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/purchase-handoffs/{handoff}", path: "/api/v1/billing/storefronts/shop/purchase-handoffs/handoff", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetPurchaseHandoff(context.Background(), "shop", "handoff")
		return err
	}},
	{id: "billing GET /api/v1/billing/storefronts/{storefront}/receipt-template", path: "/api/v1/billing/storefronts/shop/receipt-template", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetReceiptTemplate(context.Background(), "shop")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/purchases/{invoice}/set-off", path: "/api/v1/storefronts/shop/purchases/invoice/set-off", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetStorefrontSetOffEligibility(context.Background(), "shop", "invoice")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/usage-enrollments", path: "/api/v1/storefronts/shop/usage-enrollments", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListUsageEnrollments(context.Background(), StorefrontUsagePageRequest{Storefront: "shop"})
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/usage-enrollments/{enrollment}/scheduled-issuance", path: "/api/v1/storefronts/shop/usage-enrollments/enrollment/scheduled-issuance", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetUsageSchedule(context.Background(), "shop", "enrollment")
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/usage-tariffs", path: "/api/v1/storefronts/shop/usage-tariffs", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ListUsageTariffs(context.Background(), StorefrontUsagePageRequest{Storefront: "shop"})
		return err
	}},
	{id: "billing GET /api/v1/storefronts/{storefront}/usage-tariffs/{revision}", path: "/api/v1/storefronts/shop/usage-tariffs/revision", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.GetUsageTariff(context.Background(), "shop", "revision")
		return err
	}},
	{id: "billing POST /api/v1/billing/merchant/test", path: "/api/v1/billing/merchant/test", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.VerifyMerchant(context.Background(), VerifyMerchant{})
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/invitations", path: "/api/v1/billing/partners/invitations", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CreateBillingPartnerInvitation(context.Background())
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/invitations/accept", path: "/api/v1/billing/partners/invitations/accept", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.AcceptBillingPartnerInvitation(context.Background(), AcceptBillingPartnerInvitationRequest{})
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/referral-reviews/{attribution}/retry", path: "/api/v1/billing/partners/referral-reviews/attribution/retry", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RetryReferralReview(context.Background(), "attribution")
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/sellers/{seller}/settlement-preference", path: "/api/v1/billing/partners/sellers/seller/settlement-preference", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RecordPartnerSettlementPreference(context.Background(), "seller", RecordPartnerSettlementPreferenceRequest{})
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/sellers/{seller}/tax-disclosure", path: "/api/v1/billing/partners/sellers/seller/tax-disclosure", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RecordPartnerTaxDisclosure(context.Background(), "seller", RecordPartnerTaxDisclosureRequest{})
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/{partner}/payouts", path: "/api/v1/billing/partners/partner/payouts", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RecordPartnerPayout(context.Background(), "partner", RecordPartnerPayoutRequest{})
		return err
	}},
	{id: "billing POST /api/v1/billing/partners/{partner}/referral-terms", path: "/api/v1/billing/partners/partner/referral-terms", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.PublishPartnerReferralTerms(context.Background(), "partner", PublishReferralTerms{})
		return err
	}},
	{id: "billing POST /api/v1/billing/storefronts", path: "/api/v1/billing/storefronts", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RegisterStorefront(context.Background(), RegisterStorefront{})
		return err
	}},
	{id: "billing POST /api/v1/billing/storefronts/{storefront}/capacity-offers", path: "/api/v1/billing/storefronts/shop/capacity-offers", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.PublishCapacityOffer(context.Background(), "shop", PublishStorefrontCapacityOffer{})
		return err
	}},
	{id: "billing POST /api/v1/billing/storefronts/{storefront}/offers/{offer}/partner-terms", path: "/api/v1/billing/storefronts/shop/offers/offer/partner-terms", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.PublishPartnerProductTerms(context.Background(), "shop", "offer", PublishPartnerProductTermsRequest{})
		return err
	}},
	{id: "billing POST /api/v1/billing/storefronts/{storefront}/purchase-handoffs", path: "/api/v1/billing/storefronts/shop/purchase-handoffs", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CreatePurchaseHandoff(context.Background(), "shop", CreateStorefrontPurchaseHandoff{})
		return err
	}},
	{id: "billing POST /api/v1/billing/storefronts/{storefront}/purchase-handoffs/{handoff}/cancel", path: "/api/v1/billing/storefronts/shop/purchase-handoffs/handoff/cancel", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.CancelPurchaseHandoff(context.Background(), "shop", "handoff")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/invoices/{invoice}/refund", path: "/api/v1/storefronts/shop/invoices/invoice/refund", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RequestRefund(context.Background(), "shop", "invoice", StorefrontAdjustmentRequest{})
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/invoices/{invoice}/refund/reconcile", path: "/api/v1/storefronts/shop/invoices/invoice/refund/reconcile", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ReconcileRefund(context.Background(), "shop", "invoice")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/purchase-handoffs/{handoff}/redeem", path: "/api/v1/storefronts/shop/purchase-handoffs/handoff/redeem", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.RedeemPurchaseHandoff(context.Background(), "shop", "handoff")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/purchases/{invoice}/set-off", path: "/api/v1/storefronts/shop/purchases/invoice/set-off", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ApplyStorefrontSetOff(context.Background(), "shop", "invoice")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/quotes/{quote}/invoice", path: "/api/v1/storefronts/shop/quotes/quote/invoice", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.IssueStorefrontInvoice(context.Background(), "shop", "quote")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/usage-enrollments", path: "/api/v1/storefronts/shop/usage-enrollments", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.EnrollUsage(context.Background(), "shop", EnrollUsageRequest{})
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/usage-enrollments/{enrollment}/issue", path: "/api/v1/storefronts/shop/usage-enrollments/enrollment/issue", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.IssueUsageInvoice(context.Background(), "shop", "enrollment")
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/usage-enrollments/{enrollment}/scheduled-issuance", path: "/api/v1/storefronts/shop/usage-enrollments/enrollment/scheduled-issuance", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ScheduleUsageInvoice(context.Background(), "shop", "enrollment", UsageInvoiceOptIn{})
		return err
	}},
	{id: "billing POST /api/v1/storefronts/{storefront}/usage-tariffs", path: "/api/v1/storefronts/shop/usage-tariffs", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.PublishUsageTariff(context.Background(), "shop", PublishUsageTariffRequest{})
		return err
	}},
	{id: "billing PUT /api/v1/billing/merchant", path: "/api/v1/billing/merchant", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.ConfigureMerchant(context.Background(), ConfigureMerchant{})
		return err
	}},
	{id: "billing PUT /api/v1/billing/storefronts/{storefront}/analytics-destination", path: "/api/v1/billing/storefronts/shop/analytics-destination", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.SelectAnalyticsDestination(context.Background(), "shop", SelectStorefrontAnalyticsDestination{})
		return err
	}},
	{id: "billing PUT /api/v1/billing/storefronts/{storefront}/receipt-template", path: "/api/v1/billing/storefronts/shop/receipt-template", call: func(e *surfaceEnv) error {
		_, err := e.client.Storefronts.SelectReceiptTemplate(context.Background(), "shop", SelectStorefrontReceiptTemplate{})
		return err
	}},
}
