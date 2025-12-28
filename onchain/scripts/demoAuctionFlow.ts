import fs from "fs";
import path from "path";
import { ethers, network } from "hardhat";

async function main() {
  console.log("=== demoAuctionFlow start ===");

  /* ------------------------------------------------------------
   * 1. 读取部署信息
   * ------------------------------------------------------------ */
  

  const deploymentsPath = path.join(
    __dirname,
    `../deployments/${network.name}.json`
  );

  if (!fs.existsSync(deploymentsPath)) {
    throw new Error("deployments/${network.name}.json not found");
  }

  const deployments = JSON.parse(
    fs.readFileSync(deploymentsPath, "utf8")
  );

  const auctionAddress: string = deployments.auction;
  const nftAddress: string = deployments.nft;

  if (!auctionAddress || !nftAddress) {
    throw new Error("auction or nft address missing in deployments");
  }

  console.log("auction:", auctionAddress);
  console.log("nft:", nftAddress);

  /* ------------------------------------------------------------
   * 2. 获取账号
   * ------------------------------------------------------------ */
  // const [seller, bidder] = await ethers.getSigners();
  const provider = ethers.provider;

  // 卖家（用 PRIVATE_KEY）
  const seller = new ethers.Wallet(
    process.env.PRIVATE_KEY!,
    provider
  );

  // 买家（用 BIDDER_PRIVATE_KEY）
  const bidder = new ethers.Wallet(
    process.env.BIDDER_PRIVATE_KEY!,
    provider
  );
  console.log("seller:", seller.address);
  console.log("bidder:", bidder.address);

  /* ------------------------------------------------------------
   * 3. 合约实例
   * ------------------------------------------------------------ */
  const Auction = await ethers.getContractAt(
    "NFTAuctionMarketUUPS_V1",
    auctionAddress,
    seller
  );

  const NFT = await ethers.getContractAt(
    "MyNFT",
    nftAddress,
    seller
  );

  /* ------------------------------------------------------------
   * 4. mint NFT
   * ------------------------------------------------------------ */
  const mintTx = await NFT.mint(seller.address);
  await mintTx.wait();

  // MyNFT 是自增 tokenId，这里直接用 ownerOf 校验
  const tokenId = await NFT.nextTokenId();
  console.log("NFT minted:", tokenId.toString());

  /* ------------------------------------------------------------
   * 5. approve（推荐 approveForAll，避免 ERC721InvalidApprover）
   * ------------------------------------------------------------ */
  const approvedForAll = await NFT.isApprovedForAll(
    seller.address,
    auctionAddress
  );

  if (!approvedForAll) {
    await (await NFT.setApprovalForAll(auctionAddress, true)).wait();
  }

  console.log("NFT approval for all set");

  /* ------------------------------------------------------------
   * 6. create auction（关键：用 callStatic 拿 auctionId）
   * ------------------------------------------------------------ */
  const duration = 60; // seconds
  const minBidUsd18 = ethers.parseUnits("10", 18);

  // 6. create auction（ethers v6 正确方式）
  const auctionId = await Auction.createAuction.staticCall(
    nftAddress,
    tokenId,
    duration,
    minBidUsd18
  );

  await (
    await Auction.createAuction(
      nftAddress,
      tokenId,
      duration,
      minBidUsd18
    )
  ).wait();

  console.log("auction created, auctionId =", auctionId.toString());

  /* ------------------------------------------------------------
   * 7. 出价（ETH）
   * ------------------------------------------------------------ */
  await (
    await Auction.connect(bidder).bidEth(auctionId, {
      value: ethers.parseEther("0.01"),
    })
  ).wait();

  console.log("bid placed");

  /* ------------------------------------------------------------
   * 8. 快进时间
   * ------------------------------------------------------------ */
  if (network.name === "localhost" || network.name === "hardhat") {
    await ethers.provider.send("evm_increaseTime", [120]);
    await ethers.provider.send("evm_mine", []);
  } else {
    console.log("waiting for auction to end on public network...");
    await new Promise(resolve => setTimeout(resolve, 70_000));
  }


  /* ------------------------------------------------------------
   * 9. 结束拍卖
   * ------------------------------------------------------------ */
  await (
    await Auction.connect(bidder).endAuction(auctionId)
  ).wait();

  console.log("auction ended");
  console.log("=== demoAuctionFlow end ===");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});