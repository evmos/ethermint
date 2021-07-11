const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');
const yargs = require('yargs/yargs')
const { hideBin } = require('yargs/helpers')
const argv = yargs(hideBin(process.argv)).argv


const logger = {
  warn: msg => console.error(`WARN: ${msg}`),
  err: msg => console.error(`ERR: ${msg}`),
  info: msg => console.log(`INFO: ${msg}`)
}

function panic(errMsg) {
  logger.err(errMsg);
  process.exit(-1);
}

function checkTestEnv() {
  if (!fs.existsSync(path.join(__dirname, './node_modules'))) {
    panic('node_modules not existed. Please run `yarn install` before running tests.');
  }
  const runConfig = {};
  
  // Check test network
  if (!argv.network) {
    runConfig.network = 'ganache';
  }
  else {
    if (argv.network !== 'ethermint' && argv.network !== 'ganache') {
      panic('network is invalid. Must be ganache or ethermint');
    }
    else {
      runConfig.network = argv.network;
    }
  }

  // only test
  runConfig.onlyTest = argv['_'];
  return runConfig;

}

function loadTests() {
  const validTests = [];
  fs.readdirSync(path.join(__dirname, 'suites')).forEach(dirname => {
    const dirStat = fs.statSync(path.join(__dirname, 'suites', dirname));
    if (!dirStat.isDirectory) {
      logger.warn(`${dirname} is not a directory. Skip this test suite.`);
      return;
    }

    const needFiles = ['package.json', 'test'];
    for (const f of needFiles) {
      if (!fs.existsSync(path.join(__dirname, 'suites', dirname, f))) {
        logger.warn(`${dirname} does not contains file/dir: ${f}. Skip this test suite.`);
        return;
      }
    }

    // test package.json
    try {
      const testManifest = JSON.parse(fs.readFileSync(path.join(__dirname, 'suites', dirname, 'package.json'), 'utf-8'))
      const needScripts = ['test-ganache', 'test-ethermint'];
      for (const s of needScripts) {
        if (Object.keys(testManifest['scripts']).indexOf(s) === -1) {
          logger.warn(`${dirname} does not have test script: \`${s}\`. Skip this test suite.`);
          return;
        }
      }
    } catch (error) {
      logger.warn(`${dirname} test package.json load failed. Skip this test suite.`);
      logger.err(error);
      return;
    }
    validTests.push(dirname);
  })
  return validTests;
}

function performTestSuite(testName) {
  execSync(`yarn test-ganache`, {
    cwd: path.join(__dirname, 'suites', testName),
    stdio: 'inherit',
  });
}

function performTests({ allTests, runConfig }) {

  if (allTests.length === 0) {
    panic('No tests are found or all invalid!');
  }
  if (runConfig.onlyTest.length === 0) {
    logger.info('Start all tests:');
  }
  else {
    allTests = allTests.filter(t => runConfig.onlyTest.indexOf(t) !== -1);
    logger.info('Only run tests:');
  }
  console.log(allTests);

  for (const currentTestName of allTests) {
    logger.info(`Start test: ${currentTestName}`);
    performTestSuite(currentTestName);
  }

  logger.info(`${allTests.length} test suites passed!`);
}

async function main() {

  // console.log(argv);
  const runConfig = checkTestEnv();
  const allTests = loadTests(runConfig);

  performTests({allTests, runConfig});

}

main();