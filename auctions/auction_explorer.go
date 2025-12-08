package auctions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	routeUtils "skyblock-pv-backend/routes/utils"
	"skyblock-pv-backend/utils"
	"skyblock-pv-backend/utils/nbt"
	"slices"
	"strings"
	"time"
)

const AuthCacheVersion = 1

func withCacheVersion(str string) string { return fmt.Sprintf("%s_%d", str, AuthCacheVersion) }

type opMode interface {
	GetAuctions(ctx *routeUtils.RouteContext) ([]AuctionStruct, error)
	Add(ctx *routeUtils.RouteContext, respond *AuctionRespond)
	Finish(ctx *routeUtils.RouteContext)
	Debug(page int)
}

type prod struct {
	opMode
	auctions []AuctionStruct
}

func (prod *prod) Add(_ *routeUtils.RouteContext, response *AuctionRespond) {
	if prod.auctions == nil {
		prod.auctions = make([]AuctionStruct, 0)
	}
	prod.auctions = append(prod.auctions, response.Auctions...)
}

func (prod *prod) Finish(_ *routeUtils.RouteContext) {}

func (prod *prod) GetAuctions(ctx *routeUtils.RouteContext) ([]AuctionStruct, error) {
	err := fetch(*ctx, prod)
	if err != nil {
		return nil, err
	}
	return prod.auctions, nil
}

func (prod *prod) Debug(_ int) {}

type dev struct {
	opMode
	auctions []AuctionStruct
	Duration time.Duration
}

func (dev *dev) Add(ctx *routeUtils.RouteContext, response *AuctionRespond) {
	cacheAll(ctx, response)
	if dev.auctions == nil {
		dev.auctions = make([]AuctionStruct, 0)
	}
	dev.auctions = append(dev.auctions, response.Auctions...)
}

func (dev *dev) Finish(ctx *routeUtils.RouteContext) {
	_ = ctx.AddToCache(withCacheVersion("auctions"), "cached", "<3", dev.Duration)
}

func (dev *dev) GetAuctions(ctx *routeUtils.RouteContext) ([]AuctionStruct, error) {
	fmt.Println("Using dev mode, PLEASE DONT USE IN PROD :sob:")
	if ctx.IsCached(withCacheVersion("auctions"), "cached") {
		fmt.Println("Retrieving previously cached data")
		data, err := dev.readCached(ctx)
		if err != nil {
			return nil, err
		}
		return *data, nil
	}

	err := fetch(*ctx, dev)
	if err != nil {
		return nil, err
	}
	return dev.auctions, nil
}

func (dev *dev) Debug(page int) {
	fmt.Printf("Fetching page %d\n", page)
}

func (dev *dev) readCached(ctx *routeUtils.RouteContext) (*[]AuctionStruct, error) {
	auctions, err := ctx.GetAll(withCacheVersion("auctions.index"))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loading %d auctions from cache\n", len(auctions))
	dev.auctions = make([]AuctionStruct, 0)
	for _, auctionJson := range auctions {
		var auction AuctionStruct
		err = json.Unmarshal([]byte(auctionJson), &auction)
		if err != nil {
			return nil, err
		}
		dev.auctions = append(dev.auctions, auction)
	}
	return &dev.auctions, nil
}

func GetCachedAuctions(ctx *routeUtils.RouteContext) (*string, error) {
	data, err := ctx.GetFromCache(nil, withCacheVersion("auctions"), "cached")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func FetchAll(ctx *routeUtils.RouteContext) error {
	var opMode opMode
	if utils.Debug {
		opMode = &dev{}
	} else {
		opMode = &prod{}
	}

	auctions, err := opMode.GetAuctions(ctx)
	if err != nil {
		return err
	}

	items := calculateAverage(auctions)
	data, err := json.Marshal(*items)
	if err != nil {
		return err
	}

	_ = ctx.AddToCache(withCacheVersion("auctions"), "cached", data, time.Hour*2)

	fmt.Println("Finished updating!")
	return nil
}

func calculateAverage(auctions []AuctionStruct) *map[string]ItemInfo {
	var items = make(map[string][]int64)

	fmt.Printf("Searching %d auctions\n", len(auctions))
	for _, auction := range auctions {
		if !auction.Bin {
			continue
		}
		item, err := auction.GetItem()
		if err != nil {
			continue
		}
		sbId := item.GetSbId()
		if sbId == nil {
			continue
		}
		priceList := items[*sbId]
		if priceList == nil {
			priceList = make([]int64, 0)
		}

		items[*sbId] = append(priceList, auction.StartingBid/int64(item.Count()))
	}

	actualItems := make(map[string]ItemInfo)
	for key, cost := range items {

		slices.Sort(cost)
		actualItems[key] = ItemInfo{
			Highest: cost[len(cost)-1],
			Lowest:  cost[0],
			Median:  calculateMedian(cost),
			Mean:    calculateArithmeticMean(cost),
		}
	}

	return &actualItems
}

func calculateArithmeticMean(list []int64) float64 {
	var sum float64 = 0
	for _, i := range list {
		sum += float64(i)
	}
	return sum / float64(len(list))
}

func calculateMedian(list []int64) int64 {
	if len(list) == 0 {
		return 0
	} else if len(list) <= 2 {
		return list[0]
	} else if len(list)%2 == 0 {
		middle := len(list) / 2
		return (list[middle] + list[middle+1]) / 2
	} else {
		return list[len(list)/2]
	}
}

type ItemInfo struct {
	Lowest  int64   `json:"lowest"`
	Highest int64   `json:"highest"`
	Median  int64   `json:"median"`
	Mean    float64 `json:"mean"`
}

func cacheAll(ctx *routeUtils.RouteContext, auction *AuctionRespond) {
	for _, auctionStruct := range auction.Auctions {
		cache(*ctx, &auctionStruct)
	}
}

func cache(ctx routeUtils.RouteContext, auction *AuctionStruct) {
	data, err := json.Marshal(auction)
	if err != nil {
		println(err.Error())
		return
	}
	_ = ctx.AddToCache(withCacheVersion("auctions.index"), auction.Id, data, time.Hour*7)
}

func fetch(ctx routeUtils.RouteContext, mode opMode) error {
	data, err := fetchPage(ctx, 0, &mode)
	if err != nil {
		return err
	}
	mode.Add(&ctx, data)

	for i := range data.TotalPages {
		if i == 0 {
			continue
		}
		data, err = fetchPage(ctx, i, &mode)
		if err != nil {
			fmt.Printf("so sad^%d", i)
			return err
		}
		mode.Add(&ctx, data)
	}
	mode.Finish(&ctx)
	return err
}

func fetchPage(ctx routeUtils.RouteContext, page int, mode *opMode) (*AuctionRespond, error) {
	(*mode).Debug(page)
	res, err := routeUtils.GetFromHypixel(ctx, fmt.Sprintf("/v2/skyblock/auctions?page=%d", page), false)
	if err != nil {
		println("Error fetching data from hypixel")
		return nil, err
	} else if res == nil {
		println("No data received from hypixel")
		return nil, fmt.Errorf("no data received from hypixel")
	}

	var auctionRespond AuctionRespond
	err = json.NewDecoder(strings.NewReader(*res)).Decode(&auctionRespond)
	if err != nil {
		println("Error decoding data from hypixel")
		return nil, err
	}

	return &auctionRespond, nil
}

type AuctionRespond struct {
	Success       bool            `json:"success"`
	Page          int             `json:"page"`
	TotalPages    int             `json:"totalPages"`
	TotalAuctions int             `json:"totalAuctions"`
	LastUpdated   int64           `json:"lastUpdated"`
	Auctions      []AuctionStruct `json:"auctions"`
}

type AuctionStruct struct {
	Bin         bool   `json:"bin"`
	Id          string `json:"uuid"`
	StartingBid int64  `json:"starting_bid"`
	ItemName    string `json:"item_name"`
	ItemLore    string `json:"item_lore"`
	ItemBytes   string `json:"item_bytes"`
}

func (auction *AuctionStruct) GetItem() (*utils.Item, error) {
	data, err := base64.StdEncoding.DecodeString(auction.ItemBytes)
	if err != nil {
		return nil, err
	}
	tag, err := nbt.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &utils.Item{Compound: tag.AsCompound().Get("i").AsList().GetValues()[0].(*nbt.Compound)}, nil
}
