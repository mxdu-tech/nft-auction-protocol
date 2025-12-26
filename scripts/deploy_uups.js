const { ethers, upgrades, network } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const signers = await ethers.getSigners();
  console.log("signers:", signers.length);
  console.log("PRIVATE_KEY exists:", !!process.env.PRIVATE_KEY);
  console.log("RPC:", network.name);

  // ===== Sepolia 固定地址 =====
  const ETH_USD = "0x694AA1769357215DE4FAC081bf1f309aDC325306";
  const USDC = "0x07865c6E87B9F70255377e024ace6630C1Eaa37F";
  const USDC_USD = "0xA2F78ab2355fe2f984D808B5CeE7FD0A93D5270E";

  const [deployer] = await ethers.getSigners();

  const feeRecipient = deployer.address;
  const feeBps = 200; // 2%

  console.log("Deploying with:", deployer.address);
  console.log("Network:", network.name);

  const Factory = await ethers.getContractFactory(
    "NFTAuctionMarketUUPS_V1"
  );

  const proxy = await upgrades.deployProxy(
    Factory,
    [ETH_USD, USDC, USDC_USD, feeRecipient, feeBps],
    { kind: "uups" }
  );

  await proxy.waitForDeployment();
  const proxyAddress = await proxy.getAddress();

  console.log("Proxy deployed at:", proxyAddress);

  // ===== 写入 deployments =====
  const deploymentsDir = path.join(__dirname, "..", "deployments");
  if (!fs.existsSync(deploymentsDir)) fs.mkdirSync(deploymentsDir);

  const file = path.join(deploymentsDir, `${network.name}.json`);
  fs.writeFileSync(
    file,
    JSON.stringify(
      {
        proxy: proxyAddress,
        ethUsd: ETH_USD,
        usdc: USDC,
        usdcUsd: USDC_USD,
        feeRecipient,
        feeBps,
        timestamp: Date.now(),
      },
      null,
      2
    )
  );

  console.log(`Saved to deployments/${network.name}.json`);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});