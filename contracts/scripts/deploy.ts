import { ethers } from "hardhat";

async function main() {
  const tokenName = "0g-chain-wrapped ATOM";
  const tokenSymbol = "kATOM";
  const tokenDecimals = 6;

  const ERC200gChainWrappedCosmosCoin = await ethers.getContractFactory(
    "ERC200gChainWrappedCosmosCoin"
  );
  const token = await ERC200gChainWrappedCosmosCoin.deploy(
    tokenName,
    tokenSymbol,
    tokenDecimals
  );

  await token.deployed();

  console.log(
    `Token "${tokenName}" (${tokenSymbol}) with ${tokenDecimals} decimals is deployed to ${token.address}!`
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
