// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./NFTAuctionMarketUUPS_V1.sol";

contract NFTAuctionMarketUUPS_V2 is NFTAuctionMarketUUPS_V1 {
    function version() external pure returns (uint256) {
        return 2;
    }
}