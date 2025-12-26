

# NFT Auction Market（UUPS 可升级合约）

## 一、项目概述

本项目实现了一个 **支持 NFT 拍卖的去中心化市场合约**，具备以下核心能力：

- 支持 **ETH 与 ERC20（USDC）** 两种方式竞拍 NFT  
- 使用 **Chainlink Price Feed（Oracle）** 将不同币种的出价统一换算为 USD 价格，保证竞拍公平性  
- 使用 **UUPS Proxy 模式** 实现合约升级，升级后 Proxy 地址保持不变  
- 完整的本地测试（Hardhat）与 Sepolia 测试网真实部署与验证  
- 覆盖 NFT 转移、竞拍、结算、升级等完整流程  

项目结构清晰，符合 Solidity 最佳实践，功能与测试完整。

---

## 二、项目结构

```text
final-nft-auction/
├── contracts/
│   ├── MyNFT.sol                 
│   ├── NFTAuctionMarketUUPS_V1.sol    
│   └── NFTAuctionMarketUUPS_V2.sol    
│
├── scripts/
│   ├── deploy_mynft.js                # 部署 NFT 合约
│   ├── deploy_uups.js                 # 部署 UUPS Proxy（V1）
│   └── upgrade_uups.js                # 升级到 V2
│
├── test/
│   └── NFTAuctionMarket.test.js       # 拍卖功能单元测试
│
├── hardhat.config.js
├── package.json
└── README.md
```

## 三、合约说明

### 1. MyNFT.sol

一个基于 OpenZeppelin 的 ERC721 NFT 合约，用于拍卖市场测试。

**功能说明：**

- 支持 NFT 铸造（mint）
- 每个 NFT 由唯一的 `tokenId` 标识
- NFT 所有权完全遵循 ERC721 标准

**核心函数：**

- `mint(address to)`
- `ownerOf(uint256 tokenId)`
- `approve(address to, uint256 tokenId)`

---

### 2. NFTAuctionMarketUUPS_V1.sol

拍卖市场的核心实现，采用 **UUPS Upgradeable Proxy** 架构。

**主要功能：**

- 创建 NFT 拍卖
- 支持 ETH 与 USDC 两种竞价方式
- 使用 Chainlink 价格预言机将出价统一换算为 USD
- 所有竞价统一以 USD 价格进行比较
- 自动拒绝低于最低价或当前最高价的出价
- 拍卖结束后自动结算 NFT 与资金

**核心函数：**

- `createAuction(address nft, uint256 tokenId, uint256 durationSeconds, uint256 minBidUsd18)`
- `bidEth(uint256 auctionId)`
- `bidUSDC(uint256 auctionId, uint256 usdcAmount)`
- `endAuction(uint256 auctionId)`
- `getAuction(uint256 auctionId)`

**安全设计：**

- 使用 `nonReentrant` 防止重入攻击
- 使用 `OwnableUpgradeable` 控制升级权限
- 使用 `UUPSUpgradeable` 实现安全升级

---

### 3. NFTAuctionMarketUUPS_V2.sol

在 **不改变任何存储结构** 的前提下，用于验证合约升级能力。

**新增内容：**

```solidity
function version() external pure returns (uint256) {
    return 2;
}
```

## 四、价格计算机制（Chainlink Oracle）

### 使用的价格源（Sepolia）

- ETH / USD Feed  
  地址：0x694AA1769357215DE4FAC081bf1f309aDC325306

- USDC / USD Feed  
  地址：0xA2F78ab2355fe2f984D808B5CeE7FD0A93D5270E

### 价格换算逻辑

ETH 出价换算方式：

    ETH 出价数量 × ETH/USD 价格 = USD（18 位精度）

USDC 出价换算方式：

    USDC 出价数量 × USDC/USD 价格 = USD（18 位精度）

拍卖合约内部会将所有出价统一换算为 USD（18 位精度）后再进行比较，
从而保证不同币种竞价时的公平性。

---

## 五、测试说明

### 1. 本地测试（Hardhat）

测试文件路径：

    test/NFTAuctionMarket.test.js

本地测试覆盖内容包括：

- NFT 上架拍卖流程
- ETH 出价与 USDC 出价的 USD 价格比较
- 低于最低出价的竞拍失败
- 更高出价成功，较低出价失败
- 拍卖结束后的状态校验
- NFT 与资金的正确结算
- UUPS 合约升级功能验证

所有本地测试均通过。

---

### 2. 测试网真实测试（Sepolia）

在 Sepolia 测试网完成了完整的真实流程验证：

- 部署 MyNFT 合约
- 铸造 NFT 到用户钱包地址
- Approve NFT 给拍卖市场合约
- 创建拍卖（设置最低 USD 价格与拍卖持续时间）
- 多次 ETH 出价测试（成功与失败场景）
- 验证 Chainlink Oracle 返回的 ETH/USD 价格
- 拍卖结束后的状态检查
- NFT 所有权正确转移
- 成功升级到 V2 合约并验证升级结果

---

## 六、部署信息（Sepolia）

拍卖市场 Proxy 合约地址（始终不变）：

    0x085Ab5880dff3EDaE319948946E1df5bF683934e

V2 Implementation 合约地址：

    0x6E5F7Ba248F6deaC57ba0cfB565C0c7f24240f94

MyNFT 合约地址：

    0x7613Fea95c816346beaBE7132A29BC4Ab14BE83b

---

## 七、部署与升级步骤

### 1. 部署 MyNFT 合约

执行命令：

    npx hardhat run scripts/deploy_mynft.js --network sepolia

### 2. 部署拍卖市场（UUPS Proxy，V1）

执行命令：

    npx hardhat run scripts/deploy_uups.js --network sepolia

### 3. 升级拍卖市场到 V2

执行命令：

    npx hardhat run scripts/upgrade_uups.js --network sepolia

升级完成后，Proxy 地址保持不变，
可通过调用 version() 函数验证升级是否成功。

---

## 八、测试截图

### 1. NFT 合约部署与铸造

![NFT 合约部署与铸造](./screenshots/nft_contract.png)

### 2. NFT 拍卖市场功能测试

![NFT 拍卖市场测试](./screenshots/NFTAuctionMarket.png)

### 3. UUPS 合约升级验证

![合约升级成功](./screenshots/upgrade.png)


---

## 九、总结

本项目完整实现了以下功能：

- NFT 拍卖市场
- ETH / ERC20（USDC）混合竞拍
- Chainlink Oracle 价格统一换算
- 公平的 USD 价格比较机制
- UUPS 可升级合约架构
- 本地测试与测试网真实验证

