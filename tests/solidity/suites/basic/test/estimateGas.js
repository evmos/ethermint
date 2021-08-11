const Storage = artifacts.require("Storage")

contract('Storage', (accounts) => {

  let storage
  beforeEach(async () => {
    storage = await Storage.new()
  })

  it('estimated gas should match', async () => {
      // set new value
      let gasUsage = await storage.store.estimateGas(10);
      expect(gasUsage.toString()).to.equal('43754');

      await storage.store(10);

      // set existing value
      gasUsage = await storage.store.estimateGas(10);
      expect(gasUsage.toString()).to.equal('28754');
  })

})
