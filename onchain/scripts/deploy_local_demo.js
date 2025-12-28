const { ethers, upgrades, network } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log("Deployer:", deployer.address);
  console.log("Network:", network.name);

  // ===== 1. 部署 Mock NFT =====
  // 部署 MyNFT（无参数）
    const NFTFactory = await ethers.getContractFactory("MyNFT");
    const nft = await NFTFactory.deploy();
    await nft.waitForDeployment();

    const nftAddress = await nft.getAddress();
    console.log("MyNFT deployed at:", nftAddress);

    // mint 一个 NFT（tokenId = 1）
    await (await nft.mint(deployer.address)).wait();
    console.log("NFT minted to deployer");

  // ===== 2. 本地链不用真实预言机，随便占位 =====
  const dummy = deployer.address;

  const feeRecipient = deployer.address;
  const feeBps = 200;

  // ===== 3. 部署 Auction Market（UUPS）=====
  const MarketFactory = await ethers.getContractFactory(
    "NFTAuctionMarketUUPS_V1"
  );

  const market = await upgrades.deployProxy(
    MarketFactory,
    [dummy, dummy, dummy, feeRecipient, feeBps],
    { kind: "uups" }
  );

  await market.waitForDeployment();
  const marketAddress = await market.getAddress();

  console.log("Auction Market deployed at:", marketAddress);

  // ===== 4. 写 deployments =====
  const deploymentsDir = path.join(__dirname, "..", "deployments");
  if (!fs.existsSync(deploymentsDir)) fs.mkdirSync(deploymentsDir);

  const file = path.join(deploymentsDir, "localhost.json");
  fs.writeFileSync(
    file,
    JSON.stringify(
      {
        network: "localhost",
        nft: nftAddress,
        auction: marketAddress,
        deployer: deployer.address,
        timestamp: Date.now(),
      },
      null,
      2
    )
  );

  console.log("Saved deployments/localhost.json");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});