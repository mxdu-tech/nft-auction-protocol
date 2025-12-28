package indexer

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const auctionABIJSON = `[
  {"anonymous":false,"inputs":[
    {"indexed":true,"internalType":"uint256","name":"auctionId","type":"uint256"},
    {"indexed":true,"internalType":"address","name":"seller","type":"address"},
    {"indexed":true,"internalType":"address","name":"nft","type":"address"},
    {"indexed":false,"internalType":"uint256","name":"tokenId","type":"uint256"},
    {"indexed":false,"internalType":"uint256","name":"endTime","type":"uint256"},
    {"indexed":false,"internalType":"uint256","name":"minBidUsd18","type":"uint256"}
  ],"name":"AuctionCreated","type":"event"},

  {"anonymous":false,"inputs":[
    {"indexed":true,"internalType":"uint256","name":"auctionId","type":"uint256"},
    {"indexed":true,"internalType":"address","name":"bidder","type":"address"},
    {"indexed":false,"internalType":"address","name":"bidToken","type":"address"},
    {"indexed":false,"internalType":"uint256","name":"bidAmount","type":"uint256"},
    {"indexed":false,"internalType":"uint256","name":"bidUsd18","type":"uint256"}
  ],"name":"BidPlaced","type":"event"},

  {"anonymous":false,"inputs":[
    {"indexed":true,"internalType":"uint256","name":"auctionId","type":"uint256"},
    {"indexed":false,"internalType":"address","name":"winner","type":"address"},
    {"indexed":false,"internalType":"address","name":"payToken","type":"address"},
    {"indexed":false,"internalType":"uint256","name":"payAmount","type":"uint256"},
    {"indexed":false,"internalType":"uint256","name":"payUsd18","type":"uint256"}
  ],"name":"AuctionEnded","type":"event"},

  {"anonymous":false,"inputs":[
    {"indexed":true,"internalType":"uint256","name":"auctionId","type":"uint256"}
  ],"name":"AuctionCancelled","type":"event"}
]`

func mustParseABI() abi.ABI {
	ab, err := abi.JSON(strings.NewReader(auctionABIJSON))
	if err != nil {
		panic(err)
	}
	return ab
}
