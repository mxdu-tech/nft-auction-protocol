const { expect } = require("chai");
const { ethers, upgrades } = require("hardhat");

describe("NFTAuctionMarket (UUPS)", function () {
  let owner, seller, bidder1, bidder2;
  let nft, market, usdc, ethFeed, usdcFeed;

  beforeEach(async function () {
    [owner, seller, bidder1, bidder2] = await ethers.getSigners();

    // Deploy mocks
    const MockV3 = await ethers.getContractFactory("MockV3Aggregator");
    // ETH/USD = 3000 * 1e8
    ethFeed = await MockV3.deploy(8, 3000n * 10n ** 8n);
    await ethFeed.waitForDeployment();
    // USDC/USD = 1 * 1e8
    usdcFeed = await MockV3.deploy(8, 1n * 10n ** 8n);
    await usdcFeed.waitForDeployment();

    const MockUSDC = await ethers.getContractFactory("MockUSDC");
    usdc = await MockUSDC.deploy();
    await usdc.waitForDeployment();

    // mint USDC to bidders
    await usdc.mint(bidder1.address, 10_000n * 10n ** 6n);
    await usdc.mint(bidder2.address, 10_000n * 10n ** 6n);

    // Deploy NFT
    const MyNFT = await ethers.getContractFactory("MyNFT");
    nft = await MyNFT.connect(seller).deploy();
    await nft.waitForDeployment();
    await nft.connect(seller).mint(seller.address); // tokenId=1

    // Deploy Market via UUPS proxy
    const MarketV1 = await ethers.getContractFactory("NFTAuctionMarketUUPS_V1");
    market = await upgrades.deployProxy(
      MarketV1,
      [await ethFeed.getAddress(), await usdc.getAddress(), await usdcFeed.getAddress()],
      { kind: "uups" }
    );
    await market.waitForDeployment();

    // approve NFT
    await nft.connect(seller).approve(await market.getAddress(), 1);
  });

  it("createAuction: should escrow NFT", async function () {
    const minBidUsd18 = ethers.parseUnits("100", 18);
    const tx = await market.connect(seller).createAuction(
      await nft.getAddress(),
      1,
      3600,
      minBidUsd18
    );
    await tx.wait();

    expect(await nft.ownerOf(1)).to.equal(await market.getAddress());
  });

  it("bidEth vs bidUSDC: compare by USD", async function () {
    const minBidUsd18 = ethers.parseUnits("100", 18);
    await market.connect(seller).createAuction(await nft.getAddress(), 1, 3600, minBidUsd18);

    // bidder1 bids 0.05 ETH => 0.05*3000=150 USD
    await market.connect(bidder1).bidEth(1, { value: ethers.parseEther("0.05") });

    // bidder2 bids 140 USDC => 140 USD, should fail (not higher)
    await usdc.connect(bidder2).approve(await market.getAddress(), 200n * 10n ** 6n);
    await expect(
      market.connect(bidder2).bidUSDC(1, 140n * 10n ** 6n)
    ).to.be.revertedWith("Bid not high enough");

    // bidder2 bids 200 USDC => 200 USD, should win
    await market.connect(bidder2).bidUSDC(1, 200n * 10n ** 6n);

    const a = await market.getAuction(1);
    expect(a.highestBidder).to.equal(bidder2.address);
    expect(a.highestBidToken).to.equal(await usdc.getAddress());
  });

  it("endAuction: should transfer NFT + pay seller in winning token", async function () {
    const minBidUsd18 = ethers.parseUnits("100", 18);
    await market.connect(seller).createAuction(await nft.getAddress(), 1, 10, minBidUsd18);

    // bid USDC 200
    await usdc.connect(bidder1).approve(await market.getAddress(), 200n * 10n ** 6n);
    await market.connect(bidder1).bidUSDC(1, 200n * 10n ** 6n);

    // time travel
    await ethers.provider.send("evm_increaseTime", [11]);
    await ethers.provider.send("evm_mine");

    const sellerUsdcBefore = await usdc.balanceOf(seller.address);
    await market.endAuction(1);

    expect(await nft.ownerOf(1)).to.equal(bidder1.address);
    const sellerUsdcAfter = await usdc.balanceOf(seller.address);
    expect(sellerUsdcAfter).to.be.gt(sellerUsdcBefore);
  });

  it("UUPS upgrade: should upgrade to V2 and call version()", async function () {
    const MarketV2 = await ethers.getContractFactory("NFTAuctionMarketUUPS_V2");
    const upgraded = await upgrades.upgradeProxy(await market.getAddress(), MarketV2);
    expect(await upgraded.version()).to.equal(2n);
  });
});