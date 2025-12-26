// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

import "./interfaces/AggregatorV3Interface.sol";

contract NFTAuctionMarketUUPS_V1 is
    Initializable,
    OwnableUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable
{
    // address(0) 表示 ETH
    address public constant NATIVE_TOKEN = address(0);

    struct Auction {
        address seller;
        address nft;
        uint256 tokenId;

        uint256 endTime;
        bool ended;

        // 最高出价（原始金额）
        address highestBidToken;   // 0=ETH, otherwise ERC20(USDC)
        uint256 highestBidAmount;

        // 最高出价（统一换算成 USD，18 位小数，用于比较）
        uint256 highestBidUsd18;
        address highestBidder;

        // 最低起拍价（USD 18位）
        uint256 minBidUsd18;
    }

    uint256 public nextAuctionId;
    mapping(uint256 => Auction) public auctions;

    // 只支持一种 ERC20：USDC
    IERC20 public usdc;

    // Chainlink feeds
    AggregatorV3Interface public ethUsdFeed;
    AggregatorV3Interface public usdcUsdFeed;
    address public feeRecipient;
    uint256 public feeBps; // e.g. 200 = 2%

    event AuctionCreated(
        uint256 indexed auctionId,
        address indexed seller,
        address indexed nft,
        uint256 tokenId,
        uint256 endTime,
        uint256 minBidUsd18
    );

    event BidPlaced(
        uint256 indexed auctionId,
        address indexed bidder,
        address bidToken,
        uint256 bidAmount,
        uint256 bidUsd18
    );

    event AuctionEnded(
        uint256 indexed auctionId,
        address winner,
        address payToken,
        uint256 payAmount,
        uint256 payUsd18
    );

    event AuctionCancelled(uint256 indexed auctionId);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address _ethUsdFeed,
        address _usdc,
        address _usdcUsdFeed,
        address _feeRecipient,
        uint256 _feeBps
    ) public initializer {
        __Ownable_init(msg.sender);
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        ethUsdFeed = AggregatorV3Interface(_ethUsdFeed);
        usdc = IERC20(_usdc);
        usdcUsdFeed = AggregatorV3Interface(_usdcUsdFeed);

        feeRecipient = _feeRecipient;
        feeBps = _feeBps;
    }

    function _authorizeUpgrade(address newImplementation)
        internal
        override
        onlyOwner
    {}

    /* ---------------- Oracle helpers (USD 18 decimals) ---------------- */

    function _scaleTo18(uint256 value, uint8 decimals_) internal pure returns (uint256) {
        if (decimals_ == 18) return value;
        if (decimals_ < 18) return value * (10 ** (18 - decimals_));
        return value / (10 ** (decimals_ - 18));
    }

    function _readFeed(AggregatorV3Interface feed) internal view returns (uint256 answer, uint8 dec) {
        (, int256 a,, uint256 updatedAt,) = feed.latestRoundData();
        require(a > 0, "Oracle: bad answer");
        require(updatedAt > 0, "Oracle: stale");
        answer = uint256(a);
        dec = feed.decimals();
    }

    // ETH(wei, 1e18) * (ETH/USD, e.g. 3000*1e8) => USD(1e18)
    function _ethToUsd18(uint256 weiAmount) internal view returns (uint256) {
        (uint256 price, uint8 dec) = _readFeed(ethUsdFeed); // usually dec=8
        uint256 usd = (weiAmount * price) / 1e18; // now has `dec` decimals
        return _scaleTo18(usd, dec);
    }

    // USDC(usually 1e6) * (USDC/USD ~ 1*1e8) => USD(1e18)
    function _usdcToUsd18(uint256 usdcAmount) internal view returns (uint256) {
        (uint256 price, uint8 feedDec) = _readFeed(usdcUsdFeed); // usually 8
        // usdc token decimals: assume 6
        uint256 usdWithFeedDec = (usdcAmount * price) / 1e6; // now has `feedDec` decimals
        return _scaleTo18(usdWithFeedDec, feedDec);
    }

    /* ---------------- Market core ---------------- */

    function createAuction(
        address nft,
        uint256 tokenId,
        uint256 durationSeconds,
        uint256 minBidUsd18
    ) external nonReentrant returns (uint256 auctionId) {
        require(durationSeconds > 0, "Bad duration");
        require(minBidUsd18 > 0, "Bad minBidUsd");

        auctionId = ++nextAuctionId;

        auctions[auctionId] = Auction({
            seller: msg.sender,
            nft: nft,
            tokenId: tokenId,
            endTime: block.timestamp + durationSeconds,
            ended: false,
            highestBidToken: NATIVE_TOKEN,
            highestBidAmount: 0,
            highestBidUsd18: 0,
            highestBidder: address(0),
            minBidUsd18: minBidUsd18
        });

        // 托管 NFT
        IERC721(nft).transferFrom(msg.sender, address(this), tokenId);

        emit AuctionCreated(auctionId, msg.sender, nft, tokenId, auctions[auctionId].endTime, minBidUsd18);
    }

    function bidEth(uint256 auctionId) external payable nonReentrant {
        Auction storage a = auctions[auctionId];
        require(a.seller != address(0), "No auction");
        require(!a.ended, "Ended");
        require(block.timestamp < a.endTime, "Auction over");
        require(msg.value > 0, "Zero bid");

        uint256 bidUsd18 = _ethToUsd18(msg.value);
        _placeBid(a, auctionId, msg.sender, NATIVE_TOKEN, msg.value, bidUsd18);
    }

    function bidUSDC(uint256 auctionId, uint256 usdcAmount) external nonReentrant {
        Auction storage a = auctions[auctionId];
        require(a.seller != address(0), "No auction");
        require(!a.ended, "Ended");
        require(block.timestamp < a.endTime, "Auction over");
        require(usdcAmount > 0, "Zero bid");

        uint256 bidUsd18 = _usdcToUsd18(usdcAmount);

        // 先把钱转进来（如果后面失败，整笔 tx revert，转账也回滚）
        require(usdc.transferFrom(msg.sender, address(this), usdcAmount), "USDC transferFrom failed");

        _placeBid(a, auctionId, msg.sender, address(usdc), usdcAmount, bidUsd18);
    }

    function _placeBid(
        Auction storage a,
        uint256 auctionId,
        address bidder,
        address bidToken,
        uint256 bidAmount,
        uint256 bidUsd18
    ) internal {
        require(bidUsd18 >= a.minBidUsd18, "Below minBidUsd");
        require(bidUsd18 > a.highestBidUsd18, "Bid not high enough");

        // 退还上一位最高出价者（按原币种原金额退）
        if (a.highestBidder != address(0)) {
            _refund(a.highestBidder, a.highestBidToken, a.highestBidAmount);
        }

        a.highestBidder = bidder;
        a.highestBidToken = bidToken;
        a.highestBidAmount = bidAmount;
        a.highestBidUsd18 = bidUsd18;

        emit BidPlaced(auctionId, bidder, bidToken, bidAmount, bidUsd18);
    }

    function _refund(address to, address token, uint256 amount) internal {
        if (amount == 0) return;
        if (token == NATIVE_TOKEN) {
            (bool ok,) = payable(to).call{value: amount}("");
            require(ok, "Refund ETH failed");
        } else {
            require(IERC20(token).transfer(to, amount), "Refund ERC20 failed");
        }
    }

    function endAuction(uint256 auctionId) external nonReentrant {
        Auction storage a = auctions[auctionId];
        require(a.seller != address(0), "No auction");
        require(!a.ended, "Already ended");
        require(block.timestamp >= a.endTime, "Not ended yet");

        a.ended = true;

        if (a.highestBidder == address(0)) {
            // 无人出价：退回 NFT
            IERC721(a.nft).transferFrom(address(this), a.seller, a.tokenId);
            emit AuctionCancelled(auctionId);
            return;
        }

        // NFT 给赢家
        IERC721(a.nft).transferFrom(address(this), a.highestBidder, a.tokenId);

        // 资金给卖家（赢家用什么付，就给什么）
        if (a.highestBidToken == NATIVE_TOKEN) {
            (bool ok,) = payable(a.seller).call{value: a.highestBidAmount}("");
            require(ok, "Pay seller ETH failed");
        } else {
            require(IERC20(a.highestBidToken).transfer(a.seller, a.highestBidAmount), "Pay seller ERC20 failed");
        }

        emit AuctionEnded(auctionId, a.highestBidder, a.highestBidToken, a.highestBidAmount, a.highestBidUsd18);
    }

    // 便于前端/测试读取
    function getAuction(uint256 auctionId) external view returns (Auction memory) {
        return auctions[auctionId];
    }

    receive() external payable {}
}