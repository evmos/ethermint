require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.10",
  typechain: {
    outDir: "typechain",
    target: "ethers-v5",
    runOnCompile: true
  },
};
