package message

import (
	"encoding/json"
	"time"
)

// ============================================================================
// TakeoutOrder - Aligned with Grab SDK Order
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order.go
// ============================================================================

// TakeoutOrder unified takeout order model
// Fully aligned with Grab SDK Order structure
type TakeoutOrder struct {
	// OrderID The order's ID that is returned from GrabFood
	// Grab: orderID
	// Lineman: orderId
	OrderID string `json:"orderID"`

	// ShortOrderNumber The GrabFood short order number
	// Grab: shortOrderNumber
	// Lineman: orderShortCode
	ShortOrderNumber string `json:"shortOrderNumber"`

	// MerchantID The merchant's ID that is in GrabFood's database
	// Grab: merchantID
	// Lineman: storeId (mapped to PartnerMerchantID)
	MerchantID string `json:"merchantID"`

	// PartnerMerchantID The merchant's ID that is on the partner's database (optional)
	// Grab: partnerMerchantID
	// Lineman: partnerId
	PartnerMerchantID *string `json:"partnerMerchantID,omitempty"`

	// PaymentType The payment method used
	// Grab: paymentType (CASH, CASHLESS)
	// Lineman: defaults to CASH
	PaymentType string `json:"paymentType"`

	// Cutlery Whether cutlery are needed (required, non-pointer)
	// Grab: cutlery (bool, required)
	Cutlery bool `json:"cutlery"`

	// OrderTime The UTC time that a consumer places the order, based on ISO_8601/RFC3339
	// Grab: orderTime
	// Lineman: orderAcceptedTime
	OrderTime string `json:"orderTime"`

	// SubmitTime The order submit time, based on ISO_8601/RFC3339 (optional)
	// Grab: submitTime (*time.Time)
	SubmitTime *time.Time `json:"submitTime,omitempty"`

	// CompleteTime The order complete time, based on ISO_8601/RFC3339 (optional)
	// Grab: completeTime (*time.Time)
	CompleteTime *time.Time `json:"completeTime,omitempty"`

	// ScheduledTime The order scheduled time, based on ISO_8601/RFC3339 (optional)
	// Grab: scheduledTime
	ScheduledTime *string `json:"scheduledTime,omitempty"`

	// OrderState The state of the order (optional)
	// Grab: orderState
	OrderState *string `json:"orderState,omitempty"`

	// Currency Currency information (required, non-pointer)
	// Grab: currency (Currency, required)
	Currency TakeoutCurrency `json:"currency"`

	// FeatureFlags Feature related information (required, non-pointer)
	// Grab: featureFlags (OrderFeatureFlags, required)
	FeatureFlags TakeoutFeatureFlags `json:"featureFlags"`

	// Items The ordered items in an array (required)
	// Grab: items ([]OrderItem, required)
	Items []TakeoutOrderItem `json:"items"`

	// Campaigns Campaigns applicable for the order (optional)
	// Grab: campaigns
	Campaigns []TakeoutCampaign `json:"campaigns,omitempty"`

	// Promos Array of promotion objects (optional)
	// Grab: promos
	Promos []TakeoutPromo `json:"promos,omitempty"`

	// Price Order price information (required, non-pointer)
	// Grab: price (OrderPrice, required)
	Price TakeoutOrderPrice `json:"price"`

	// DineIn Dine-in information (nullable)
	// Grab: dineIn (NullableDineIn)
	DineIn NullableTakeoutDineIn `json:"dineIn,omitempty"`

	// Receiver Receiver information for delivery orders (nullable)
	// Grab: receiver (NullableReceiver)
	Receiver NullableTakeoutReceiver `json:"receiver,omitempty"`

	// OrderReadyEstimation Order ready time estimation (optional)
	// Grab: orderReadyEstimation
	OrderReadyEstimation *TakeoutOrderReadyEstimation `json:"orderReadyEstimation,omitempty"`

	// MembershipID Membership ID for loyalty project (optional)
	// Grab: membershipID
	// Lineman: memberId
	MembershipID *string `json:"membershipID,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	// Used for platform-specific fields (e.g., Lineman additionalItems)
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutOrderItem - Aligned with Grab SDK OrderItem
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_item.go
// ============================================================================

// TakeoutOrderItem order item structure
type TakeoutOrderItem struct {
	// ID The item's externalID in the partner system (required)
	// Grab: id
	// Lineman: id
	ID string `json:"id"`

	// GrabItemID The item's ID in Grab system (required)
	// Grab: grabItemID (string, required)
	GrabItemID string `json:"grabItemID"`

	// Quantity The number of the item ordered (required)
	// Grab: quantity (int32, required)
	Quantity int32 `json:"quantity"`

	// Price The price for a single item in minor unit (required)
	// Grab: price (int64, required) - in minor unit (e.g., satang for THB)
	// Lineman: unitPrice (float64, THB) - converted to minor unit
	Price int64 `json:"price"`

	// Tax Tax in minor format for a single item (optional)
	// Grab: tax (*int64)
	Tax *int64 `json:"tax,omitempty"`

	// Specifications An extra note for the merchant (optional)
	// Grab: specifications
	// Lineman: memo
	Specifications *string `json:"specifications,omitempty"`

	// OutOfStockInstruction Instructions when item is out of stock (nullable)
	// Grab: outOfStockInstruction (NullableOutOfStockInstruction)
	OutOfStockInstruction NullableTakeoutOutOfStockInstruction `json:"outOfStockInstruction,omitempty"`

	// Modifiers Array of modifiers (optional)
	// Grab: modifiers
	// Lineman: properties (converted)
	Modifiers []TakeoutModifier `json:"modifiers,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutModifier - Aligned with Grab SDK OrderItemModifier
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_item_modifier.go
// ============================================================================

// TakeoutModifier order item modifier structure
type TakeoutModifier struct {
	// ID The modifier's ID in the partner's system (optional)
	// Grab: id (*string)
	ID *string `json:"id,omitempty"`

	// Price The modifier's price in minor format (optional)
	// Grab: price (*int64)
	Price *int64 `json:"price,omitempty"`

	// Tax Tax in minor format for 1 modifier (optional)
	// Grab: tax (*int64)
	Tax *int64 `json:"tax,omitempty"`

	// Quantity The number of modifiers present (optional)
	// Grab: quantity (*int32)
	Quantity *int32 `json:"quantity,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutOrderPrice - Aligned with Grab SDK OrderPrice
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_price.go
// ============================================================================

// TakeoutOrderPrice order price information in minor unit format
type TakeoutOrderPrice struct {
	// Subtotal Total item and modifier price (tax-inclusive) in minor unit (required)
	// Grab: subtotal (int64, required)
	// Lineman: restaurantRevenue (float64, THB) - converted to minor unit
	Subtotal int64 `json:"subtotal"`

	// Tax GrabFood's tax in minor unit (optional)
	// Grab: tax (*int64)
	Tax *int64 `json:"tax,omitempty"`

	// MerchantChargeFee Additional fee charged by merchant (tax-inclusive) (optional)
	// Grab: merchantChargeFee (*int64)
	MerchantChargeFee *int64 `json:"merchantChargeFee,omitempty"`

	// GrabFundPromo GrabFood's promo fund in minor unit (optional)
	// Grab: grabFundPromo (*int64)
	GrabFundPromo *int64 `json:"grabFundPromo,omitempty"`

	// MerchantFundPromo Merchant's promo fund in minor unit (optional)
	// Grab: merchantFundPromo (*int64)
	MerchantFundPromo *int64 `json:"merchantFundPromo,omitempty"`

	// BasketPromo Total promo applied to basket items in minor unit (optional)
	// Grab: basketPromo (*int64)
	BasketPromo *int64 `json:"basketPromo,omitempty"`

	// DeliveryFee Delivery fee in minor unit (optional)
	// Grab: deliveryFee (*int64)
	DeliveryFee *int64 `json:"deliveryFee,omitempty"`

	// SmallOrderFee Fee for orders below minimum value in minor unit (optional)
	// Grab: smallOrderFee (*int64)
	SmallOrderFee *int64 `json:"smallOrderFee,omitempty"`

	// EaterPayment Total amount paid by consumer in minor unit (optional)
	// Grab: eaterPayment (*int64)
	// Lineman: restaurantRevenue - converted to minor unit
	EaterPayment *int64 `json:"eaterPayment,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutCurrency - Aligned with Grab SDK Currency
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_currency.go
// ============================================================================

// TakeoutCurrency currency information
type TakeoutCurrency struct {
	// Code The three-letter ISO currency code (required)
	// Grab: code (string, required)
	Code string `json:"code"`

	// Symbol The currency symbol (required)
	// Grab: symbol (string, required)
	Symbol string `json:"symbol"`

	// Exponent The log base 10 of major to minor unit multiplier (required)
	// Grab: exponent (int32, required)
	// 0 for VN, 2 for other countries (SG/MY/ID/TH/PH/KH/MM)
	Exponent int32 `json:"exponent"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutFeatureFlags - Aligned with Grab SDK OrderFeatureFlags
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_feature_flags.go
// ============================================================================

// TakeoutFeatureFlags feature related information
type TakeoutFeatureFlags struct {
	// OrderAcceptedType The acceptance type for the order (required)
	// Grab: orderAcceptedType (string, required)
	// Values: AUTO, MANUAL
	OrderAcceptedType string `json:"orderAcceptedType"`

	// OrderType The type of order (required)
	// Grab: orderType (string, required)
	// Values: DineIn, Delivery, Pickup, etc.
	OrderType string `json:"orderType"`

	// IsMexEditOrder Whether the order is edited by merchant (optional)
	// Grab: isMexEditOrder (*bool)
	IsMexEditOrder *bool `json:"isMexEditOrder,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutCampaign - Aligned with Grab SDK OrderCampaign
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_campaign.go
// ============================================================================

// TakeoutCampaign campaign information
type TakeoutCampaign struct {
	// ID The campaign's ID returned by GrabFood (optional)
	// Grab: id (*string)
	ID *string `json:"id,omitempty"`

	// Name The name of the campaign (optional)
	// Grab: name (*string)
	Name *string `json:"name,omitempty"`

	// CampaignNameForMex The campaign name provided by merchant (optional)
	// Grab: campaignNameForMex (*string)
	CampaignNameForMex *string `json:"campaignNameForMex,omitempty"`

	// Level The campaign level (optional)
	// Grab: level (*string)
	Level *string `json:"level,omitempty"`

	// Type The type of campaign (optional)
	// Grab: type (*string)
	Type *string `json:"type,omitempty"`

	// UsageCount The redemption count of same campaign in this order (optional)
	// Grab: usageCount (*int64)
	UsageCount *int64 `json:"usageCount,omitempty"`

	// MexFundedRatio The ratio funded by merchant in percentage (optional)
	// Grab: mexFundedRatio (*int32)
	MexFundedRatio *int32 `json:"mexFundedRatio,omitempty"`

	// DeductedAmount Total discount amount in minor unit (optional)
	// Grab: deductedAmount (*int64)
	DeductedAmount *int64 `json:"deductedAmount,omitempty"`

	// DeductedPart The part that the campaign is applied (optional)
	// Grab: deductedPart (*string)
	DeductedPart *string `json:"deductedPart,omitempty"`

	// AppliedItemIDs Array of item IDs that get discount (optional)
	// Grab: appliedItemIDs ([]string)
	AppliedItemIDs []string `json:"appliedItemIDs,omitempty"`

	// FreeItem Free item information for freeItem campaign (optional)
	// Grab: freeItem (*OrderFreeItem)
	FreeItem *TakeoutFreeItem `json:"freeItem,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutFreeItem - Aligned with Grab SDK OrderFreeItem
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_free_item.go
// ============================================================================

// TakeoutFreeItem free item information for campaign
type TakeoutFreeItem struct {
	// ID The free item's externalID (optional)
	// Grab: id (*string)
	ID *string `json:"id,omitempty"`

	// Name The name of the free item (optional)
	// Grab: name (*string)
	Name *string `json:"name,omitempty"`

	// Quantity The item's quantity, max is 1 (optional)
	// Grab: quantity (*int32)
	Quantity *int32 `json:"quantity,omitempty"`

	// Price The item's price in minor unit format (optional)
	// Grab: price (*int64)
	Price *int64 `json:"price,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutPromo - Aligned with Grab SDK OrderPromo
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_promo.go
// ============================================================================

// TakeoutPromo promotion information
type TakeoutPromo struct {
	// Code Promo code applied in the order (optional)
	// Grab: code (*string)
	// Lineman: promotionId
	Code *string `json:"code,omitempty"`

	// Description Promo description (optional)
	// Grab: description (*string)
	Description *string `json:"description,omitempty"`

	// Name Name of the promotion (optional)
	// Grab: name (*string)
	Name *string `json:"name,omitempty"`

	// PromoAmount Promo amount in local currency (optional)
	// Grab: promoAmount (*int64)
	PromoAmount *int64 `json:"promoAmount,omitempty"`

	// MexFundedRatio Merchant's funded ratio in percentage (optional)
	// Grab: mexFundedRatio (*int32)
	MexFundedRatio *int32 `json:"mexFundedRatio,omitempty"`

	// MexFundedAmount Merchant's promo fund in minor unit (optional)
	// Grab: mexFundedAmount (*int64)
	// Lineman: discount
	MexFundedAmount *int64 `json:"mexFundedAmount,omitempty"`

	// TargetedPrice Subtotal of the order basket in minor unit (optional)
	// Grab: targetedPrice (*int64)
	TargetedPrice *int64 `json:"targetedPrice,omitempty"`

	// PromoAmountInMin Promo amount in minor unit (optional)
	// Grab: promoAmountInMin (*int64)
	PromoAmountInMin *int64 `json:"promoAmountInMin,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutDineIn - Aligned with Grab SDK DineIn
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_dine_in.go
// ============================================================================

// TakeoutDineIn dine-in information
type TakeoutDineIn struct {
	// TableID Table number (optional)
	// Grab: tableID (*string)
	TableID *string `json:"tableID,omitempty"`

	// EaterCount The number of eaters (optional)
	// Grab: eaterCount (*int64)
	EaterCount *int64 `json:"eaterCount,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutReceiver - Aligned with Grab SDK Receiver
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_receiver.go
// ============================================================================

// TakeoutReceiver receiver information
type TakeoutReceiver struct {
	// Name The name of the receiver (optional)
	// Grab: name (*string)
	Name *string `json:"name,omitempty"`

	// Phones The receiver's phone number (optional)
	// Grab: phones (*string) - note: singular "phones" in Grab SDK
	Phones *string `json:"phones,omitempty"`

	// Address Receiver's location information (optional)
	// Grab: address (*Address)
	Address *TakeoutAddress `json:"address,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutAddress - Aligned with Grab SDK Address
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_address.go
// ============================================================================

// TakeoutAddress delivery address information
type TakeoutAddress struct {
	// UnitNumber The delivery address' unit number (optional)
	// Grab: unitNumber (*string)
	UnitNumber *string `json:"unitNumber,omitempty"`

	// DeliveryInstruction Instructions for the delivery (optional)
	// Grab: deliveryInstruction (*string)
	DeliveryInstruction *string `json:"deliveryInstruction,omitempty"`

	// PoiSource POI source (optional)
	// Grab: poiSource (*string)
	PoiSource *string `json:"poiSource,omitempty"`

	// PoiID POI ID, empty when poiSource is GRAB (optional)
	// Grab: poiID (*string)
	PoiID *string `json:"poiID,omitempty"`

	// Address The delivery address (optional)
	// Grab: address (*string)
	Address *string `json:"address,omitempty"`

	// Postcode The postcode of the delivery address (optional)
	// Grab: postcode (*string)
	Postcode *string `json:"postcode,omitempty"`

	// Coordinates GPS coordinates (optional)
	// Grab: coordinates (*Coordinates)
	Coordinates *TakeoutCoordinates `json:"coordinates,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutCoordinates - Aligned with Grab SDK Coordinates
// ============================================================================

// TakeoutCoordinates GPS coordinates
type TakeoutCoordinates struct {
	// Latitude GPS latitude (optional)
	// Grab: latitude (*float64)
	Latitude *float64 `json:"latitude,omitempty"`

	// Longitude GPS longitude (optional)
	// Grab: longitude (*float64)
	Longitude *float64 `json:"longitude,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutOrderReadyEstimation - Aligned with Grab SDK OrderReadyEstimation
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_ready_estimation.go
// ============================================================================

// TakeoutOrderReadyEstimation order ready time estimation
type TakeoutOrderReadyEstimation struct {
	// AllowChange Whether the order ready time can be changed (required)
	// Grab: allowChange (bool, required)
	AllowChange bool `json:"allowChange"`

	// EstimatedOrderReadyTime The estimated order ready time (required)
	// Grab: estimatedOrderReadyTime (time.Time, required)
	EstimatedOrderReadyTime time.Time `json:"estimatedOrderReadyTime"`

	// MaxOrderReadyTime The max allowed order ready time (required)
	// Grab: maxOrderReadyTime (time.Time, required)
	MaxOrderReadyTime time.Time `json:"maxOrderReadyTime"`

	// NewOrderReadyTime The new order ready time (nullable)
	// Grab: newOrderReadyTime (NullableTime)
	NewOrderReadyTime NullableTime `json:"newOrderReadyTime,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// TakeoutOutOfStockInstruction - Aligned with Grab SDK OutOfStockInstruction
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_out_of_stock_instruction.go
// ============================================================================

// TakeoutOutOfStockInstruction out-of-stock instructions
type TakeoutOutOfStockInstruction struct {
	// Title The short instruction message (optional)
	// Grab: title (*string)
	Title *string `json:"title,omitempty"`

	// InstructionType Type of out-of-stock instruction (optional)
	// Grab: instructionType (*string)
	// Values: CONTACT, SPECIFIC_ITEM, etc.
	InstructionType *string `json:"instructionType,omitempty"`

	// ReplacementItemID The preferred item's ID in partner system (optional)
	// Grab: replacementItemID (*string)
	ReplacementItemID *string `json:"replacementItemID,omitempty"`

	// ReplacementGrabItemID The preferred item's ID in Grab system (optional)
	// Grab: replacementGrabItemID (*string)
	ReplacementGrabItemID *string `json:"replacementGrabItemID,omitempty"`

	// AdditionalProperties Extra properties not defined in the schema
	// Grab: AdditionalProperties map[string]interface{}
	AdditionalProperties map[string]any `json:"-"`
}

// ============================================================================
// Nullable Types - Following Grab SDK Pattern
// ============================================================================

// NullableTakeoutDineIn nullable wrapper for TakeoutDineIn
type NullableTakeoutDineIn struct {
	value *TakeoutDineIn
	isSet bool
}

// Get returns the value
func (v NullableTakeoutDineIn) Get() *TakeoutDineIn {
	return v.value
}

// Set sets the value
func (v *NullableTakeoutDineIn) Set(val *TakeoutDineIn) {
	v.value = val
	v.isSet = true
}

// IsSet returns whether the value has been set
func (v NullableTakeoutDineIn) IsSet() bool {
	return v.isSet
}

// Unset removes the value
func (v *NullableTakeoutDineIn) Unset() {
	v.value = nil
	v.isSet = false
}

// NewNullableTakeoutDineIn creates a new NullableTakeoutDineIn
func NewNullableTakeoutDineIn(val *TakeoutDineIn) *NullableTakeoutDineIn {
	return &NullableTakeoutDineIn{value: val, isSet: true}
}

// MarshalJSON implements json.Marshaler
func (v NullableTakeoutDineIn) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *NullableTakeoutDineIn) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

// NullableTakeoutReceiver nullable wrapper for TakeoutReceiver
type NullableTakeoutReceiver struct {
	value *TakeoutReceiver
	isSet bool
}

// Get returns the value
func (v NullableTakeoutReceiver) Get() *TakeoutReceiver {
	return v.value
}

// Set sets the value
func (v *NullableTakeoutReceiver) Set(val *TakeoutReceiver) {
	v.value = val
	v.isSet = true
}

// IsSet returns whether the value has been set
func (v NullableTakeoutReceiver) IsSet() bool {
	return v.isSet
}

// Unset removes the value
func (v *NullableTakeoutReceiver) Unset() {
	v.value = nil
	v.isSet = false
}

// NewNullableTakeoutReceiver creates a new NullableTakeoutReceiver
func NewNullableTakeoutReceiver(val *TakeoutReceiver) *NullableTakeoutReceiver {
	return &NullableTakeoutReceiver{value: val, isSet: true}
}

// MarshalJSON implements json.Marshaler
func (v NullableTakeoutReceiver) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *NullableTakeoutReceiver) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

// NullableTakeoutOutOfStockInstruction nullable wrapper for TakeoutOutOfStockInstruction
type NullableTakeoutOutOfStockInstruction struct {
	value *TakeoutOutOfStockInstruction
	isSet bool
}

// Get returns the value
func (v NullableTakeoutOutOfStockInstruction) Get() *TakeoutOutOfStockInstruction {
	return v.value
}

// Set sets the value
func (v *NullableTakeoutOutOfStockInstruction) Set(val *TakeoutOutOfStockInstruction) {
	v.value = val
	v.isSet = true
}

// IsSet returns whether the value has been set
func (v NullableTakeoutOutOfStockInstruction) IsSet() bool {
	return v.isSet
}

// Unset removes the value
func (v *NullableTakeoutOutOfStockInstruction) Unset() {
	v.value = nil
	v.isSet = false
}

// NewNullableTakeoutOutOfStockInstruction creates a new NullableTakeoutOutOfStockInstruction
func NewNullableTakeoutOutOfStockInstruction(val *TakeoutOutOfStockInstruction) *NullableTakeoutOutOfStockInstruction {
	return &NullableTakeoutOutOfStockInstruction{value: val, isSet: true}
}

// MarshalJSON implements json.Marshaler
func (v NullableTakeoutOutOfStockInstruction) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *NullableTakeoutOutOfStockInstruction) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

// NullableTime nullable wrapper for time.Time
type NullableTime struct {
	value *time.Time
	isSet bool
}

// Get returns the value
func (v NullableTime) Get() *time.Time {
	return v.value
}

// Set sets the value
func (v *NullableTime) Set(val *time.Time) {
	v.value = val
	v.isSet = true
}

// IsSet returns whether the value has been set
func (v NullableTime) IsSet() bool {
	return v.isSet
}

// Unset removes the value
func (v *NullableTime) Unset() {
	v.value = nil
	v.isSet = false
}

// NewNullableTime creates a new NullableTime
func NewNullableTime(val *time.Time) *NullableTime {
	return &NullableTime{value: val, isSet: true}
}

// MarshalJSON implements json.Marshaler
func (v NullableTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

// UnmarshalJSON implements json.Unmarshaler
func (v *NullableTime) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
