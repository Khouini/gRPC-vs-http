package types

// Hotel represents a hotel in the system
type Hotel struct {
	SupplierId               *int32             `json:"supplierId,omitempty"`
	SupplierIds              []int32            `json:"supplierIds,omitempty"`
	HotelId                  *string            `json:"hotelId,omitempty"`
	HotelIds                 []string           `json:"hotelIds,omitempty"`
	GiataId                  *int32             `json:"giataId,omitempty"`
	HUid                     *int32             `json:"hUid,omitempty"`
	Name                     *string            `json:"name,omitempty"`
	Rating                   *float32           `json:"rating,omitempty"`
	Address                  *string            `json:"address,omitempty"`
	Score                    *float64           `json:"score,omitempty"`
	HotelChainId             *int32             `json:"hotelChainId,omitempty"`
	AccTypeId                *int32             `json:"accTypeId,omitempty"`
	City                     *string            `json:"city,omitempty"`
	CityId                   *int32             `json:"cityId,omitempty"`
	ZoneId                   int32              `json:"zoneId"`
	Zone                     string             `json:"zone"`
	Country                  *string            `json:"country,omitempty"`
	CountryCode              *string            `json:"countryCode,omitempty"`
	CountryId                *int32             `json:"countryId,omitempty"`
	Lat                      *float64           `json:"lat,omitempty"`
	Long                     *float64           `json:"long,omitempty"`
	MarketingText            *string            `json:"marketingText,omitempty"`
	MinRate                  *float64           `json:"minRate,omitempty"`
	MaxRate                  *float64           `json:"maxRate,omitempty"`
	Currency                 *string            `json:"currency,omitempty"`
	Photos                   []string           `json:"photos,omitempty"`
	Rooms                    []Room             `json:"rooms,omitempty"`
	Supplements              []Supplement       `json:"supplements,omitempty"`
	Total                    *float32           `json:"total,omitempty"`
	Distances                map[string]float32 `json:"distances,omitempty"`
	Neighborhood             *Neighborhood      `json:"neighborhood,omitempty"`
	Strength                 map[string]bool    `json:"strength,omitempty"`
	Review                   *Review            `json:"review,omitempty"`
	Available                *bool              `json:"available,omitempty"`
	Boards                   []string           `json:"boards,omitempty"`
	Tag                      *string            `json:"tag,omitempty"`
	CityLat                  *float64           `json:"cityLat,omitempty"`
	CityLong                 *float64           `json:"cityLong,omitempty"`
	Reviews                  []HotelReview      `json:"reviews,omitempty"`
	ReviewsSubratingsAverage map[string]float32 `json:"reviewsSubratingsAverage,omitempty"`
	AllNRF                   *bool              `json:"allNRF,omitempty"`
	AllRF                    *bool              `json:"allRF,omitempty"`
	PartialNRF               *bool              `json:"partialNRF,omitempty"`
}

type Room struct {
	Code         *string  `json:"code,omitempty"`
	Codes        []string `json:"codes,omitempty"`
	Name         *string  `json:"name,omitempty"`
	Names        []string `json:"names,omitempty"`
	Rates        []Rate   `json:"rates,omitempty"`
	Category     *string  `json:"category,omitempty"`
	Total        *float64 `json:"total,omitempty"`
	OriginalCode *string  `json:"originalCode,omitempty"`
	OriginalName *string  `json:"originalName,omitempty"`
}

type Rate struct {
	RateKey              *string              `json:"rateKey,omitempty"`
	RateClass            *string              `json:"rateClass,omitempty"`
	ContractId           *int32               `json:"contractId,omitempty"`
	RateType             *string              `json:"rateType,omitempty"`
	PaymentType          *string              `json:"paymentType,omitempty"`
	Allotment            *int32               `json:"allotment,omitempty"`
	Availability         *string              `json:"availability,omitempty"`
	Amount               *float64             `json:"amount,omitempty"`
	Currency             *string              `json:"currency,omitempty"`
	BoardCode            *string              `json:"boardCode,omitempty"`
	BoardName            *string              `json:"boardName,omitempty"`
	Nrf                  *bool                `json:"nrf,omitempty"`
	CancellationPolicies []CancellationPolicy `json:"cancellationPolicies,omitempty"`
	Offers               []Offer              `json:"offers,omitempty"`
	Promotions           []Promotion          `json:"promotions,omitempty"`
	Supplements          []Supplement         `json:"supplements,omitempty"`
	Taxes                []Tax                `json:"taxes,omitempty"`
	Rooms                *int32               `json:"rooms,omitempty"`
	Adults               *string              `json:"adults,omitempty"`
	Children             *string              `json:"children,omitempty"`
	Infant               *string              `json:"infant,omitempty"`
	ChildrenAges         *string              `json:"childrenAges,omitempty"`
	RateComments         *string              `json:"rateComments,omitempty"`
	Packaging            *bool                `json:"packaging,omitempty"`
	Total                *float64             `json:"total,omitempty"`
	PurchasePrice        *float64             `json:"purchasePrice,omitempty"`
}

type CancellationPolicy struct {
	Amount        *float64 `json:"amount,omitempty"`
	From          *string  `json:"from,omitempty"`
	RealFrom      *string  `json:"realFrom,omitempty"`
	Name          *string  `json:"name,omitempty"`
	PurchasePrice *float64 `json:"purchasePrice,omitempty"`
}

type Offer struct {
	Amount *float64 `json:"amount,omitempty"`
	Code   *string  `json:"code,omitempty"`
	Name   *string  `json:"name,omitempty"`
}

type Promotion struct {
	Remark *string `json:"remark,omitempty"`
	Name   *string `json:"name,omitempty"`
	Code   *string `json:"code,omitempty"`
}

type Supplement struct {
	Name     *string  `json:"name,omitempty"`
	Amount   *float64 `json:"amount,omitempty"`
	Currency *string  `json:"currency,omitempty"`
	Included *bool    `json:"included,omitempty"`
}

type Tax struct {
	Name     *string  `json:"name,omitempty"`
	Amount   *float64 `json:"amount,omitempty"`
	Currency *string  `json:"currency,omitempty"`
	Included *bool    `json:"included,omitempty"`
	Type     *string  `json:"type,omitempty"`
}

type Neighborhood struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Review struct {
	Score   float64 `json:"score"`
	Count   int32   `json:"count"`
	Average float64 `json:"average"`
}

type HotelReview struct {
	Id         string             `json:"id"`
	Rating     float32            `json:"rating"`
	Comment    string             `json:"comment"`
	Author     string             `json:"author"`
	Date       string             `json:"date"`
	Subratings map[string]float32 `json:"subratings"`
}

// Metadata contains information about the dataset
type Metadata struct {
	GeneratedAt  string  `json:"generatedAt"`
	TotalHotels  int     `json:"totalHotels"`
	GeneratedBy  string  `json:"generatedBy"`
	ActualSizeMB float64 `json:"actualSizeMB"`
	ActualHotels int     `json:"actualHotels"`
}

// DataFile represents the structure of the data.json file
type DataFile struct {
	Metadata Metadata `json:"metadata"`
	Hotels   []Hotel  `json:"hotels"`
}

// StatsResponse represents the response from the gateway
type StatsResponse struct {
	TotalHotels     int     `json:"totalHotels"`
	AvailableHotels int     `json:"availableHotels"`
	DataSize        float64 `json:"dataSize"`
	ProcessTimeMs   int64   `json:"processTimeMs"`
}
