package message

import (
	"encoding/json"
	"reflect"
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
	// Grab: featureFlags (TakeoutFeatureFlags, required)
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
	// Grab: price (TakeoutOrderPrice, required)
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

// GetOrderID returns the OrderID field value
func (o *TakeoutOrder) GetOrderID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OrderID
}

// GetOrderIDOk returns a tuple with the OrderID field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetOrderIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OrderID, true
}

// SetOrderID sets field value
func (o *TakeoutOrder) SetOrderID(v string) {
	o.OrderID = v
}

// GetShortOrderNumber returns the ShortOrderNumber field value
func (o *TakeoutOrder) GetShortOrderNumber() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ShortOrderNumber
}

// GetShortOrderNumberOk returns a tuple with the ShortOrderNumber field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetShortOrderNumberOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ShortOrderNumber, true
}

// SetShortOrderNumber sets field value
func (o *TakeoutOrder) SetShortOrderNumber(v string) {
	o.ShortOrderNumber = v
}

// GetMerchantID returns the MerchantID field value
func (o *TakeoutOrder) GetMerchantID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.MerchantID
}

// GetMerchantIDOk returns a tuple with the MerchantID field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetMerchantIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.MerchantID, true
}

// SetMerchantID sets field value
func (o *TakeoutOrder) SetMerchantID(v string) {
	o.MerchantID = v
}

// GetPartnerMerchantID returns the PartnerMerchantID field value if set, zero value otherwise.
func (o *TakeoutOrder) GetPartnerMerchantID() string {
	if o == nil || IsNil(o.PartnerMerchantID) {
		var ret string
		return ret
	}
	return *o.PartnerMerchantID
}

// GetPartnerMerchantIDOk returns a tuple with the PartnerMerchantID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetPartnerMerchantIDOk() (*string, bool) {
	if o == nil || IsNil(o.PartnerMerchantID) {
		return nil, false
	}
	return o.PartnerMerchantID, true
}

// HasPartnerMerchantID returns a boolean if a field has been set.
func (o *TakeoutOrder) HasPartnerMerchantID() bool {
	if o != nil && !IsNil(o.PartnerMerchantID) {
		return true
	}

	return false
}

// SetPartnerMerchantID gets a reference to the given string and assigns it to the PartnerMerchantID field.
func (o *TakeoutOrder) SetPartnerMerchantID(v string) {
	o.PartnerMerchantID = &v
}

// GetPaymentType returns the PaymentType field value
func (o *TakeoutOrder) GetPaymentType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.PaymentType
}

// GetPaymentTypeOk returns a tuple with the PaymentType field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetPaymentTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PaymentType, true
}

// SetPaymentType sets field value
func (o *TakeoutOrder) SetPaymentType(v string) {
	o.PaymentType = v
}

// GetCutlery returns the Cutlery field value
func (o *TakeoutOrder) GetCutlery() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Cutlery
}

// GetCutleryOk returns a tuple with the Cutlery field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetCutleryOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Cutlery, true
}

// SetCutlery sets field value
func (o *TakeoutOrder) SetCutlery(v bool) {
	o.Cutlery = v
}

// GetOrderTime returns the OrderTime field value
func (o *TakeoutOrder) GetOrderTime() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OrderTime
}

// GetOrderTimeOk returns a tuple with the OrderTime field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetOrderTimeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OrderTime, true
}

// SetOrderTime sets field value
func (o *TakeoutOrder) SetOrderTime(v string) {
	o.OrderTime = v
}

// GetSubmitTime returns the SubmitTime field value if set, zero value otherwise.
func (o *TakeoutOrder) GetSubmitTime() time.Time {
	if o == nil || IsNil(o.SubmitTime) {
		var ret time.Time
		return ret
	}
	return *o.SubmitTime
}

// GetSubmitTimeOk returns a tuple with the SubmitTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetSubmitTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.SubmitTime) {
		return nil, false
	}
	return o.SubmitTime, true
}

// HasSubmitTime returns a boolean if a field has been set.
func (o *TakeoutOrder) HasSubmitTime() bool {
	if o != nil && !IsNil(o.SubmitTime) {
		return true
	}

	return false
}

// SetSubmitTime gets a reference to the given time.Time and assigns it to the SubmitTime field.
func (o *TakeoutOrder) SetSubmitTime(v time.Time) {
	o.SubmitTime = &v
}

// GetCompleteTime returns the CompleteTime field value if set, zero value otherwise.
func (o *TakeoutOrder) GetCompleteTime() time.Time {
	if o == nil || IsNil(o.CompleteTime) {
		var ret time.Time
		return ret
	}
	return *o.CompleteTime
}

// GetCompleteTimeOk returns a tuple with the CompleteTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetCompleteTimeOk() (*time.Time, bool) {
	if o == nil || IsNil(o.CompleteTime) {
		return nil, false
	}
	return o.CompleteTime, true
}

// HasCompleteTime returns a boolean if a field has been set.
func (o *TakeoutOrder) HasCompleteTime() bool {
	if o != nil && !IsNil(o.CompleteTime) {
		return true
	}

	return false
}

// SetCompleteTime gets a reference to the given time.Time and assigns it to the CompleteTime field.
func (o *TakeoutOrder) SetCompleteTime(v time.Time) {
	o.CompleteTime = &v
}

// GetScheduledTime returns the ScheduledTime field value if set, zero value otherwise.
func (o *TakeoutOrder) GetScheduledTime() string {
	if o == nil || IsNil(o.ScheduledTime) {
		var ret string
		return ret
	}
	return *o.ScheduledTime
}

// GetScheduledTimeOk returns a tuple with the ScheduledTime field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetScheduledTimeOk() (*string, bool) {
	if o == nil || IsNil(o.ScheduledTime) {
		return nil, false
	}
	return o.ScheduledTime, true
}

// HasScheduledTime returns a boolean if a field has been set.
func (o *TakeoutOrder) HasScheduledTime() bool {
	if o != nil && !IsNil(o.ScheduledTime) {
		return true
	}

	return false
}

// SetScheduledTime gets a reference to the given string and assigns it to the ScheduledTime field.
func (o *TakeoutOrder) SetScheduledTime(v string) {
	o.ScheduledTime = &v
}

// GetOrderState returns the OrderState field value if set, zero value otherwise.
func (o *TakeoutOrder) GetOrderState() string {
	if o == nil || IsNil(o.OrderState) {
		var ret string
		return ret
	}
	return *o.OrderState
}

// GetOrderStateOk returns a tuple with the OrderState field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetOrderStateOk() (*string, bool) {
	if o == nil || IsNil(o.OrderState) {
		return nil, false
	}
	return o.OrderState, true
}

// HasOrderState returns a boolean if a field has been set.
func (o *TakeoutOrder) HasOrderState() bool {
	if o != nil && !IsNil(o.OrderState) {
		return true
	}

	return false
}

// SetOrderState gets a reference to the given string and assigns it to the OrderState field.
func (o *TakeoutOrder) SetOrderState(v string) {
	o.OrderState = &v
}

// GetCurrency returns the Currency field value
func (o *TakeoutOrder) GetCurrency() TakeoutCurrency {
	if o == nil {
		var ret TakeoutCurrency
		return ret
	}

	return o.Currency
}

// GetCurrencyOk returns a tuple with the Currency field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetCurrencyOk() (*TakeoutCurrency, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Currency, true
}

// SetCurrency sets field value
func (o *TakeoutOrder) SetCurrency(v TakeoutCurrency) {
	o.Currency = v
}

// GetFeatureFlags returns the FeatureFlags field value
func (o *TakeoutOrder) GetFeatureFlags() TakeoutFeatureFlags {
	if o == nil {
		var ret TakeoutFeatureFlags
		return ret
	}

	return o.FeatureFlags
}

// GetFeatureFlagsOk returns a tuple with the FeatureFlags field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetFeatureFlagsOk() (*TakeoutFeatureFlags, bool) {
	if o == nil {
		return nil, false
	}
	return &o.FeatureFlags, true
}

// SetFeatureFlags sets field value
func (o *TakeoutOrder) SetFeatureFlags(v TakeoutFeatureFlags) {
	o.FeatureFlags = v
}

// GetItems returns the Items field value
func (o *TakeoutOrder) GetItems() []TakeoutOrderItem {
	if o == nil {
		var ret []TakeoutOrderItem
		return ret
	}

	return o.Items
}

// GetItemsOk returns a tuple with the Items field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetItemsOk() ([]TakeoutOrderItem, bool) {
	if o == nil {
		return nil, false
	}
	return o.Items, true
}

// SetItems sets field value
func (o *TakeoutOrder) SetItems(v []TakeoutOrderItem) {
	o.Items = v
}

// GetCampaigns returns the Campaigns field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *TakeoutOrder) GetCampaigns() []TakeoutCampaign {
	if o == nil {
		var ret []TakeoutCampaign
		return ret
	}
	return o.Campaigns
}

// GetCampaignsOk returns a tuple with the Campaigns field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *TakeoutOrder) GetCampaignsOk() ([]TakeoutCampaign, bool) {
	if o == nil || IsNil(o.Campaigns) {
		return nil, false
	}
	return o.Campaigns, true
}

// HasCampaigns returns a boolean if a field has been set.
func (o *TakeoutOrder) HasCampaigns() bool {
	if o != nil && !IsNil(o.Campaigns) {
		return true
	}

	return false
}

// SetCampaigns gets a reference to the given []OrderCampaign and assigns it to the Campaigns field.
func (o *TakeoutOrder) SetCampaigns(v []TakeoutCampaign) {
	o.Campaigns = v
}

// GetPromos returns the Promos field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *TakeoutOrder) GetPromos() []TakeoutPromo {
	if o == nil {
		var ret []TakeoutPromo
		return ret
	}
	return o.Promos
}

// GetPromosOk returns a tuple with the Promos field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *TakeoutOrder) GetPromosOk() ([]TakeoutPromo, bool) {
	if o == nil || IsNil(o.Promos) {
		return nil, false
	}
	return o.Promos, true
}

// HasPromos returns a boolean if a field has been set.
func (o *TakeoutOrder) HasPromos() bool {
	if o != nil && !IsNil(o.Promos) {
		return true
	}

	return false
}

// SetPromos gets a reference to the given []OrderPromo and assigns it to the Promos field.
func (o *TakeoutOrder) SetPromos(v []TakeoutPromo) {
	o.Promos = v
}

// GetPrice returns the Price field value
func (o *TakeoutOrder) GetPrice() TakeoutOrderPrice {
	if o == nil {
		var ret TakeoutOrderPrice
		return ret
	}

	return o.Price
}

// GetPriceOk returns a tuple with the Price field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetPriceOk() (*TakeoutOrderPrice, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Price, true
}

// SetPrice sets field value
func (o *TakeoutOrder) SetPrice(v TakeoutOrderPrice) {
	o.Price = v
}

// GetDineIn returns the DineIn field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *TakeoutOrder) GetDineIn() TakeoutDineIn {
	if o == nil || IsNil(o.DineIn.Get()) {
		var ret TakeoutDineIn
		return ret
	}
	return *o.DineIn.Get()
}

// GetDineInOk returns a tuple with the DineIn field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *TakeoutOrder) GetDineInOk() (*TakeoutDineIn, bool) {
	if o == nil {
		return nil, false
	}
	return o.DineIn.Get(), o.DineIn.IsSet()
}

// HasDineIn returns a boolean if a field has been set.
func (o *TakeoutOrder) HasDineIn() bool {
	if o != nil && o.DineIn.IsSet() {
		return true
	}

	return false
}

// SetDineIn gets a reference to the given NullableDineIn and assigns it to the DineIn field.
func (o *TakeoutOrder) SetDineIn(v TakeoutDineIn) {
	o.DineIn.Set(&v)
}

// SetDineInNil sets the value for DineIn to be an explicit nil
func (o *TakeoutOrder) SetDineInNil() {
	o.DineIn.Set(nil)
}

// UnsetDineIn ensures that no value is present for DineIn, not even an explicit nil
func (o *TakeoutOrder) UnsetDineIn() {
	o.DineIn.Unset()
}

// GetReceiver returns the Receiver field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *TakeoutOrder) GetReceiver() TakeoutReceiver {
	if o == nil || IsNil(o.Receiver.Get()) {
		var ret TakeoutReceiver
		return ret
	}
	return *o.Receiver.Get()
}

// GetReceiverOk returns a tuple with the Receiver field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *TakeoutOrder) GetReceiverOk() (*TakeoutReceiver, bool) {
	if o == nil {
		return nil, false
	}
	return o.Receiver.Get(), o.Receiver.IsSet()
}

// HasReceiver returns a boolean if a field has been set.
func (o *TakeoutOrder) HasReceiver() bool {
	if o != nil && o.Receiver.IsSet() {
		return true
	}

	return false
}

// SetReceiver gets a reference to the given NullableReceiver and assigns it to the Receiver field.
func (o *TakeoutOrder) SetReceiver(v TakeoutReceiver) {
	o.Receiver.Set(&v)
}

// SetReceiverNil sets the value for Receiver to be an explicit nil
func (o *TakeoutOrder) SetReceiverNil() {
	o.Receiver.Set(nil)
}

// UnsetReceiver ensures that no value is present for Receiver, not even an explicit nil
func (o *TakeoutOrder) UnsetReceiver() {
	o.Receiver.Unset()
}

// GetOrderReadyEstimation returns the OrderReadyEstimation field value if set, zero value otherwise.
func (o *TakeoutOrder) GetOrderReadyEstimation() TakeoutOrderReadyEstimation {
	if o == nil || IsNil(o.OrderReadyEstimation) {
		var ret TakeoutOrderReadyEstimation
		return ret
	}
	return *o.OrderReadyEstimation
}

// GetOrderReadyEstimationOk returns a tuple with the OrderReadyEstimation field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetOrderReadyEstimationOk() (*TakeoutOrderReadyEstimation, bool) {
	if o == nil || IsNil(o.OrderReadyEstimation) {
		return nil, false
	}
	return o.OrderReadyEstimation, true
}

// HasOrderReadyEstimation returns a boolean if a field has been set.
func (o *TakeoutOrder) HasOrderReadyEstimation() bool {
	if o != nil && !IsNil(o.OrderReadyEstimation) {
		return true
	}

	return false
}

// SetOrderReadyEstimation gets a reference to the given OrderReadyEstimation and assigns it to the OrderReadyEstimation field.
func (o *TakeoutOrder) SetOrderReadyEstimation(v TakeoutOrderReadyEstimation) {
	o.OrderReadyEstimation = &v
}

// GetMembershipID returns the MembershipID field value if set, zero value otherwise.
func (o *TakeoutOrder) GetMembershipID() string {
	if o == nil || IsNil(o.MembershipID) {
		var ret string
		return ret
	}
	return *o.MembershipID
}

// GetMembershipIDOk returns a tuple with the MembershipID field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrder) GetMembershipIDOk() (*string, bool) {
	if o == nil || IsNil(o.MembershipID) {
		return nil, false
	}
	return o.MembershipID, true
}

// HasMembershipID returns a boolean if a field has been set.
func (o *TakeoutOrder) HasMembershipID() bool {
	if o != nil && !IsNil(o.MembershipID) {
		return true
	}

	return false
}

// SetMembershipID gets a reference to the given string and assigns it to the MembershipID field.
func (o *TakeoutOrder) SetMembershipID(v string) {
	o.MembershipID = &v
}

// MarshalJSON implements json.Marshaler interface
// This method ensures AdditionalProperties are merged into the JSON output
func (o TakeoutOrder) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

// ToMap converts the TakeoutOrder to a map for custom JSON serialization
func (o TakeoutOrder) ToMap() (map[string]any, error) {
	toSerialize := map[string]any{}

	// Required fields
	toSerialize["orderID"] = o.OrderID
	toSerialize["shortOrderNumber"] = o.ShortOrderNumber
	toSerialize["merchantID"] = o.MerchantID
	toSerialize["paymentType"] = o.PaymentType
	toSerialize["cutlery"] = o.Cutlery
	toSerialize["orderTime"] = o.OrderTime
	toSerialize["currency"] = o.Currency
	toSerialize["featureFlags"] = o.FeatureFlags
	toSerialize["items"] = o.Items
	toSerialize["price"] = o.Price

	// Optional fields
	if !IsNil(o.PartnerMerchantID) {
		toSerialize["partnerMerchantID"] = o.PartnerMerchantID
	}
	if !IsNil(o.SubmitTime) {
		toSerialize["submitTime"] = o.SubmitTime
	}
	if !IsNil(o.CompleteTime) {
		toSerialize["completeTime"] = o.CompleteTime
	}
	if !IsNil(o.ScheduledTime) {
		toSerialize["scheduledTime"] = o.ScheduledTime
	}
	if !IsNil(o.OrderState) {
		toSerialize["orderState"] = o.OrderState
	}
	if !IsNil(o.Campaigns) {
		toSerialize["campaigns"] = o.Campaigns
	}
	if !IsNil(o.Promos) {
		toSerialize["promos"] = o.Promos
	}
	if o.DineIn.IsSet() {
		toSerialize["dineIn"] = o.DineIn.Get()
	}
	if o.Receiver.IsSet() {
		toSerialize["receiver"] = o.Receiver.Get()
	}
	if !IsNil(o.OrderReadyEstimation) {
		toSerialize["orderReadyEstimation"] = o.OrderReadyEstimation
	}
	if !IsNil(o.MembershipID) {
		toSerialize["membershipID"] = o.MembershipID
	}

	// Merge AdditionalProperties into the output
	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
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

// GetId returns the Id field value
func (o *TakeoutOrderItem) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ID
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ID, true
}

// SetId sets field value
func (o *TakeoutOrderItem) SetId(v string) {
	o.ID = v
}

// GetGrabItemID returns the GrabItemID field value
func (o *TakeoutOrderItem) GetGrabItemID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.GrabItemID
}

// GetGrabItemIDOk returns a tuple with the GrabItemID field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetGrabItemIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.GrabItemID, true
}

// SetGrabItemID sets field value
func (o *TakeoutOrderItem) SetGrabItemID(v string) {
	o.GrabItemID = v
}

// GetQuantity returns the Quantity field value
func (o *TakeoutOrderItem) GetQuantity() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Quantity
}

// GetQuantityOk returns a tuple with the Quantity field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetQuantityOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Quantity, true
}

// SetQuantity sets field value
func (o *TakeoutOrderItem) SetQuantity(v int32) {
	o.Quantity = v
}

// GetPrice returns the Price field value
func (o *TakeoutOrderItem) GetPrice() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.Price
}

// GetPriceOk returns a tuple with the Price field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetPriceOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Price, true
}

// SetPrice sets field value
func (o *TakeoutOrderItem) SetPrice(v int64) {
	o.Price = v
}

// GetTax returns the Tax field value if set, zero value otherwise.
func (o *TakeoutOrderItem) GetTax() int64 {
	if o == nil || IsNil(o.Tax) {
		var ret int64
		return ret
	}
	return *o.Tax
}

// GetTaxOk returns a tuple with the Tax field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetTaxOk() (*int64, bool) {
	if o == nil || IsNil(o.Tax) {
		return nil, false
	}
	return o.Tax, true
}

// HasTax returns a boolean if a field has been set.
func (o *TakeoutOrderItem) HasTax() bool {
	if o != nil && !IsNil(o.Tax) {
		return true
	}

	return false
}

// SetTax gets a reference to the given int64 and assigns it to the Tax field.
func (o *TakeoutOrderItem) SetTax(v int64) {
	o.Tax = &v
}

// GetSpecifications returns the Specifications field value if set, zero value otherwise.
func (o *TakeoutOrderItem) GetSpecifications() string {
	if o == nil || IsNil(o.Specifications) {
		var ret string
		return ret
	}
	return *o.Specifications
}

// GetSpecificationsOk returns a tuple with the Specifications field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetSpecificationsOk() (*string, bool) {
	if o == nil || IsNil(o.Specifications) {
		return nil, false
	}
	return o.Specifications, true
}

// HasSpecifications returns a boolean if a field has been set.
func (o *TakeoutOrderItem) HasSpecifications() bool {
	if o != nil && !IsNil(o.Specifications) {
		return true
	}

	return false
}

// SetSpecifications gets a reference to the given string and assigns it to the Specifications field.
func (o *TakeoutOrderItem) SetSpecifications(v string) {
	o.Specifications = &v
}

// GetOutOfStockInstruction returns the OutOfStockInstruction field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *TakeoutOrderItem) GetOutOfStockInstruction() TakeoutOutOfStockInstruction {
	if o == nil || IsNil(o.OutOfStockInstruction.Get()) {
		var ret TakeoutOutOfStockInstruction
		return ret
	}
	return *o.OutOfStockInstruction.Get()
}

// GetOutOfStockInstructionOk returns a tuple with the OutOfStockInstruction field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *TakeoutOrderItem) GetOutOfStockInstructionOk() (*TakeoutOutOfStockInstruction, bool) {
	if o == nil {
		return nil, false
	}
	return o.OutOfStockInstruction.Get(), o.OutOfStockInstruction.IsSet()
}

// HasOutOfStockInstruction returns a boolean if a field has been set.
func (o *TakeoutOrderItem) HasOutOfStockInstruction() bool {
	if o != nil && o.OutOfStockInstruction.IsSet() {
		return true
	}

	return false
}

// SetOutOfStockInstruction gets a reference to the given NullableOutOfStockInstruction and assigns it to the OutOfStockInstruction field.
func (o *TakeoutOrderItem) SetOutOfStockInstruction(v TakeoutOutOfStockInstruction) {
	o.OutOfStockInstruction.Set(&v)
}

// SetOutOfStockInstructionNil sets the value for OutOfStockInstruction to be an explicit nil
func (o *TakeoutOrderItem) SetOutOfStockInstructionNil() {
	o.OutOfStockInstruction.Set(nil)
}

// UnsetOutOfStockInstruction ensures that no value is present for OutOfStockInstruction, not even an explicit nil
func (o *TakeoutOrderItem) UnsetOutOfStockInstruction() {
	o.OutOfStockInstruction.Unset()
}

// GetModifiers returns the Modifiers field value if set, zero value otherwise.
func (o *TakeoutOrderItem) GetModifiers() []TakeoutModifier {
	if o == nil || IsNil(o.Modifiers) {
		var ret []TakeoutModifier
		return ret
	}
	return o.Modifiers
}

// GetModifiersOk returns a tuple with the Modifiers field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderItem) GetModifiersOk() ([]TakeoutModifier, bool) {
	if o == nil || IsNil(o.Modifiers) {
		return nil, false
	}
	return o.Modifiers, true
}

// HasModifiers returns a boolean if a field has been set.
func (o *TakeoutOrderItem) HasModifiers() bool {
	if o != nil && !IsNil(o.Modifiers) {
		return true
	}

	return false
}

// SetModifiers gets a reference to the given []OrderItemModifier and assigns it to the Modifiers field.
func (o *TakeoutOrderItem) SetModifiers(v []TakeoutModifier) {
	o.Modifiers = v
}

// MarshalJSON implements json.Marshaler interface
// This method ensures AdditionalProperties are merged into the JSON output
func (o TakeoutOrderItem) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

// ToMap converts the TakeoutOrderItem to a map for custom JSON serialization
func (o TakeoutOrderItem) ToMap() (map[string]any, error) {
	toSerialize := map[string]any{}

	// Required fields
	toSerialize["id"] = o.ID
	toSerialize["grabItemID"] = o.GrabItemID
	toSerialize["quantity"] = o.Quantity
	toSerialize["price"] = o.Price

	// Optional fields
	if !IsNil(o.Tax) {
		toSerialize["tax"] = o.Tax
	}
	if !IsNil(o.Specifications) {
		toSerialize["specifications"] = o.Specifications
	}
	if o.OutOfStockInstruction.IsSet() {
		toSerialize["outOfStockInstruction"] = o.OutOfStockInstruction.Get()
	}
	if !IsNil(o.Modifiers) {
		toSerialize["modifiers"] = o.Modifiers
	}

	// Merge AdditionalProperties into the output
	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
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

// GetId returns the Id field value if set, zero value otherwise.
func (o *TakeoutModifier) GetId() string {
	if o == nil || IsNil(o.ID) {
		var ret string
		return ret
	}
	return *o.ID
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutModifier) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.ID) {
		return nil, false
	}
	return o.ID, true
}

// HasId returns a boolean if a field has been set.
func (o *TakeoutModifier) HasId() bool {
	if o != nil && !IsNil(o.ID) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *TakeoutModifier) SetId(v string) {
	o.ID = &v
}

// GetPrice returns the Price field value if set, zero value otherwise.
func (o *TakeoutModifier) GetPrice() int64 {
	if o == nil || IsNil(o.Price) {
		var ret int64
		return ret
	}
	return *o.Price
}

// GetPriceOk returns a tuple with the Price field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutModifier) GetPriceOk() (*int64, bool) {
	if o == nil || IsNil(o.Price) {
		return nil, false
	}
	return o.Price, true
}

// HasPrice returns a boolean if a field has been set.
func (o *TakeoutModifier) HasPrice() bool {
	if o != nil && !IsNil(o.Price) {
		return true
	}

	return false
}

// SetPrice gets a reference to the given int64 and assigns it to the Price field.
func (o *TakeoutModifier) SetPrice(v int64) {
	o.Price = &v
}

// GetTax returns the Tax field value if set, zero value otherwise.
func (o *TakeoutModifier) GetTax() int64 {
	if o == nil || IsNil(o.Tax) {
		var ret int64
		return ret
	}
	return *o.Tax
}

// GetTaxOk returns a tuple with the Tax field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutModifier) GetTaxOk() (*int64, bool) {
	if o == nil || IsNil(o.Tax) {
		return nil, false
	}
	return o.Tax, true
}

// HasTax returns a boolean if a field has been set.
func (o *TakeoutModifier) HasTax() bool {
	if o != nil && !IsNil(o.Tax) {
		return true
	}

	return false
}

// SetTax gets a reference to the given int64 and assigns it to the Tax field.
func (o *TakeoutModifier) SetTax(v int64) {
	o.Tax = &v
}

// GetQuantity returns the Quantity field value if set, zero value otherwise.
func (o *TakeoutModifier) GetQuantity() int32 {
	if o == nil || IsNil(o.Quantity) {
		var ret int32
		return ret
	}
	return *o.Quantity
}

// GetQuantityOk returns a tuple with the Quantity field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutModifier) GetQuantityOk() (*int32, bool) {
	if o == nil || IsNil(o.Quantity) {
		return nil, false
	}
	return o.Quantity, true
}

// HasQuantity returns a boolean if a field has been set.
func (o *TakeoutModifier) HasQuantity() bool {
	if o != nil && !IsNil(o.Quantity) {
		return true
	}

	return false
}

// ============================================================================
// TakeoutTakeoutOrderPrice - Aligned with Grab SDK TakeoutOrderPrice
// Reference: github.com/grab/grabfood-api-sdk-go@v1.0.2/model_order_price.go
// ============================================================================

// TakeoutTakeoutOrderPrice order price information in minor unit format
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

	//Grab SDK 尚未实现，但 API 文档有提及的字段。暂时新增用于解析 API 响应。
	Total *int64 `json:"total,omitempty"`
}

// GetSubtotal returns the Subtotal field value
func (o *TakeoutOrderPrice) GetSubtotal() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.Subtotal
}

// GetSubtotalOk returns a tuple with the Subtotal field value
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetSubtotalOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Subtotal, true
}

// SetSubtotal sets field value
func (o *TakeoutOrderPrice) SetSubtotal(v int64) {
	o.Subtotal = v
}

// GetTax returns the Tax field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetTax() int64 {
	if o == nil || IsNil(o.Tax) {
		var ret int64
		return ret
	}
	return *o.Tax
}

// GetTaxOk returns a tuple with the Tax field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetTaxOk() (*int64, bool) {
	if o == nil || IsNil(o.Tax) {
		return nil, false
	}
	return o.Tax, true
}

// HasTax returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasTax() bool {
	if o != nil && !IsNil(o.Tax) {
		return true
	}

	return false
}

// SetTax gets a reference to the given int64 and assigns it to the Tax field.
func (o *TakeoutOrderPrice) SetTax(v int64) {
	o.Tax = &v
}

// GetMerchantChargeFee returns the MerchantChargeFee field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetMerchantChargeFee() int64 {
	if o == nil || IsNil(o.MerchantChargeFee) {
		var ret int64
		return ret
	}
	return *o.MerchantChargeFee
}

// GetMerchantChargeFeeOk returns a tuple with the MerchantChargeFee field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetMerchantChargeFeeOk() (*int64, bool) {
	if o == nil || IsNil(o.MerchantChargeFee) {
		return nil, false
	}
	return o.MerchantChargeFee, true
}

// HasMerchantChargeFee returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasMerchantChargeFee() bool {
	if o != nil && !IsNil(o.MerchantChargeFee) {
		return true
	}

	return false
}

// SetMerchantChargeFee gets a reference to the given int64 and assigns it to the MerchantChargeFee field.
func (o *TakeoutOrderPrice) SetMerchantChargeFee(v int64) {
	o.MerchantChargeFee = &v
}

// GetGrabFundPromo returns the GrabFundPromo field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetGrabFundPromo() int64 {
	if o == nil || IsNil(o.GrabFundPromo) {
		var ret int64
		return ret
	}
	return *o.GrabFundPromo
}

// GetGrabFundPromoOk returns a tuple with the GrabFundPromo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetGrabFundPromoOk() (*int64, bool) {
	if o == nil || IsNil(o.GrabFundPromo) {
		return nil, false
	}
	return o.GrabFundPromo, true
}

// HasGrabFundPromo returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasGrabFundPromo() bool {
	if o != nil && !IsNil(o.GrabFundPromo) {
		return true
	}

	return false
}

// SetGrabFundPromo gets a reference to the given int64 and assigns it to the GrabFundPromo field.
func (o *TakeoutOrderPrice) SetGrabFundPromo(v int64) {
	o.GrabFundPromo = &v
}

// GetMerchantFundPromo returns the MerchantFundPromo field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetMerchantFundPromo() int64 {
	if o == nil || IsNil(o.MerchantFundPromo) {
		var ret int64
		return ret
	}
	return *o.MerchantFundPromo
}

// GetMerchantFundPromoOk returns a tuple with the MerchantFundPromo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetMerchantFundPromoOk() (*int64, bool) {
	if o == nil || IsNil(o.MerchantFundPromo) {
		return nil, false
	}
	return o.MerchantFundPromo, true
}

// HasMerchantFundPromo returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasMerchantFundPromo() bool {
	if o != nil && !IsNil(o.MerchantFundPromo) {
		return true
	}

	return false
}

// SetMerchantFundPromo gets a reference to the given int64 and assigns it to the MerchantFundPromo field.
func (o *TakeoutOrderPrice) SetMerchantFundPromo(v int64) {
	o.MerchantFundPromo = &v
}

// GetBasketPromo returns the BasketPromo field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetBasketPromo() int64 {
	if o == nil || IsNil(o.BasketPromo) {
		var ret int64
		return ret
	}
	return *o.BasketPromo
}

// GetBasketPromoOk returns a tuple with the BasketPromo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetBasketPromoOk() (*int64, bool) {
	if o == nil || IsNil(o.BasketPromo) {
		return nil, false
	}
	return o.BasketPromo, true
}

// HasBasketPromo returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasBasketPromo() bool {
	if o != nil && !IsNil(o.BasketPromo) {
		return true
	}

	return false
}

// SetBasketPromo gets a reference to the given int64 and assigns it to the BasketPromo field.
func (o *TakeoutOrderPrice) SetBasketPromo(v int64) {
	o.BasketPromo = &v
}

// GetDeliveryFee returns the DeliveryFee field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetDeliveryFee() int64 {
	if o == nil || IsNil(o.DeliveryFee) {
		var ret int64
		return ret
	}
	return *o.DeliveryFee
}

// GetDeliveryFeeOk returns a tuple with the DeliveryFee field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetDeliveryFeeOk() (*int64, bool) {
	if o == nil || IsNil(o.DeliveryFee) {
		return nil, false
	}
	return o.DeliveryFee, true
}

// HasDeliveryFee returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasDeliveryFee() bool {
	if o != nil && !IsNil(o.DeliveryFee) {
		return true
	}

	return false
}

// SetDeliveryFee gets a reference to the given int64 and assigns it to the DeliveryFee field.
func (o *TakeoutOrderPrice) SetDeliveryFee(v int64) {
	o.DeliveryFee = &v
}

// GetSmallOrderFee returns the SmallOrderFee field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetSmallOrderFee() int64 {
	if o == nil || IsNil(o.SmallOrderFee) {
		var ret int64
		return ret
	}
	return *o.SmallOrderFee
}

// GetSmallOrderFeeOk returns a tuple with the SmallOrderFee field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetSmallOrderFeeOk() (*int64, bool) {
	if o == nil || IsNil(o.SmallOrderFee) {
		return nil, false
	}
	return o.SmallOrderFee, true
}

// HasSmallOrderFee returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasSmallOrderFee() bool {
	if o != nil && !IsNil(o.SmallOrderFee) {
		return true
	}

	return false
}

// SetSmallOrderFee gets a reference to the given int64 and assigns it to the SmallOrderFee field.
func (o *TakeoutOrderPrice) SetSmallOrderFee(v int64) {
	o.SmallOrderFee = &v
}

// GetEaterPayment returns the EaterPayment field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetEaterPayment() int64 {
	if o == nil || IsNil(o.EaterPayment) {
		var ret int64
		return ret
	}
	return *o.EaterPayment
}

// GetEaterPaymentOk returns a tuple with the EaterPayment field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetEaterPaymentOk() (*int64, bool) {
	if o == nil || IsNil(o.EaterPayment) {
		return nil, false
	}
	return o.EaterPayment, true
}

// GetTotal returns the Total field value if set, zero value otherwise.
func (o *TakeoutOrderPrice) GetTotal() int64 {
	if o == nil || IsNil(o.Total) {
		var ret int64
		return ret
	}
	return *o.Total
}

// GetTotalOk returns a tuple with the Total field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutOrderPrice) GetTotalOk() (*int64, bool) {
	if o == nil || IsNil(o.Total) {
		return nil, false
	}
	return o.Total, true
}

// HasEaterPayment returns a boolean if a field has been set.
func (o *TakeoutOrderPrice) HasEaterPayment() bool {
	if o != nil && !IsNil(o.EaterPayment) {
		return true
	}

	return false
}

// SetEaterPayment gets a reference to the given int64 and assigns it to the EaterPayment field.
func (o *TakeoutOrderPrice) SetEaterPayment(v int64) {
	o.EaterPayment = &v
}

func (o TakeoutOrderPrice) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TakeoutOrderPrice) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["subtotal"] = o.Subtotal
	if !IsNil(o.Tax) {
		toSerialize["tax"] = o.Tax
	}
	if !IsNil(o.MerchantChargeFee) {
		toSerialize["merchantChargeFee"] = o.MerchantChargeFee
	}
	if !IsNil(o.GrabFundPromo) {
		toSerialize["grabFundPromo"] = o.GrabFundPromo
	}
	if !IsNil(o.MerchantFundPromo) {
		toSerialize["merchantFundPromo"] = o.MerchantFundPromo
	}
	if !IsNil(o.BasketPromo) {
		toSerialize["basketPromo"] = o.BasketPromo
	}
	if !IsNil(o.DeliveryFee) {
		toSerialize["deliveryFee"] = o.DeliveryFee
	}
	if !IsNil(o.SmallOrderFee) {
		toSerialize["smallOrderFee"] = o.SmallOrderFee
	}
	if !IsNil(o.EaterPayment) {
		toSerialize["eaterPayment"] = o.EaterPayment
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
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
// TakeoutFeatureFlags - Aligned with Grab SDK TakeoutFeatureFlags
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

// GetOrderAcceptedType returns the OrderAcceptedType field value
func (o *TakeoutFeatureFlags) GetOrderAcceptedType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OrderAcceptedType
}

// GetOrderAcceptedTypeOk returns a tuple with the OrderAcceptedType field value
// and a boolean to check if the value has been set.
func (o *TakeoutFeatureFlags) GetOrderAcceptedTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OrderAcceptedType, true
}

// SetOrderAcceptedType sets field value
func (o *TakeoutFeatureFlags) SetOrderAcceptedType(v string) {
	o.OrderAcceptedType = v
}

// GetOrderType returns the OrderType field value
func (o *TakeoutFeatureFlags) GetOrderType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.OrderType
}

// GetOrderTypeOk returns a tuple with the OrderType field value
// and a boolean to check if the value has been set.
func (o *TakeoutFeatureFlags) GetOrderTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.OrderType, true
}

// SetOrderType sets field value
func (o *TakeoutFeatureFlags) SetOrderType(v string) {
	o.OrderType = v
}

// GetIsMexEditOrder returns the IsMexEditOrder field value if set, zero value otherwise.
func (o *TakeoutFeatureFlags) GetIsMexEditOrder() bool {
	if o == nil || IsNil(o.IsMexEditOrder) {
		var ret bool
		return ret
	}
	return *o.IsMexEditOrder
}

// GetIsMexEditOrderOk returns a tuple with the IsMexEditOrder field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *TakeoutFeatureFlags) GetIsMexEditOrderOk() (*bool, bool) {
	if o == nil || IsNil(o.IsMexEditOrder) {
		return nil, false
	}
	return o.IsMexEditOrder, true
}

// HasIsMexEditOrder returns a boolean if a field has been set.
func (o *TakeoutFeatureFlags) HasIsMexEditOrder() bool {
	if o != nil && !IsNil(o.IsMexEditOrder) {
		return true
	}

	return false
}

// SetIsMexEditOrder gets a reference to the given bool and assigns it to the IsMexEditOrder field.
func (o *TakeoutFeatureFlags) SetIsMexEditOrder(v bool) {
	o.IsMexEditOrder = &v
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

// IsNil checks if an input is nil
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	case reflect.Array:
		return reflect.ValueOf(i).IsZero()
	}
	return false
}
