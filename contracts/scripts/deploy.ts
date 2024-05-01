import { ethers } from "hardhat";

async function main() {
  const tokenName = "0g-chain-wrapped ATOM";
  const tokenSymbol = "kATOM";
  const tokenDecimals = 6;

  const ERC20ZgChainWrappedCosmosCoin = await ethers.getContractFactory(
    "ERC20ZgChainWrappedCosmosCoin"
  );
  const token = await ERC20ZgChainWrappedCosmosCoin.deploy(
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
