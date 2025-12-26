const { ethers, upgrades, network } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const file = path.join(__dirname, "..", "deployments", `${network.name}.json`);
  if (!fs.existsSync(file)) {
    throw new Error(`Missing deployments/${network.name}.json`);
  }

  const { proxy } = JSON.parse(fs.readFileSync(file, "utf8"));
  if (!proxy) throw new Error("Proxy address missing");

  console.log("Upgrading proxy:", proxy);

  const FactoryV2 = await ethers.getContractFactory(
    "NFTAuctionMarketUUPS_V2"
  );

  const upgraded = await upgrades.upgradeProxy(proxy, FactoryV2);
  await upgraded.waitForDeployment();

  console.log("Upgrade done.");

  const impl = await upgrades.erc1967.getImplementationAddress(
    await upgraded.getAddress()
  );
  console.log("New implementation:", impl);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});